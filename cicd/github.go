package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func init() {
	Register("github", func() PipelineProvider { return &GitHubPipelineProvider{} })
}

// GitHubPipelineProvider 通过 GitHub Actions API 获取 CI/CD 数据。
type GitHubPipelineProvider struct {
	client *http.Client
}

func (p *GitHubPipelineProvider) getClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}
	return p.client
}

// --- GitHub API 响应结构 ---

type githubWorkflowRunsResponse struct {
	TotalCount   int                 `json:"total_count"`
	WorkflowRuns []githubWorkflowRun `json:"workflow_runs"`
}

type githubWorkflowRun struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	RunStartedAt string `json:"run_started_at"`
	HTMLURL    string `json:"html_url"`
}

// --- ListPipelines ---

func (p *GitHubPipelineProvider) ListPipelines(ctx context.Context, params ListPipelinesParams) ([]Pipeline, error) {
	if params.Limit <= 0 {
		params.Limit = 5
	}

	apiBase := githubAPIURL(params.ServerURL)
	apiURL := fmt.Sprintf("%s/repos/%s/actions/runs?per_page=%d",
		apiBase, params.RepoPath, params.Limit)

	var resp githubWorkflowRunsResponse
	if err := p.get(ctx, params.Token, apiURL, &resp); err != nil {
		return nil, fmt.Errorf("github: list workflow runs: %w", err)
	}

	result := make([]Pipeline, 0, len(resp.WorkflowRuns))
	for _, run := range resp.WorkflowRuns {
		result = append(result, mapGitHubWorkflowRun(run))
	}
	return result, nil
}

// --- GetPipeline ---

func (p *GitHubPipelineProvider) GetPipeline(ctx context.Context, params GetPipelineParams) (*Pipeline, error) {
	apiBase := githubAPIURL(params.ServerURL)
	apiURL := fmt.Sprintf("%s/repos/%s/actions/runs/%d",
		apiBase, params.RepoPath, params.PipelineID)

	var run githubWorkflowRun
	if err := p.get(ctx, params.Token, apiURL, &run); err != nil {
		return nil, fmt.Errorf("github: get workflow run: %w", err)
	}

	pipeline := mapGitHubWorkflowRun(run)
	return &pipeline, nil
}

// --- 映射 ---

func mapGitHubStatus(status, conclusion string) PipelineStatus {
	switch status {
	case "in_progress":
		return StatusRunning
	case "queued", "waiting", "requested", "pending":
		return StatusPending
	case "completed":
		switch conclusion {
		case "success":
			return StatusSuccess
		case "failure", "timed_out", "start_failure":
			return StatusFailed
		case "cancelled":
			return StatusCanceled
		default:
			// neutral, skipped, action_required 等
			return StatusCanceled
		}
	default:
		return StatusPending
	}
}

func mapGitHubWorkflowRun(run githubWorkflowRun) Pipeline {
	status := mapGitHubStatus(run.Status, run.Conclusion)

	var startedAt time.Time
	if run.RunStartedAt != "" {
		startedAt, _ = time.Parse(time.RFC3339, run.RunStartedAt)
	} else if run.CreatedAt != "" {
		startedAt, _ = time.Parse(time.RFC3339, run.CreatedAt)
	}

	// 计算持续时间
	var duration time.Duration
	if !startedAt.IsZero() && run.UpdatedAt != "" {
		if updatedAt, err := time.Parse(time.RFC3339, run.UpdatedAt); err == nil {
			if run.Status == "completed" {
				duration = updatedAt.Sub(startedAt)
				if duration < 0 {
					duration = 0
				}
			}
		}
	}

	sha := run.HeadSHA
	if len(sha) > 7 {
		sha = sha[:7]
	}

	return Pipeline{
		ID:        run.ID,
		Status:    status,
		Branch:    run.HeadBranch,
		SHA:       sha,
		StartedAt: startedAt,
		Duration:  duration,
		URL:       run.HTMLURL,
	}
}

// --- URL 转换 ---

// githubAPIURL 将用户可见的 server URL 转换为 GitHub REST API endpoint。
// 复用 provider/github.go 的逻辑。
func githubAPIURL(serverURL string) string {
	u := strings.TrimRight(serverURL, "/")
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if u == "github.com" {
		return "https://api.github.com"
	}
	return serverURL
}

// --- HTTP client ---

func (p *GitHubPipelineProvider) get(ctx context.Context, token, apiURL string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.getClient().Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed (invalid token)")
	}

	if resp.StatusCode == http.StatusForbidden {
		if remaining := resp.Header.Get("X-Ratelimit-Remaining"); remaining == "0" {
			return fmt.Errorf("rate limit exceeded")
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
