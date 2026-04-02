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
	Register("gitlab", func() Provider { return &GitLabProvider{} })
}

type GitLabProvider struct {
	client *http.Client
}

type gitlabProject struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	Name              string `json:"name"`
}

type gitlabGroup struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
}

func (g *GitLabProvider) getClient() *http.Client {
	if g.client == nil {
		g.client = &http.Client{Timeout: 30 * time.Second}
	}
	return g.client
}

func (g *GitLabProvider) ListRepos(ctx context.Context, params ListReposParams) ([]Repo, error) {
	var allRepos []Repo

	for _, group := range params.Groups {
		repos, err := g.listGroupRepos(ctx, params, group.Path, group.Recursive)
		if err != nil {
			return nil, fmt.Errorf("gitlab: group %s: %w", group.Path, err)
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

func (g *GitLabProvider) ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error) {
	var allGroups []RemoteGroup
	page := 1

	for {
		url := fmt.Sprintf("%s/api/v4/groups?per_page=100&page=%d", params.ServerURL, page)

		var groups []gitlabGroup
		nextPage, err := g.getWithPagination(ctx, params.Token, url, &groups)
		if err != nil {
			return nil, err
		}

		for _, grp := range groups {
			allGroups = append(allGroups, RemoteGroup{
				Name:     grp.Path,
				Path:     grp.FullPath,
				Provider: "gitlab",
			})
		}

		if nextPage == 0 {
			break
		}
		page = nextPage
	}

	return allGroups, nil
}

func (g *GitLabProvider) listGroupRepos(ctx context.Context, params ListReposParams, groupPath string, recursive bool) ([]Repo, error) {
	group, err := g.getGroupByPath(ctx, params, groupPath)
	if err != nil {
		return nil, err
	}

	var allRepos []Repo
	queue := []gitlabGroup{*group}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		projects, err := g.getGroupProjects(ctx, params, current.ID)
		if err != nil {
			return nil, err
		}

		for _, p := range projects {
			allRepos = append(allRepos, Repo{
				Name:     p.Name,
				CloneURL: p.HTTPURLToRepo,
				SSHURL:   p.SSHURLToRepo,
				Path:     p.PathWithNamespace,
				Provider: "gitlab",
			})
		}

		if recursive {
			subgroups, err := g.getSubgroups(ctx, params, current.ID)
			if err != nil {
				return nil, err
			}
			queue = append(queue, subgroups...)
		}
	}

	return allRepos, nil
}

func (g *GitLabProvider) getGroupByPath(ctx context.Context, params ListReposParams, path string) (*gitlabGroup, error) {
	encodedPath := strings.ReplaceAll(path, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/groups/%s", params.ServerURL, encodedPath)

	var group gitlabGroup
	if err := g.get(ctx, params.Token, url, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

func (g *GitLabProvider) getGroupProjects(ctx context.Context, params ListReposParams, groupID int) ([]gitlabProject, error) {
	var allProjects []gitlabProject
	page := 1

	for {
		url := fmt.Sprintf("%s/api/v4/groups/%d/projects?per_page=100&page=%d", params.ServerURL, groupID, page)

		var projects []gitlabProject
		nextPage, err := g.getWithPagination(ctx, params.Token, url, &projects)
		if err != nil {
			return nil, err
		}

		allProjects = append(allProjects, projects...)

		if nextPage == 0 {
			break
		}
		page = nextPage
	}

	return allProjects, nil
}

func (g *GitLabProvider) getSubgroups(ctx context.Context, params ListReposParams, groupID int) ([]gitlabGroup, error) {
	url := fmt.Sprintf("%s/api/v4/groups/%d/subgroups?per_page=100", params.ServerURL, groupID)

	var subgroups []gitlabGroup
	_, err := g.getWithPagination(ctx, params.Token, url, &subgroups)
	if err != nil {
		return nil, err
	}

	return subgroups, nil
}

func (g *GitLabProvider) get(ctx context.Context, token, url string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := g.getClient().Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkGitLabRateLimit(resp); err != nil {
		return err
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

func (g *GitLabProvider) getWithPagination(ctx context.Context, token, url string, v interface{}) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := g.getClient().Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkGitLabRateLimit(resp); err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return parseNextPage(resp.Header.Get("Link"))
}

func checkGitLabRateLimit(resp *http.Response) error {
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, _ := strconv.Atoi(retryAfter)
			return fmt.Errorf("rate limited, retry after %d seconds", seconds)
		}
		return fmt.Errorf("rate limited by API")
	}
	return nil
}

func parseNextPage(linkHeader string) (int, error) {
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
