package mergerequest

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/wii/grepom/codeupapi"
)

func init() {
	Register("codeup", func() MergeRequestProvider { return &CodeupMRProvider{} })
}

// CodeupMRProvider implements MergeRequestProvider for Alibaba Cloud Codeup (云效)
// using the OAPI v1 changeRequests endpoints.
type CodeupMRProvider struct {
	client *codeupapi.Client
}

func (p *CodeupMRProvider) getClient() *codeupapi.Client {
	if p.client == nil {
		p.client = codeupapi.NewClient()
	}
	return p.client
}

// --- Codeup OAPI v1 请求/响应结构 ---

// codeupRepoEntry 映射 ListRepositories 接口返回的仓库条目（仅取路径解析所需字段）。
type codeupRepoEntry struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"pathWithNamespace"`
}

// codeupCreateMRRequest 映射 CreateChangeRequest 请求体。
// 同库 MR 时 sourceProjectId 与 targetProjectId 均为仓库 ID。
type codeupCreateMRRequest struct {
	SourceBranch    string `json:"sourceBranch"`
	SourceProjectID int    `json:"sourceProjectId"`
	TargetBranch    string `json:"targetBranch"`
	TargetProjectID int    `json:"targetProjectId"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
}

// codeupMRResponse 映射合并请求响应（创建与列表共用）。
// state 取值：UNDER_DEV / UNDER_REVIEW / TO_BE_MERGED / CLOSED / MERGED。
type codeupMRResponse struct {
	LocalID        int    `json:"localId"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	State          string `json:"state"`  // 主状态字段
	Status         string `json:"status"` // 部分接口的冗余字段，作为兜底
	WebURL         string `json:"webUrl"`
	DetailURL      string `json:"detailUrl"`
	SourceBranch   string `json:"sourceBranch"`
	TargetBranch   string `json:"targetBranch"`
	WorkInProgress bool   `json:"workInProgress"`
	ProjectID      int    `json:"projectId"`
}

// CreateMergeRequest 通过 Codeup OAPI v1 创建合并请求。
// 流程：仓库路径 → 仓库 ID → 查重（已有打开中 MR 幂等返回）→ 创建。
func (p *CodeupMRProvider) CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (*MergeRequest, error) {
	if params.OrganizationID == "" {
		return nil, fmt.Errorf("codeup: organization_id is required (set organization_id in your codeup resource config)")
	}

	apiBase := codeupapi.APIBaseURL(params.ServerURL, "codeup", params.OrganizationID)

	// Step 1: 仓库路径 → 仓库 ID
	repoID, err := p.resolveRepoID(ctx, params.Token, apiBase, params.RepoPath)
	if err != nil {
		return nil, err
	}

	// Step 2: 幂等查重
	existing, err := p.findOpenMR(ctx, params.Token, apiBase, repoID, params.SourceBranch)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Step 3: 创建（Codeup 无 draft 字段，--draft 时在标题加 "Draft: " 前缀）
	title := params.Title
	if params.Draft {
		title = "Draft: " + title
	}

	apiURL := fmt.Sprintf("%s/repositories/%d/changeRequests", apiBase, repoID)
	reqBody := codeupCreateMRRequest{
		SourceBranch:    params.SourceBranch,
		SourceProjectID: repoID,
		TargetBranch:    params.TargetBranch,
		TargetProjectID: repoID,
		Title:           title,
		Description:     params.Description,
	}

	var mrResp codeupMRResponse
	if err := p.getClient().Post(ctx, params.Token, apiURL, reqBody, &mrResp); err != nil {
		return nil, err
	}

	return mapCodeupMR(mrResp, false), nil
}

// resolveRepoID 通过 ListRepositories 接口分页查找 pathWithNamespace 精确匹配的仓库 ID。
func (p *CodeupMRProvider) resolveRepoID(ctx context.Context, token, apiBase, repoPath string) (int, error) {
	page := 1
	for {
		apiURL := fmt.Sprintf("%s/repositories?page=%d&perPage=100", apiBase, page)

		var repos []codeupRepoEntry
		pg, err := p.getClient().GetWithPagination(ctx, token, apiURL, &repos)
		if err != nil {
			return 0, err
		}

		for _, r := range repos {
			if r.PathWithNamespace == repoPath {
				return r.ID, nil
			}
		}

		if pg.NextPage <= 0 || pg.NextPage <= page {
			break
		}
		page = pg.NextPage
	}

	return 0, fmt.Errorf("codeup: repository not found for path %q", repoPath)
}

// findOpenMR 查询仓库的打开中合并请求，按源分支客户端过滤；无匹配返回 (nil, nil)。
// ListChangeRequests 接口不支持按源分支过滤，state=opened 由服务端过滤。
func (p *CodeupMRProvider) findOpenMR(ctx context.Context, token, apiBase string, repoID int, sourceBranch string) (*MergeRequest, error) {
	apiURL := fmt.Sprintf("%s/changeRequests?projectIds=%d&state=opened&perPage=100",
		apiBase, repoID)

	var mrList []codeupMRResponse
	if err := p.getClient().Get(ctx, token, apiURL, &mrList); err != nil {
		return nil, fmt.Errorf("codeup: search existing merge requests: %w", err)
	}

	for _, mr := range mrList {
		if mr.SourceBranch == sourceBranch {
			return mapCodeupMR(mr, true), nil
		}
	}

	return nil, nil
}

// mapCodeupMR 将 Codeup 合并请求响应映射为 MergeRequest 结构。
func mapCodeupMR(mr codeupMRResponse, alreadyExists bool) *MergeRequest {
	webURL := mr.WebURL
	if webURL == "" {
		webURL = mr.DetailURL
	}

	state := mr.State
	if state == "" {
		state = mr.Status
	}

	return &MergeRequest{
		ID:            mr.LocalID,
		Number:        mr.LocalID,
		Title:         mr.Title,
		Description:   mr.Description,
		URL:           webURL,
		State:         mapCodeupMRState(state),
		SourceBranch:  mr.SourceBranch,
		TargetBranch:  mr.TargetBranch,
		Draft:         mr.WorkInProgress || strings.HasPrefix(mr.Title, "Draft: "),
		AlreadyExists: alreadyExists,
	}
}

// mapCodeupMRState 映射 Codeup MR 状态到通用状态。
// UNDER_DEV / UNDER_REVIEW / TO_BE_MERGED 均为打开中状态。
func mapCodeupMRState(state string) string {
	switch strings.ToUpper(state) {
	case "UNDER_DEV", "UNDER_REVIEW", "TO_BE_MERGED", "OPENED", "OPEN":
		return "open"
	case "MERGED":
		return "merged"
	case "CLOSED":
		return "closed"
	default:
		return strings.ToLower(state)
	}
}

// BuildWebURL 构建 Codeup MR 浏览器创建页面 URL。
func (p *CodeupMRProvider) BuildWebURL(params WebURLParams) string {
	webBase := strings.TrimRight(params.ServerURL, "/")
	return fmt.Sprintf("%s/%s/merge_requests/new?source=%s&target=%s",
		webBase, params.RepoPath,
		url.QueryEscape(params.SourceBranch),
		url.QueryEscape(params.TargetBranch),
	)
}
