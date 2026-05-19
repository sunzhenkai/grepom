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
	Register("github", func() MergeRequestProvider { return &GitHubMRProvider{} })
}

// GitHubMRProvider implements MergeRequestProvider for GitHub.
type GitHubMRProvider struct {
	client *http.Client
}

func (p *GitHubMRProvider) getClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}
	return p.client
}

// --- GitHub API 请求/响应结构 ---

type githubCreatePRRequest struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft,omitempty"`
}

type githubPRResponse struct {
	ID       int    `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	State    string `json:"state"`
	HTMLURL  string `json:"html_url"`
	Head     githubPRBranch `json:"head"`
	Base     githubPRBranch `json:"base"`
	Draft    bool   `json:"draft"`
}

type githubPRBranch struct {
	Ref string `json:"ref"`
}

type githubErrorResponse struct {
	Message string `json:"message"`
}

// --- githubAPIURL ensures the server URL points to GitHub's REST API endpoint. ---

func githubAPIURL(serverURL string) string {
	u := strings.TrimRight(serverURL, "/")
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if u == "github.com" {
		return "https://api.github.com"
	}
	return serverURL
}

// CreateMergeRequest creates a Pull Request via GitHub REST API.
func (p *GitHubMRProvider) CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (*MergeRequest, error) {
	apiBase := githubAPIURL(params.ServerURL)
	apiURL := fmt.Sprintf("%s/repos/%s/pulls", apiBase, params.RepoPath)

	reqBody := githubCreatePRRequest{
		Title: params.Title,
		Body:  params.Description,
		Head:  params.SourceBranch,
		Base:  params.TargetBranch,
		Draft: params.Draft,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+params.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("github: authentication failed (invalid token)")
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("github: forbidden (check token permissions)")
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422: A PR already exists for this branch (or validation error).
		// Try to find and return the existing PR.
		existing, findErr := p.findOpenPR(ctx, params)
		if findErr != nil {
			// Search failed; return original error
			var errResp githubErrorResponse
			if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
				return nil, fmt.Errorf("github API error %d: %s", resp.StatusCode, errResp.Message)
			}
			return nil, fmt.Errorf("github API error %d: %s", resp.StatusCode, string(respBody))
		}
		if existing != nil {
			return existing, nil
		}
		// No existing PR found; return original error
		var errResp githubErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
			return nil, fmt.Errorf("github API error %d: %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("github API error %d: %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp githubErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
			return nil, fmt.Errorf("github API error %d: %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("github API error %d: %s", resp.StatusCode, string(respBody))
	}

	var prResp githubPRResponse
	if err := json.Unmarshal(respBody, &prResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &MergeRequest{
		ID:           prResp.ID,
		Number:       prResp.Number,
		Title:        prResp.Title,
		Description:  prResp.Body,
		URL:          prResp.HTMLURL,
		State:        prResp.State,
		SourceBranch: prResp.Head.Ref,
		TargetBranch: prResp.Base.Ref,
		Draft:        prResp.Draft,
	}, nil
}

// findOpenPR searches for an existing open PR for the given source branch.
// It calls GitHub's list PRs API with head and state=open filters.
func (p *GitHubMRProvider) findOpenPR(ctx context.Context, params CreateMergeRequestParams) (*MergeRequest, error) {
	apiBase := githubAPIURL(params.ServerURL)

	// GitHub's head parameter requires "owner:branch" format
	// Extract owner from RepoPath (which is "owner/repo")
	owner := params.RepoPath
	if idx := strings.Index(owner, "/"); idx >= 0 {
		owner = owner[:idx]
	}
	head := fmt.Sprintf("%s:%s", owner, params.SourceBranch)

	apiURL := fmt.Sprintf("%s/repos/%s/pulls?head=%s&state=open", apiBase, params.RepoPath, url.QueryEscape(head))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+params.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github search API error %d: %s", resp.StatusCode, string(respBody))
	}

	var prList []githubPRResponse
	if err := json.Unmarshal(respBody, &prList); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	if len(prList) == 0 {
		return nil, nil
	}

	// Take the first (typically only) open PR
	pr := prList[0]
	return &MergeRequest{
		ID:           pr.ID,
		Number:       pr.Number,
		Title:        pr.Title,
		Description:  pr.Body,
		URL:          pr.HTMLURL,
		State:        pr.State,
		SourceBranch: pr.Head.Ref,
		TargetBranch: pr.Base.Ref,
		Draft:        pr.Draft,
		AlreadyExists: true,
	}, nil
}

// BuildWebURL constructs a GitHub PR creation URL for the browser.
func (p *GitHubMRProvider) BuildWebURL(params WebURLParams) string {
	baseURL := buildGitHubWebBaseURL(params.ServerURL)
	u := fmt.Sprintf("%s/%s/compare/%s...%s?expand=1",
		baseURL, params.RepoPath,
		url.QueryEscape(params.TargetBranch),
		url.QueryEscape(params.SourceBranch),
	)
	if params.Draft {
		u += "&draft=1"
	}
	return u
}

// buildGitHubWebBaseURL converts an API URL back to the web URL for GitHub.
func buildGitHubWebBaseURL(serverURL string) string {
	u := strings.TrimRight(serverURL, "/")
	// api.github.com → github.com
	if u == "https://api.github.com" {
		return "https://github.com"
	}
	// GHE: strip /api/v3 suffix if present
	u = strings.TrimSuffix(u, "/api/v3")
	// If it lost the scheme, add it back
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "https://" + u
	}
	return u
}
