package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/wii/grepom/codeupapi"
)

func init() {
	Register("codeup", func() PipelineProvider { return &CodeupPipelineProvider{} })
}

// CodeupPipelineProvider 通过云效 Flow OAPI v1 获取 Codeup 仓库关联的流水线运行记录。
// Flow 流水线是组织级资源，需要按流水线代码源地址与目标仓库匹配。
type CodeupPipelineProvider struct {
	client *codeupapi.Client
}

func (p *CodeupPipelineProvider) getClient() *codeupapi.Client {
	if p.client == nil {
		p.client = codeupapi.NewClient()
	}
	return p.client
}

// --- 云效 Flow OAPI v1 响应结构 ---
// 注：Flow 接口字段在不同版本文档中命名有差异，以下结构对常见命名做了兼容
// （id/pipelineId、status、startTime/createTime 等），映射时取第一个非零值。

// flowPipeline 映射 ListPipelines 接口返回的流水线条目。
type flowPipeline struct {
	ID         int64  `json:"id"`
	PipelineID int64  `json:"pipelineId"`
	Name       string `json:"name"`
	Sources    []struct {
		Type string `json:"type"`
		Repo string `json:"repo"` // 代码源仓库地址（http/ssh）
	} `json:"sources"`
}

// flowID 返回流水线的有效 ID（兼容 id / pipelineId 两种命名）。
func (fp flowPipeline) flowID() int64 {
	if fp.ID != 0 {
		return fp.ID
	}
	return fp.PipelineID
}

// flowRun 映射流水线运行记录。
type flowRun struct {
	ID            int64           `json:"id"`
	PipelineRunID int64           `json:"pipelineRunId"`
	Status        string          `json:"status"`     // RUNNING / QUEUING / SUCCESS / FAIL / CANCELED ...
	StartTime     json.RawMessage `json:"startTime"`  // epoch ms 或 RFC3339
	EndTime       json.RawMessage `json:"endTime"`
	CreateTime    json.RawMessage `json:"createTime"`
	FinishTime    json.RawMessage `json:"finishTime"`
	Sources       []struct {
		Branch string `json:"branch"`
		SHA    string `json:"sha"`
		Commit string `json:"commit"`
	} `json:"sources"`
	SourceBranch string `json:"sourceBranch"`
	CommitID     string `json:"commitId"`
}

// runID 返回运行记录的有效 ID。
func (r flowRun) runID() int64 {
	if r.ID != 0 {
		return r.ID
	}
	return r.PipelineRunID
}

// branch 返回运行记录的分支（sources 优先，其次 sourceBranch）。
func (r flowRun) branch() string {
	if len(r.Sources) > 0 && r.Sources[0].Branch != "" {
		return r.Sources[0].Branch
	}
	return r.SourceBranch
}

// sha 返回运行记录的短 commit hash。
func (r flowRun) sha() string {
	sha := r.CommitID
	if len(r.Sources) > 0 {
		if r.Sources[0].SHA != "" {
			sha = r.Sources[0].SHA
		} else if r.Sources[0].Commit != "" {
			sha = r.Sources[0].Commit
		}
	}
	if len(sha) > 7 {
		sha = sha[:7]
	}
	return sha
}

// --- ListPipelines ---

func (p *CodeupPipelineProvider) ListPipelines(ctx context.Context, params ListPipelinesParams) ([]Pipeline, error) {
	if params.OrganizationID == "" {
		return nil, fmt.Errorf("codeup: organization_id is required (set organization_id in your codeup resource config)")
	}

	if params.Limit <= 0 {
		params.Limit = 5
	}

	flowBase := codeupapi.APIBaseURL(params.ServerURL, "flow", params.OrganizationID)

	// Step 1: 列出组织流水线，按代码源地址匹配目标仓库
	matched, err := p.matchPipelines(ctx, params.Token, flowBase, params.RepoPath)
	if err != nil {
		return nil, err
	}
	if len(matched) == 0 {
		return nil, nil
	}

	// Step 2: 查询各匹配流水线的运行记录，合并排序
	var all []Pipeline
	for _, pl := range matched {
		runs, err := p.listRuns(ctx, params.Token, flowBase, pl.flowID())
		if err != nil {
			return nil, err
		}
		for _, run := range runs {
			all = append(all, mapFlowRun(run, pl.flowID(), params.ServerURL))
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].StartedAt.After(all[j].StartedAt)
	})

	if len(all) > params.Limit {
		all = all[:params.Limit]
	}

	return all, nil
}

// --- GetPipeline ---

func (p *CodeupPipelineProvider) GetPipeline(ctx context.Context, params GetPipelineParams) (*Pipeline, error) {
	if params.OrganizationID == "" {
		return nil, fmt.Errorf("codeup: organization_id is required (set organization_id in your codeup resource config)")
	}

	flowBase := codeupapi.APIBaseURL(params.ServerURL, "flow", params.OrganizationID)

	// Flow 查询单次运行需要 pipelineId + runId，先在匹配流水线中定位该运行
	matched, err := p.matchPipelines(ctx, params.Token, flowBase, params.RepoPath)
	if err != nil {
		return nil, err
	}

	for _, pl := range matched {
		runs, err := p.listRuns(ctx, params.Token, flowBase, pl.flowID())
		if err != nil {
			return nil, err
		}
		for _, run := range runs {
			if run.runID() == int64(params.PipelineID) {
				result := mapFlowRun(run, pl.flowID(), params.ServerURL)
				return &result, nil
			}
		}
	}

	return nil, fmt.Errorf("codeup: pipeline run #%d not found for repo %q", params.PipelineID, params.RepoPath)
}

// --- 流水线匹配与运行记录查询 ---

// matchPipelines 分页列出组织流水线，返回代码源地址指向 repoPath 的流水线。
func (p *CodeupPipelineProvider) matchPipelines(ctx context.Context, token, flowBase, repoPath string) ([]flowPipeline, error) {
	var matched []flowPipeline
	page := 1

	for {
		apiURL := fmt.Sprintf("%s/pipelines?page=%d&perPage=100", flowBase, page)

		var raw json.RawMessage
		pg, err := p.getClient().GetWithPagination(ctx, token, apiURL, &raw)
		if err != nil {
			return nil, fmt.Errorf("codeup: list flow pipelines: %w", err)
		}

		pipelines, err := decodeFlowList[flowPipeline](raw, "pipelines", "items", "list", "data")
		if err != nil {
			return nil, fmt.Errorf("codeup: decode flow pipelines: %w", err)
		}

		for _, pl := range pipelines {
			for _, src := range pl.Sources {
				if normalizeGitPath(src.Repo) == repoPath {
					matched = append(matched, pl)
					break
				}
			}
		}

		if pg.NextPage <= 0 || pg.NextPage <= page {
			break
		}
		page = pg.NextPage
	}

	return matched, nil
}

// listRuns 查询指定流水线的运行记录（单页，取最近记录）。
func (p *CodeupPipelineProvider) listRuns(ctx context.Context, token, flowBase string, pipelineID int64) ([]flowRun, error) {
	apiURL := fmt.Sprintf("%s/pipelines/%d/runs?page=1&perPage=20", flowBase, pipelineID)

	var raw json.RawMessage
	if err := p.getClient().Get(ctx, token, apiURL, &raw); err != nil {
		return nil, fmt.Errorf("codeup: list runs for pipeline %d: %w", pipelineID, err)
	}

	runs, err := decodeFlowList[flowRun](raw, "runs", "items", "list", "data")
	if err != nil {
		return nil, fmt.Errorf("codeup: decode runs for pipeline %d: %w", pipelineID, err)
	}

	return runs, nil
}

// decodeFlowList 将响应解码为数组，容忍两种响应形态：
// 顶层 JSON 数组，或包装对象（按 keys 候选字段取数组）。
func decodeFlowList[T any](raw json.RawMessage, keys ...string) ([]T, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	// 形态 1：顶层数组
	var direct []T
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, nil
	}

	// 形态 2：包装对象
	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("unexpected response shape: %w", err)
	}
	for _, key := range keys {
		if v, ok := wrapped[key]; ok {
			var items []T
			if err := json.Unmarshal(v, &items); err == nil {
				return items, nil
			}
		}
	}

	return nil, nil
}

// normalizeGitPath 从 git URL（http/ssh/scp）中提取仓库路径（去 scheme、host、.git 后缀）。
func normalizeGitPath(rawURL string) string {
	u := strings.TrimSpace(rawURL)
	u = strings.TrimSuffix(u, ".git")

	if strings.HasPrefix(u, "git@") {
		// scp 风格：git@host:path
		u = strings.TrimPrefix(u, "git@")
		if idx := strings.Index(u, ":"); idx >= 0 {
			return u[idx+1:]
		}
		return u
	}

	if parsed, err := url.Parse(u); err == nil && parsed.Path != "" {
		return strings.TrimPrefix(parsed.Path, "/")
	}

	return u
}

// --- 映射 ---

// mapFlowStatus 映射 Flow 运行状态到通用 PipelineStatus。
func mapFlowStatus(status string) PipelineStatus {
	switch strings.ToUpper(status) {
	case "RUNNING":
		return StatusRunning
	case "QUEUING", "WAITING", "PENDING", "PAUSED":
		return StatusPending
	case "SUCCESS":
		return StatusSuccess
	case "FAIL", "FAILED", "ERROR":
		return StatusFailed
	case "CANCELED", "CANCELLED":
		return StatusCanceled
	default:
		return PipelineStatus(strings.ToLower(status))
	}
}

// mapFlowRun 将 Flow 运行记录映射为 Pipeline。
func mapFlowRun(run flowRun, pipelineID int64, serverURL string) Pipeline {
	startedAt := parseFlowTime(run.StartTime)
	if startedAt.IsZero() {
		startedAt = parseFlowTime(run.CreateTime)
	}

	endAt := parseFlowTime(run.EndTime)
	if endAt.IsZero() {
		endAt = parseFlowTime(run.FinishTime)
	}

	var duration time.Duration
	if !startedAt.IsZero() && !endAt.IsZero() && endAt.After(startedAt) {
		duration = endAt.Sub(startedAt)
	}

	return Pipeline{
		ID:        int(run.runID()),
		Status:    mapFlowStatus(run.Status),
		Branch:    run.branch(),
		SHA:       run.sha(),
		StartedAt: startedAt,
		Duration:  duration,
		URL:       flowRunWebURL(serverURL, pipelineID),
	}
}

// flowRunWebURL 构造 Flow 流水线的 Web URL。
// 官方环境使用 flow.aliyun.com；自定义/测试环境基于 serverURL 构造。
func flowRunWebURL(serverURL string, pipelineID int64) string {
	if codeupapi.CloneHost(serverURL) == "codeup.aliyun.com" {
		return fmt.Sprintf("https://flow.aliyun.com/pipelines/%d", pipelineID)
	}
	base := strings.TrimRight(serverURL, "/")
	return fmt.Sprintf("%s/flow/pipelines/%d", base, pipelineID)
}

// parseFlowTime 解析 Flow 接口的时间字段，兼容 epoch 毫秒与 RFC3339 字符串。
func parseFlowTime(raw json.RawMessage) time.Time {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}
	}

	// epoch 毫秒
	var ms int64
	if err := json.Unmarshal(raw, &ms); err == nil && ms > 0 {
		return time.UnixMilli(ms)
	}

	// RFC3339 字符串
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}

	return time.Time{}
}
