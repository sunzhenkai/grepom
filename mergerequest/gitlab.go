package mergerequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	Register("gitlab", func() MergeRequestProvider { return &GitLabMRProvider{} })
}

// GitLabMRProvider implements MergeRequestProvider for GitLab.
type GitLabMRProvider struct {
	client *http.Client
}

func (p *GitLabMRProvider) getClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}
	return p.client
}

// --- GitLab API 请求/响应结构 ---

type gitlabCreateMRRequest struct {
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
}

type gitlabMRResponse struct {
	ID          int    `json:"id"`
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	WebURL      string `json:"web_url"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
}

// CreateMergeRequest creates a Merge Request via GitLab REST API.
func (p *GitLabMRProvider) CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (*MergeRequest, error) {
	encodedPath := url.PathEscape(params.RepoPath)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests", params.ServerURL, encodedPath)

	title := params.Title
	if params.Draft {
		title = "Draft: " + title
	}

	reqBody := gitlabCreateMRRequest{
		SourceBranch: params.SourceBranch,
		TargetBranch: params.TargetBranch,
		Title:        title,
		Description:  params.Description,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", params.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("gitlab: authentication failed (invalid token)")
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("gitlab API error %d: %s", resp.StatusCode, string(respBody))
	}

	var mrResp gitlabMRResponse
	if err := json.Unmarshal(respBody, &mrResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	isDraft := strings.HasPrefix(mrResp.Title, "Draft: ")

	return &MergeRequest{
		ID:           mrResp.ID,
		Number:       mrResp.IID,
		Title:        mrResp.Title,
		Description:  mrResp.Description,
		URL:          mrResp.WebURL,
		State:        mrResp.State,
		SourceBranch: mrResp.SourceBranch,
		TargetBranch: mrResp.TargetBranch,
		Draft:        isDraft,
	}, nil
}

// BuildWebURL constructs a GitLab MR creation URL for the browser.
func (p *GitLabMRProvider) BuildWebURL(params WebURLParams) string {
	// Trim trailing slash and scheme for web URL
	webBase := strings.TrimRight(params.ServerURL, "/")
	webBase = strings.TrimPrefix(webBase, "https://")
	webBase = strings.TrimPrefix(webBase, "http://")
	if !strings.HasPrefix(params.ServerURL, "http://") {
		webBase = "https://" + webBase
	} else {
		webBase = "http://" + webBase
	}

	u := fmt.Sprintf("%s/%s/-/merge_requests/new?merge_request[source_branch]=%s&merge_request[target_branch]=%s",
		webBase, params.RepoPath,
		url.QueryEscape(params.SourceBranch),
		url.QueryEscape(params.TargetBranch),
	)
	if params.Title != "" {
		u += "&merge_request[title]=" + url.QueryEscape(params.Title)
	}
	if params.Draft {
		u += "&merge_request[draft]=true"
	}
	return u
}
