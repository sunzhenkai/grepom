package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register("github", func() Provider { return &GitHubProvider{} })
}

type GitHubProvider struct {
	client *http.Client
}

type githubRepo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
	HTMLURL  string `json:"html_url"`
	Private  bool   `json:"private"`
}

func (g *GitHubProvider) getClient() *http.Client {
	if g.client == nil {
		g.client = &http.Client{Timeout: 30 * time.Second}
	}
	return g.client
}

func (g *GitHubProvider) ListRepos(ctx context.Context, params ListReposParams) ([]Repo, error) {
	var allRepos []Repo

	for _, org := range params.Orgs {
		repos, err := g.listOrgRepos(ctx, params, org)
		if err != nil {
			return nil, fmt.Errorf("github: org %s: %w", org, err)
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

func (g *GitHubProvider) listOrgRepos(ctx context.Context, params ListReposParams, orgName string) ([]Repo, error) {
	var allRepos []Repo
	page := 1

	for {
		apiURL := fmt.Sprintf("%s/orgs/%s/repos?per_page=100&page=%d&type=all", params.ServerURL, orgName, page)

		var repos []githubRepo
		nextPage, err := g.getWithPagination(ctx, params.Token, apiURL, &repos)
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			allRepos = append(allRepos, Repo{
				Name:     r.Name,
				CloneURL: r.CloneURL,
				SSHURL:   r.SSHURL,
				Path:     r.FullName,
				Provider: "github",
			})
		}

		if nextPage == 0 {
			break
		}
		page = nextPage
	}

	return allRepos, nil
}

func (g *GitHubProvider) getWithPagination(ctx context.Context, token, apiURL string, v interface{}) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.getClient().Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkGitHubRateLimit(resp); err != nil {
		return 0, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return 0, fmt.Errorf("github: authentication failed (invalid token)")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("github API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return parseGitHubNextPage(resp.Header.Get("Link"))
}

func checkGitHubRateLimit(resp *http.Response) error {
	if resp.StatusCode == http.StatusForbidden {
		if remaining := resp.Header.Get("X-Ratelimit-Remaining"); remaining == "0" {
			resetStr := resp.Header.Get("X-Ratelimit-Reset")
			if reset, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
				resetTime := time.Unix(reset, 0)
				return fmt.Errorf("github: rate limit exceeded, resets at %s", resetTime.Format(time.RFC3339))
			}
			return fmt.Errorf("github: rate limit exceeded")
		}
	}
	return nil
}

func parseGitHubNextPage(linkHeader string) (int, error) {
	if linkHeader == "" {
		return 0, nil
	}

	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if strings.HasSuffix(part, `rel="next"`) {
			urlPart := strings.TrimSpace(strings.Split(part, ";")[0])
			urlPart = strings.Trim(urlPart, "<>")
			for _, param := range strings.Split(urlPart, "&") {
				if strings.HasPrefix(param, "page=") {
					return strconv.Atoi(strings.TrimPrefix(param, "page="))
				}
			}
		}
	}

	return 0, nil
}
