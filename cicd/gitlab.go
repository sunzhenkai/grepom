package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	Register("gitlab", func() PipelineProvider { return &GitLabPipelineProvider{} })
}

// GitLabPipelineProvider 通过 GitLab Pipelines API 获取 CI/CD 数据。
type GitLabPipelineProvider struct {
	client *http.Client
}

func (p *GitLabPipelineProvider) getClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}
	return p.client
}

// --- GitLab API 响应结构 ---

type gitlabPipeline struct {
	ID        int    `json:"id"`
	SHA       string `json:"sha"`
	Ref       string `json:"ref"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	StartedAt string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	Duration  float64 `json:"duration"`
	WebURL    string `json:"web_url"`
}

// --- ListPipelines ---

func (p *GitLabPipelineProvider) ListPipelines(ctx context.Context, params ListPipelinesParams) ([]Pipeline, error) {
	if params.Limit <= 0 {
		params.Limit = 5
	}

	encodedPath := url.PathEscape(params.RepoPath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipelines?per_page=%d&order_by=id&sort=desc",
		params.ServerURL, encodedPath, params.Limit)

	var pipelines []gitlabPipeline
	if err := p.get(ctx, params.Token, apiURL, &pipelines); err != nil {
		return nil, fmt.Errorf("gitlab: list pipelines: %w", err)
	}

	result := make([]Pipeline, 0, len(pipelines))
	for _, pl := range pipelines {
		result = append(result, mapGitLabPipeline(pl))
	}
	return result, nil
}

// --- GetPipeline ---

func (p *GitLabPipelineProvider) GetPipeline(ctx context.Context, params GetPipelineParams) (*Pipeline, error) {
	encodedPath := url.PathEscape(params.RepoPath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipelines/%d",
		params.ServerURL, encodedPath, params.PipelineID)

	var pl gitlabPipeline
	if err := p.get(ctx, params.Token, apiURL, &pl); err != nil {
		return nil, fmt.Errorf("gitlab: get pipeline: %w", err)
	}

	pipeline := mapGitLabPipeline(pl)
	return &pipeline, nil
}

// --- 映射 ---

func mapGitLabStatus(status string) PipelineStatus {
	switch status {
	case "running":
		return StatusRunning
	case "pending":
		return StatusPending
	case "success":
		return StatusSuccess
	case "failed", "canceled_for_retry":
		return StatusFailed
	case "canceled":
		return StatusCanceled
	default:
		return PipelineStatus(status)
	}
}

func mapGitLabPipeline(pl gitlabPipeline) Pipeline {
	status := mapGitLabStatus(pl.Status)

	var startedAt time.Time
	if pl.StartedAt != "" {
		startedAt, _ = time.Parse(time.RFC3339, pl.StartedAt)
	}

	var duration time.Duration
	if pl.Duration > 0 {
		duration = time.Duration(pl.Duration * float64(time.Second))
	}

	sha := pl.SHA
	if len(sha) > 7 {
		sha = sha[:7]
	}

	return Pipeline{
		ID:        pl.ID,
		Status:    status,
		Branch:    pl.Ref,
		SHA:       sha,
		StartedAt: startedAt,
		Duration:  duration,
		URL:       pl.WebURL,
	}
}

// --- HTTP client ---

func (p *GitLabPipelineProvider) get(ctx context.Context, token, apiURL string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := p.getClient().Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
