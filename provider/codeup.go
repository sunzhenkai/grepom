package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	Register("codeup", func() Provider { return &CodeupProvider{} })
}

// CodeupProvider implements the Provider interface for Alibaba Cloud Codeup (云效).
type CodeupProvider struct {
	client *http.Client
}

// --- Codeup API response structures ---

// codeupResponse is the unified response wrapper for all Codeup APIs.
type codeupResponse struct {
	RequestID    string          `json:"requestId"`
	Success      bool            `json:"success"`
	ErrorCode    string          `json:"errorCode"`
	ErrorMessage string          `json:"errorMessage"`
	Total        int64           `json:"total"`
	Result       json.RawMessage `json:"result"`
}

// codeupRepo maps a single repository from the Codeup ListRepositories API.
type codeupRepo struct {
	ID                int64  `json:"Id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"pathWithNamespace"`
	NameWithNamespace string `json:"nameWithNamespace"`
	Description       string `json:"description"`
	WebURL            string `json:"webUrl"`
	VisibilityLevel   int    `json:"visibilityLevel"`
	AccessLevel       int    `json:"accessLevel"`
	Archive           bool   `json:"archive"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	LastActivityAt    string `json:"lastActivityAt"`
	NamespaceID       int64  `json:"namespaceId"`
}

// codeupGroup maps a single group from the Codeup ListRepositoryGroups API.
type codeupGroup struct {
	ID                int64  `json:"id"`
	Path              string `json:"path"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"pathWithNamespace"`
	NameWithNamespace string `json:"nameWithNamespace"`
	VisibilityLevel   int    `json:"visibilityLevel"`
	AvatarURL         string `json:"avatarUrl"`
	WebURL            string `json:"webUrl"`
	Type              string `json:"type"`
	Description       string `json:"description"`
	ParentID          int64  `json:"parentId"`
	OwnerID           int64  `json:"ownerId"`
	AccessLevel       int    `json:"accessLevel"`
	ProjectCount      int64  `json:"projectCount"`
	GroupCount        int64  `json:"groupCount"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

// codeupGroupDetail maps the response from identityGetGroupByPath API.
type codeupGroupDetail struct {
	ID                int64  `json:"id"`
	Path              string `json:"path"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"pathWithNamespace"`
	NameWithNamespace string `json:"nameWithNamespace"`
	VisibilityLevel   int    `json:"visibilityLevel"`
	Description       string `json:"description"`
	AvatarURL         string `json:"avatarUrl"`
	WebURL            string `json:"webUrl"`
	ParentID          string `json:"parentId"`
	OwnerID           string `json:"ownerId"`
}

func (p *CodeupProvider) getClient() *http.Client {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}
	return p.client
}

// codeupAPIURL maps the user-facing clone URL to the Codeup API base URL.
// For codeup.aliyun.com → https://devops.aliyun.com
func codeupAPIURL(serverURL string) string {
	u := strings.TrimRight(serverURL, "/")
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if u == "codeup.aliyun.com" {
		return "https://devops.aliyun.com"
	}
	// For potential custom deployments, use the URL as-is
	if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
		return serverURL
	}
	return "https://" + serverURL
}

// get makes a GET request to the Codeup API with accessToken as a query parameter.
func (p *CodeupProvider) get(ctx context.Context, token, apiURL string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
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

	// Parse unified response wrapper
	var cr codeupResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if !cr.Success {
		return fmt.Errorf("codeup API error: [%s] %s", cr.ErrorCode, cr.ErrorMessage)
	}

	if err := json.Unmarshal(cr.Result, v); err != nil {
		return fmt.Errorf("decode result: %w", err)
	}

	return nil
}

// getWithTotal makes a GET request and returns both the parsed result and total count.
func (p *CodeupProvider) getWithTotal(ctx context.Context, token, apiURL string, v interface{}) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.getClient().Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return 0, fmt.Errorf("codeup: authentication failed (invalid token)")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("codeup API error %d: %s", resp.StatusCode, string(body))
	}

	var cr codeupResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	if !cr.Success {
		return 0, fmt.Errorf("codeup API error: [%s] %s", cr.ErrorCode, cr.ErrorMessage)
	}

	if err := json.Unmarshal(cr.Result, v); err != nil {
		return 0, fmt.Errorf("decode result: %w", err)
	}

	return cr.Total, nil
}

// --- ListRepos implementation ---

func (p *CodeupProvider) ListRepos(ctx context.Context, params ListReposParams) ([]Repo, error) {
	apiBase := codeupAPIURL(params.ServerURL)
	var allRepos []Repo

	for _, group := range params.Groups {
		repos, err := p.listGroupRepos(ctx, params, apiBase, group)
		if err != nil {
			return nil, fmt.Errorf("codeup: group %s: %w", group.Path, err)
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

// listGroupRepos fetches all repos for an organization and filters by group path prefix.
func (p *CodeupProvider) listGroupRepos(ctx context.Context, params ListReposParams, apiBase string, group GroupQuery) ([]Repo, error) {
	// Determine the clone host from the original server URL
	cloneHost := strings.TrimRight(params.ServerURL, "/")
	cloneHost = strings.TrimPrefix(cloneHost, "https://")
	cloneHost = strings.TrimPrefix(cloneHost, "http://")

	var allCodeupRepos []codeupRepo
	page := int64(1)
	perPage := int64(100)

	for {
		apiURL := fmt.Sprintf("%s/repository/list?organizationId=%s&perPage=%d&page=%d",
			apiBase, url.QueryEscape(params.OrganizationID), perPage, page)
		if params.Token != "" {
			apiURL += "&accessToken=" + url.QueryEscape(params.Token)
		}

		var repos []codeupRepo
		total, err := p.getWithTotal(ctx, "", apiURL, &repos)
		if err != nil {
			return nil, err
		}

		allCodeupRepos = append(allCodeupRepos, repos...)

		// Calculate if there are more pages
		totalPages := int64(math.Ceil(float64(total) / float64(perPage)))
		if page >= totalPages {
			break
		}
		page++
	}

	// Filter by group path prefix if specified
	var filtered []Repo
	for _, r := range allCodeupRepos {
		if group.Path != "" {
			prefix := group.Path + "/"
			if !strings.HasPrefix(r.PathWithNamespace, prefix) {
				continue
			}
		}

		cloneURL := "https://" + cloneHost + "/" + r.PathWithNamespace + ".git"
		sshURL := "git@" + cloneHost + ":" + r.PathWithNamespace + ".git"

		filtered = append(filtered, Repo{
			Name:     r.Name,
			CloneURL: cloneURL,
			SSHURL:   sshURL,
			Path:     r.PathWithNamespace,
			Provider: "codeup",
		})
	}

	return filtered, nil
}

// --- ListGroups implementation ---

func (p *CodeupProvider) ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error) {
	apiBase := codeupAPIURL(params.ServerURL)

	if params.OrganizationID == "" {
		return nil, nil
	}

	// Step 1: Try to get root namespaceId via identityGetGroupByPath
	rootNamespaceID, err := p.getRootNamespaceID(ctx, params.Token, apiBase, params.OrganizationID)
	if err != nil {
		// Graceful degradation: return empty list on failure
		return nil, nil
	}

	// Step 2: List top-level groups under the organization
	return p.listTopLevelGroups(ctx, params.Token, apiBase, params.OrganizationID, rootNamespaceID)
}

// getRootNamespaceID attempts to find the enterprise root namespace ID
// by querying the organization path via identityGetGroupByPath.
func (p *CodeupProvider) getRootNamespaceID(ctx context.Context, token, apiBase, orgID string) (int64, error) {
	apiURL := fmt.Sprintf("%s/api/4/groups/find_by_path?organizationId=%s&identity=%s",
		apiBase, url.QueryEscape(orgID), url.QueryEscape(orgID))

	if token != "" {
		apiURL += "&accessToken=" + url.QueryEscape(token)
	}

	var detail codeupGroupDetail
	if err := p.get(ctx, "", apiURL, &detail); err != nil {
		return 0, err
	}

	return detail.ID, nil
}

// listTopLevelGroups fetches the first-level groups under the organization root.
func (p *CodeupProvider) listTopLevelGroups(ctx context.Context, token, apiBase, orgID string, parentID int64) ([]RemoteGroup, error) {
	var allGroups []RemoteGroup
	page := int64(1)
	pageSize := int64(100)

	for {
		apiURL := fmt.Sprintf("%s/repository/groups/get/all?organizationId=%s&parentId=%d&pageSize=%d&page=%d",
			apiBase, url.QueryEscape(orgID), parentID, pageSize, page)

		if token != "" {
			apiURL += "&accessToken=" + url.QueryEscape(token)
		}

		var groups []codeupGroup
		total, err := p.getWithTotal(ctx, "", apiURL, &groups)
		if err != nil {
			return nil, err
		}

		for _, g := range groups {
			allGroups = append(allGroups, RemoteGroup{
				Name:     g.Name,
				Path:     g.PathWithNamespace,
				Provider: "codeup",
			})
		}

		totalPages := int64(math.Ceil(float64(total) / float64(pageSize)))
		if page >= totalPages {
			break
		}
		page++
	}

	return allGroups, nil
}
