package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- 4.1 codeupAPIBaseURL tests ---

func TestCodeupAPIBaseURL(t *testing.T) {
	orgID := "test-org"
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"codeup.aliyun.com maps to openapi-rdc", "codeup.aliyun.com", "https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/test-org"},
		{"https://codeup.aliyun.com", "https://codeup.aliyun.com", "https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/test-org"},
		{"http://codeup.aliyun.com", "http://codeup.aliyun.com", "https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/test-org"},
		{"custom host preserves http", "http://custom.example.com", "http://custom.example.com/oapi/v1/codeup/organizations/test-org"},
		{"custom host with https", "https://custom.example.com", "https://custom.example.com/oapi/v1/codeup/organizations/test-org"},
		{"plain host defaults to https", "custom.example.com", "https://custom.example.com/oapi/v1/codeup/organizations/test-org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := codeupAPIBaseURL(tt.input, orgID)
			if result != tt.expected {
				t.Errorf("codeupAPIBaseURL(%q, %q) = %q, want %q", tt.input, orgID, result, tt.expected)
			}
		})
	}
}

// --- 4.2 ListRepos: two-step query with group path resolution ---

func TestCodeupProvider_ListRepos_PathFilter(t *testing.T) {
	orgID := "org123"
	groupPath := "wii/solo"
	groupID := 500

	// Mock repos returned by ListGroupRepositories
	groupRepos := []codeupRepo{
		{Name: "grepom", Path: "grepom", PathWithNamespace: "wii/solo/grepom"},
		{Name: "lib", Path: "lib", PathWithNamespace: "wii/solo/lib"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Verify x-yunxiao-token header
		if r.Header.Get("x-yunxiao-token") != "test-token" {
			t.Errorf("missing or wrong x-yunxiao-token header: %s", r.Header.Get("x-yunxiao-token"))
		}

		path := r.URL.Path

		if strings.Contains(path, "/namespaces") {
			// resolveGroupID: return namespace matching groupPath
			namespaces := []codeupNamespace{
				{ID: groupID, Path: "solo", PathWithNamespace: groupPath},
				{ID: 501, Path: "other", PathWithNamespace: "wii/other"},
			}
			w.Header().Set("x-total", "2")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(namespaces)
		} else if strings.Contains(path, fmt.Sprintf("/groups/%d/repositories", groupID)) {
			// ListGroupRepositories
			w.Header().Set("x-total", "2")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(groupRepos)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "unexpected path: %s", path)
		}
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: orgID,
		Groups:         []GroupQuery{{Path: groupPath}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].Name != "grepom" {
		t.Errorf("expected repo name 'grepom', got %s", repos[0].Name)
	}
	if repos[0].Path != "wii/solo/grepom" {
		t.Errorf("expected path 'wii/solo/grepom', got %s", repos[0].Path)
	}

	// Verify clone URLs use original serverURL as host
	cloneHost := strings.TrimPrefix(strings.TrimPrefix(ts.URL, "https://"), "http://")
	expectedHTTPS := "https://" + cloneHost + "/wii/solo/grepom.git"
	if repos[0].CloneURL != expectedHTTPS {
		t.Errorf("expected clone URL %s, got %s", expectedHTTPS, repos[0].CloneURL)
	}
	expectedSSH := "git@" + cloneHost + ":wii/solo/grepom.git"
	if repos[0].SSHURL != expectedSSH {
		t.Errorf("expected SSH URL %s, got %s", expectedSSH, repos[0].SSHURL)
	}
}

// --- 4.3 ListRepos: pagination via x-next-page ---

func TestCodeupProvider_ListRepos_Pagination(t *testing.T) {
	groupID := 100

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page := r.URL.Query().Get("page")

		if strings.Contains(r.URL.Path, "/namespaces") {
			// resolveGroupID
			namespaces := []codeupNamespace{
				{ID: groupID, PathWithNamespace: "org"},
			}
			w.Header().Set("x-total", "1")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(namespaces)
			return
		}

		// ListGroupRepositories: return 3 repos across 2 pages
		if page == "1" || page == "" {
			result := []codeupRepo{
				{Name: "repo-1", PathWithNamespace: "org/repo-1"},
				{Name: "repo-2", PathWithNamespace: "org/repo-2"},
			}
			w.Header().Set("x-total", "3")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-next-page", "2")
			w.Header().Set("x-per-page", "2")
			json.NewEncoder(w).Encode(result)
		} else {
			result := []codeupRepo{
				{Name: "repo-3", PathWithNamespace: "org/repo-3"},
			}
			w.Header().Set("x-total", "3")
			w.Header().Set("x-page", "2")
			w.Header().Set("x-per-page", "2")
			// No x-next-page → last page
			json.NewEncoder(w).Encode(result)
		}
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: "org123",
		Groups:         []GroupQuery{{Path: "org"}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 3 {
		t.Fatalf("expected 3 repos across 2 pages, got %d", len(repos))
	}
}

func TestCodeupProvider_ListRepos_MultiPage(t *testing.T) {
	groupID := 200
	callCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/namespaces") {
			namespaces := []codeupNamespace{
				{ID: groupID, PathWithNamespace: "org"},
			}
			w.Header().Set("x-total", "1")
			json.NewEncoder(w).Encode(namespaces)
			return
		}

		callCount++
		page := r.URL.Query().Get("page")

		var result []codeupRepo
		if page == "1" || page == "" {
			for i := 0; i < 100; i++ {
				result = append(result, codeupRepo{Name: fmt.Sprintf("repo-%d", i), PathWithNamespace: fmt.Sprintf("org/repo-%d", i)})
			}
			w.Header().Set("x-total", "250")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-next-page", "2")
			w.Header().Set("x-per-page", "100")
		} else if page == "2" {
			for i := 100; i < 200; i++ {
				result = append(result, codeupRepo{Name: fmt.Sprintf("repo-%d", i), PathWithNamespace: fmt.Sprintf("org/repo-%d", i)})
			}
			w.Header().Set("x-total", "250")
			w.Header().Set("x-page", "2")
			w.Header().Set("x-next-page", "3")
			w.Header().Set("x-per-page", "100")
		} else {
			for i := 200; i < 250; i++ {
				result = append(result, codeupRepo{Name: fmt.Sprintf("repo-%d", i), PathWithNamespace: fmt.Sprintf("org/repo-%d", i)})
			}
			w.Header().Set("x-total", "250")
			w.Header().Set("x-page", "3")
			w.Header().Set("x-per-page", "100")
			// No x-next-page → last page
		}
		json.NewEncoder(w).Encode(result)
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: "org123",
		Groups:         []GroupQuery{{Path: "org"}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 250 {
		t.Fatalf("expected 250 repos across 3 pages, got %d", len(repos))
	}

	if callCount != 3 {
		t.Errorf("expected 3 ListGroupRepositories API calls, got %d", callCount)
	}
}

// --- 4.4 ListRepos: auth failure and API error ---

func TestCodeupProvider_ListRepos_AuthFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "bad-token",
		OrganizationID: "org123",
		Groups:         []GroupQuery{{Path: "org"}},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestCodeupProvider_ListRepos_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: "org123",
		Groups:         []GroupQuery{{Path: "org"}},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected API error")
	}
	if !strings.Contains(err.Error(), "codeup API error 500") {
		t.Errorf("expected error with status code, got: %v", err)
	}
}

// --- 4.5 ListGroups: using ListNamespaces ---

func TestCodeupProvider_ListGroups(t *testing.T) {
	orgID := "60de7a6852743a5162b5f957"

	namespaces := []codeupNamespace{
		{ID: 100, Path: "frontend", PathWithNamespace: "org/frontend", Name: "Frontend"},
		{ID: 101, Path: "backend", PathWithNamespace: "org/backend", Name: "Backend"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Verify auth header
		if r.Header.Get("x-yunxiao-token") != "test-token" {
			t.Errorf("missing x-yunxiao-token header")
		}

		if !strings.Contains(r.URL.Path, "/namespaces") {
			w.WriteHeader(404)
			fmt.Fprintf(w, "unexpected path: %s", r.URL.Path)
			return
		}

		w.Header().Set("x-total", "2")
		w.Header().Set("x-page", "1")
		w.Header().Set("x-per-page", "100")
		json.NewEncoder(w).Encode(namespaces)
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListGroupsParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: orgID,
	}

	groups, err := p.ListGroups(context.Background(), params)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	names := map[string]bool{}
	for _, g := range groups {
		names[g.Name] = true
		if g.Provider != "codeup" {
			t.Errorf("expected provider 'codeup', got %s", g.Provider)
		}
	}
	for _, name := range []string{"frontend", "backend"} {
		if !names[name] {
			t.Errorf("missing group %s", name)
		}
	}
}

func TestCodeupProvider_ListGroups_APINamespacesFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListGroupsParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: "bad-org",
	}

	groups, err := p.ListGroups(context.Background(), params)
	if err != nil {
		t.Fatalf("expected graceful degradation (nil error), got: %v", err)
	}
	if groups != nil {
		t.Errorf("expected nil groups on API failure, got %d groups", len(groups))
	}
}

// --- 4.6 ListRepos: fallback to full list when namespace not found ---

func TestCodeupProvider_ListRepos_FallbackToFullList(t *testing.T) {
	orgID := "org123"
	groupPath := "nonexistent"

	// All repos in org (returned by fallback ListRepositories)
	allRepos := []codeupRepo{
		{Name: "repo-a", PathWithNamespace: "wii/solo/repo-a"},
		{Name: "repo-b", PathWithNamespace: "other/team/repo-b"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/namespaces") {
			// Return namespaces that don't match the group path → triggers fallback
			namespaces := []codeupNamespace{
				{ID: 999, PathWithNamespace: "other/path"},
			}
			w.Header().Set("x-total", "1")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(namespaces)
		} else if strings.Contains(path, "/repositories") && !strings.Contains(path, "/groups/") {
			// Fallback: ListRepositories (full list)
			w.Header().Set("x-total", "2")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(allRepos)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "unexpected path: %s", path)
		}
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: orgID,
		Groups:         []GroupQuery{{Path: groupPath}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	// No repos match "nonexistent/" prefix, so result should be empty
	if len(repos) != 0 {
		t.Fatalf("expected 0 repos (no prefix match), got %d", len(repos))
	}
}

// --- 4.7 ListRepos: includeSubgroups for recursive ---

func TestCodeupProvider_ListRepos_Recursive(t *testing.T) {
	orgID := "org123"
	groupPath := "wii"
	groupID := 300

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/namespaces") {
			namespaces := []codeupNamespace{
				{ID: groupID, PathWithNamespace: groupPath},
			}
			w.Header().Set("x-total", "1")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(namespaces)
		} else if strings.Contains(path, fmt.Sprintf("/groups/%d/repositories", groupID)) {
			// Verify includeSubgroups parameter
			includeSubgroups := r.URL.Query().Get("includeSubgroups")
			if includeSubgroups != "true" {
				t.Errorf("expected includeSubgroups=true for recursive, got %s", includeSubgroups)
			}

			repos := []codeupRepo{
				{Name: "main", PathWithNamespace: "wii/main"},
				{Name: "sub-repo", PathWithNamespace: "wii/sub/deep"},
			}
			w.Header().Set("x-total", "2")
			w.Header().Set("x-page", "1")
			w.Header().Set("x-per-page", "100")
			json.NewEncoder(w).Encode(repos)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "unexpected path: %s", path)
		}
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: orgID,
		Groups:         []GroupQuery{{Path: groupPath, Recursive: true}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos (recursive), got %d", len(repos))
	}
}

func TestCodeupProvider_ListRepos_NonRecursive(t *testing.T) {
	orgID := "org123"
	groupPath := "wii"
	groupID := 300

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/namespaces") {
			namespaces := []codeupNamespace{
				{ID: groupID, PathWithNamespace: groupPath},
			}
			w.Header().Set("x-total", "1")
			json.NewEncoder(w).Encode(namespaces)
		} else if strings.Contains(path, fmt.Sprintf("/groups/%d/repositories", groupID)) {
			// Verify includeSubgroups parameter is false
			includeSubgroups := r.URL.Query().Get("includeSubgroups")
			if includeSubgroups != "false" {
				t.Errorf("expected includeSubgroups=false for non-recursive, got %s", includeSubgroups)
			}

			repos := []codeupRepo{
				{Name: "main", PathWithNamespace: "wii/main"},
			}
			w.Header().Set("x-total", "1")
			json.NewEncoder(w).Encode(repos)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: orgID,
		Groups:         []GroupQuery{{Path: groupPath, Recursive: false}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo (non-recursive), got %d", len(repos))
	}
}

// mustMarshal is a test helper that marshals v to json.RawMessage.
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(data)
}
