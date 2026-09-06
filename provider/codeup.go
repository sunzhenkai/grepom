package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/wii/grepom/codeupapi"
	"github.com/wii/grepom/config"
)

func init() {
	Register("codeup", func() Provider { return &CodeupProvider{} })
}

// CodeupProvider implements the Provider interface for Alibaba Cloud Codeup (云效)
// using the new OAPI v1 endpoints.
type CodeupProvider struct {
	client *codeupapi.Client
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

func (p *CodeupProvider) getClient() *codeupapi.Client {
	if p.client == nil {
		p.client = codeupapi.NewClient()
	}
	return p.client
}

// codeupAPIBaseURL 保留给包内测试与既有调用点使用，实际逻辑在 codeupapi.APIBaseURL。
func codeupAPIBaseURL(serverURL, orgID string) string {
	return codeupapi.APIBaseURL(serverURL, "codeup", orgID)
}

// codeupCloneHost 保留给包内测试与既有调用点使用，实际逻辑在 codeupapi.CloneHost。
func codeupCloneHost(serverURL string) string {
	return codeupapi.CloneHost(serverURL)
}

// --- ListRepos implementation ---

func (p *CodeupProvider) ListRepos(ctx context.Context, params ListReposParams) ([]Repo, error) {
	apiBase := codeupAPIBaseURL(params.ServerURL, params.OrganizationID)
	cloneHost := codeupCloneHost(params.ServerURL)

	var allRepos []Repo

	for _, group := range params.Groups {
		repos, err := p.listGroupRepos(ctx, params.Token, apiBase, cloneHost, group, params.IncludeDeleted)
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
func (p *CodeupProvider) listGroupRepos(ctx context.Context, token, apiBase, cloneHost string, group GroupQuery, includeDeleted bool) ([]Repo, error) {
	// Empty group path: full list, no filtering
	if group.Path == "" {
		return p.listAllReposFull(ctx, token, apiBase, cloneHost, "", includeDeleted)
	}

	// Step 1: try to resolve group path to groupId
	groupID, err := p.resolveGroupID(ctx, token, apiBase, group.Path)
	if err != nil {
		return nil, err
	}

	if groupID > 0 {
		// Step 2: fetch repos by groupId
		return p.listGroupReposByID(ctx, token, apiBase, cloneHost, groupID, group.Recursive, includeDeleted)
	}

	// Fallback: full list + client-side prefix filtering
	return p.listAllReposFull(ctx, token, apiBase, cloneHost, group.Path, includeDeleted)
}

// resolveGroupID searches namespaces for an exact pathWithNamespace match.
// Returns 0 (without error) if no match found (triggers fallback).
func (p *CodeupProvider) resolveGroupID(ctx context.Context, token, apiBase, groupPath string) (int, error) {
	apiURL := fmt.Sprintf("%s/namespaces?search=%s&perPage=100",
		apiBase, url.QueryEscape(groupPath))

	var namespaces []codeupNamespace
	_, err := p.getClient().GetWithPagination(ctx, token, apiURL, &namespaces)
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
func (p *CodeupProvider) listGroupReposByID(ctx context.Context, token, apiBase, cloneHost string, groupID int, recursive, includeDeleted bool) ([]Repo, error) {
	var allRepos []Repo
	page := 1
	skipped := 0

	for {
		apiURL := fmt.Sprintf("%s/groups/%d/repositories?page=%d&perPage=100&includeSubgroups=%v",
			apiBase, groupID, page, recursive)

		var repos []codeupRepo
		pg, err := p.getClient().GetWithPagination(ctx, token, apiURL, &repos)
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			if !includeDeleted && IsDeletionScheduled(r.Name, r.PathWithNamespace) {
				skipped++
				continue
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

	if skipped > 0 {
		config.Verbose("codeup: skipped %d deletion_scheduled repos in group %d", skipped, groupID)
	}

	return allRepos, nil
}

// listAllReposFull fetches all repos via ListRepositories and optionally filters by path prefix.
func (p *CodeupProvider) listAllReposFull(ctx context.Context, token, apiBase, cloneHost, pathPrefix string, includeDeleted bool) ([]Repo, error) {
	var allRepos []Repo
	page := 1
	skipped := 0

	for {
		apiURL := fmt.Sprintf("%s/repositories?page=%d&perPage=100", apiBase, page)

		var repos []codeupRepo
		pg, err := p.getClient().GetWithPagination(ctx, token, apiURL, &repos)
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

			if !includeDeleted && IsDeletionScheduled(r.Name, r.PathWithNamespace) {
				skipped++
				continue
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

	if skipped > 0 {
		scope := pathPrefix
		if scope == "" {
			scope = "organization"
		}
		config.Verbose("codeup: skipped %d deletion_scheduled repos under %s", skipped, scope)
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
		pg, err := p.getClient().GetWithPagination(ctx, params.Token, apiURL, &namespaces)
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
