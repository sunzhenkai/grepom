package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register("codeup", func() Provider { return &CodeupProvider{} })
}

// CodeupProvider implements the Provider interface for Alibaba Cloud Codeup (云效)
// using the new OAPI v1 endpoints.
type CodeupProvider struct {
	client *http.Client
}

// --- OAPI v1 response structures ---

// codeupNamespace maps a namespace entry from the ListNamespaces API.
type codeupNamespace struct {
	ID                int    `json:"id"`
	Path              string `json:"path"`
	FullPath          string `json:"fullPath"`
	PathWithNamespace string `json:"pathWithNamespace"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"nameWithNamespace"`
	ParentID          int    `json:"parentId"`
	Kind              string `json:"kind"`
	Visibility        string `json:"visibility"`
	AvatarURL         string `json:"avatarUrl"`
	WebURL            string `json:"webUrl"`
}

// codeupRepo maps a repository entry from the ListRepositories / ListGroupRepositories API.
type codeupRepo struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"pathWithNamespace"`
	NameWithNamespace string `json:"nameWithNamespace"`
	Description       string `json:"description"`
	WebURL            string `json:"webUrl"`
	Visibility        string `json:"visibility"`
	Archived          bool   `json:"archived"`
	NamespaceID       int    `json:"namespaceId"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	LastActivityAt    string `json:"lastActivityAt"`
	AccessLevel       int    `json:"accessLevel"`
}

// --- Helpers ---

func (p *CodeupProvider) getClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}
	return p.client
}

// codeupAPIBaseURL maps the user-facing clone URL to the OAPI v1 API base URL.
// For codeup.aliyun.com → https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/{orgId}
// For custom/test URLs, the original scheme and host are preserved.
// The original serverURL is used only for clone URL construction.
func codeupAPIBaseURL(serverURL, orgID string) string {
	h := strings.TrimRight(serverURL, "/")
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")

	var scheme string
	if strings.HasPrefix(serverURL, "http://") {
		scheme = "http"
	} else {
		scheme = "https"
	}

	apiHost := h
	if h == "codeup.aliyun.com" {
		apiHost = "openapi-rdc.aliyuncs.com"
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/oapi/v1/codeup/organizations/%s", scheme, apiHost, url.PathEscape(orgID))
}

// codeupCloneHost extracts the clone host from the user-facing server URL.
func codeupCloneHost(serverURL string) string {
	h := strings.TrimRight(serverURL, "/")
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	return h
}

// codeupPagination holds pagination info extracted from response headers.
type codeupPagination struct {
	Total      int
	TotalPages int
	Page       int
	PerPage    int
	NextPage   int // 0 means no next page
}

// parsePagination extracts pagination info from response headers.
func parsePagination(resp *http.Response) codeupPagination {
	var pg codeupPagination
	pg.Total, _ = strconv.Atoi(resp.Header.Get("x-total"))
	pg.TotalPages, _ = strconv.Atoi(resp.Header.Get("x-total-pages"))
	pg.Page, _ = strconv.Atoi(resp.Header.Get("x-page"))
	pg.PerPage, _ = strconv.Atoi(resp.Header.Get("x-per-page"))
	pg.NextPage, _ = strconv.Atoi(resp.Header.Get("x-next-page"))
	return pg
}

// get makes a GET request to the OAPI v1 with x-yunxiao-token header authentication.
// It decodes the JSON response body directly into v (no wrapper).
func (p *CodeupProvider) get(ctx context.Context, token, apiURL string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if token != "" {
		req.Header.Set("x-yunxiao-token", token)
	}

	resp, err := p.getClient().Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("codeup: authentication failed (invalid token)")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("codeup API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// getWithPagination makes a GET request and returns both parsed result and pagination info.
func (p *CodeupProvider) getWithPagination(ctx context.Context, token, apiURL string, v interface{}) (codeupPagination, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return codeupPagination{}, fmt.Errorf("create request: %w", err)
	}

	if token != "" {
		req.Header.Set("x-yunxiao-token", token)
	}

	resp, err := p.getClient().Do(req)
	if err != nil {
		return codeupPagination{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return codeupPagination{}, fmt.Errorf("codeup: authentication failed (invalid token)")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return codeupPagination{}, fmt.Errorf("codeup API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return codeupPagination{}, fmt.Errorf("decode response: %w", err)
	}

	return parsePagination(resp), nil
}

// --- ListRepos implementation ---

func (p *CodeupProvider) ListRepos(ctx context.Context, params ListReposParams) ([]Repo, error) {
	apiBase := codeupAPIBaseURL(params.ServerURL, params.OrganizationID)
	cloneHost := codeupCloneHost(params.ServerURL)

	var allRepos []Repo

	for _, group := range params.Groups {
		repos, err := p.listGroupRepos(ctx, params.Token, apiBase, cloneHost, group)
		if err != nil {
			return nil, fmt.Errorf("codeup: group %s: %w", group.Path, err)
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

// listGroupRepos fetches repos for a specific group path using the two-step strategy:
// 1. resolve group path to groupId via ListNamespaces
// 2. fetch repos via ListGroupRepositories
// Falls back to full list + client-side filtering if group not found.
func (p *CodeupProvider) listGroupRepos(ctx context.Context, token, apiBase, cloneHost string, group GroupQuery) ([]Repo, error) {
	// Empty group path: full list, no filtering
	if group.Path == "" {
		return p.listAllReposFull(ctx, token, apiBase, cloneHost, "")
	}

	// Step 1: try to resolve group path to groupId
	groupID, err := p.resolveGroupID(ctx, token, apiBase, group.Path)
	if err != nil {
		return nil, err
	}

	if groupID > 0 {
		// Step 2: fetch repos by groupId
		return p.listGroupReposByID(ctx, token, apiBase, cloneHost, groupID, group.Recursive)
	}

	// Fallback: full list + client-side prefix filtering
	return p.listAllReposFull(ctx, token, apiBase, cloneHost, group.Path)
}

// resolveGroupID searches namespaces for an exact pathWithNamespace match.
// Returns 0 (without error) if no match found (triggers fallback).
func (p *CodeupProvider) resolveGroupID(ctx context.Context, token, apiBase, groupPath string) (int, error) {
	apiURL := fmt.Sprintf("%s/namespaces?search=%s&perPage=100",
		apiBase, url.QueryEscape(groupPath))

	var namespaces []codeupNamespace
	_, err := p.getWithPagination(ctx, token, apiURL, &namespaces)
	if err != nil {
		return 0, err
	}

	for _, ns := range namespaces {
		if ns.PathWithNamespace == groupPath {
			return ns.ID, nil
		}
	}

	// No exact match — return 0 to trigger fallback
	return 0, nil
}

// listGroupReposByID fetches repos for a specific group using ListGroupRepositories API.
func (p *CodeupProvider) listGroupReposByID(ctx context.Context, token, apiBase, cloneHost string, groupID int, recursive bool) ([]Repo, error) {
	var allRepos []Repo
	page := 1

	for {
		apiURL := fmt.Sprintf("%s/groups/%d/repositories?page=%d&perPage=100&includeSubgroups=%v",
			apiBase, groupID, page, recursive)

		var repos []codeupRepo
		pg, err := p.getWithPagination(ctx, token, apiURL, &repos)
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			allRepos = append(allRepos, Repo{
				Name:     r.Name,
				CloneURL: "https://" + cloneHost + "/" + r.PathWithNamespace + ".git",
				SSHURL:   "git@" + cloneHost + ":" + r.PathWithNamespace + ".git",
				Path:     r.PathWithNamespace,
				Provider: "codeup",
			})
		}

		if pg.NextPage <= 0 || pg.NextPage <= page {
			break
		}
		page = pg.NextPage
	}

	return allRepos, nil
}

// listAllReposFull fetches all repos via ListRepositories and optionally filters by path prefix.
func (p *CodeupProvider) listAllReposFull(ctx context.Context, token, apiBase, cloneHost, pathPrefix string) ([]Repo, error) {
	var allRepos []Repo
	page := 1

	for {
		apiURL := fmt.Sprintf("%s/repositories?page=%d&perPage=100", apiBase, page)

		var repos []codeupRepo
		pg, err := p.getWithPagination(ctx, token, apiURL, &repos)
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			// Filter by path prefix if specified
			if pathPrefix != "" {
				if !strings.HasPrefix(r.PathWithNamespace, pathPrefix+"/") {
					continue
				}
			}

			allRepos = append(allRepos, Repo{
				Name:     r.Name,
				CloneURL: "https://" + cloneHost + "/" + r.PathWithNamespace + ".git",
				SSHURL:   "git@" + cloneHost + ":" + r.PathWithNamespace + ".git",
				Path:     r.PathWithNamespace,
				Provider: "codeup",
			})
		}

		if pg.NextPage <= 0 || pg.NextPage <= page {
			break
		}
		page = pg.NextPage
	}

	return allRepos, nil
}

// --- ListGroups implementation ---

func (p *CodeupProvider) ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error) {
	if params.OrganizationID == "" {
		return nil, nil
	}

	apiBase := codeupAPIBaseURL(params.ServerURL, params.OrganizationID)

	var allGroups []RemoteGroup
	page := 1

	for {
		apiURL := fmt.Sprintf("%s/namespaces?page=%d&perPage=100", apiBase, page)

		var namespaces []codeupNamespace
		pg, err := p.getWithPagination(ctx, params.Token, apiURL, &namespaces)
		if err != nil {
			// Graceful degradation: return empty list on failure
			return nil, nil
		}

		for _, ns := range namespaces {
			path := ns.PathWithNamespace
			if path == "" {
				path = ns.FullPath
			}
			allGroups = append(allGroups, RemoteGroup{
				Name:     ns.Path,
				Path:     path,
				Provider: "codeup",
			})
		}

		if pg.NextPage <= 0 || pg.NextPage <= page {
			break
		}
		page = pg.NextPage
	}

	return allGroups, nil
}
