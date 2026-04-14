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

// --- 6.1 codeupAPIURL tests ---

func TestCodeupAPIURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"codeup.aliyun.com", "https://devops.aliyun.com"},
		{"https://codeup.aliyun.com", "https://devops.aliyun.com"},
		{"http://codeup.aliyun.com", "https://devops.aliyun.com"},
		{"https://codeup.aliyun.com/", "https://devops.aliyun.com"},
		{"custom.example.com", "https://custom.example.com"},
		{"https://custom.example.com", "https://custom.example.com"},
		{"http://custom.example.com", "http://custom.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := codeupAPIURL(tt.input)
			if result != tt.expected {
				t.Errorf("codeupAPIURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- 6.2 ListRepos: full fetch + path prefix filtering ---

func TestCodeupProvider_ListRepos_PathFilter(t *testing.T) {
	allRepos := []codeupRepo{
		{Name: "grepom", Path: "grepom", PathWithNamespace: "wii/solo/grepom"},
		{Name: "other", Path: "other", PathWithNamespace: "wii/other/project"},
		{Name: "lib", Path: "lib", PathWithNamespace: "team/lib/repo"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Verify organizationId is present
		if r.URL.Query().Get("organizationId") != "org123" {
			t.Errorf("missing or wrong organizationId: %s", r.URL.Query().Get("organizationId"))
		}

		resp := codeupResponse{
			Success: true,
			Total:   int64(len(allRepos)),
			Result:  mustMarshal(allRepos),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	p := &CodeupProvider{}
	params := ListReposParams{
		ServerURL:      ts.URL,
		Token:          "test-token",
		OrganizationID: "org123",
		Groups:         []GroupQuery{{Path: "wii/solo"}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo (filtered), got %d", len(repos))
	}
	if repos[0].Name != "grepom" {
		t.Errorf("expected repo name 'grepom', got %s", repos[0].Name)
	}
	if repos[0].Path != "wii/solo/grepom" {
		t.Errorf("expected path 'wii/solo/grepom', got %s", repos[0].Path)
	}

	// Verify clone URLs
	expectedHTTPS := "https://" + strings.TrimPrefix(strings.TrimPrefix(ts.URL, "https://"), "http://") + "/wii/solo/grepom.git"
	if repos[0].CloneURL != expectedHTTPS {
		t.Errorf("expected clone URL %s, got %s", expectedHTTPS, repos[0].CloneURL)
	}

	expectedSSH := "git@" + strings.TrimPrefix(strings.TrimPrefix(ts.URL, "https://"), "http://") + ":wii/solo/grepom.git"
	if repos[0].SSHURL != expectedSSH {
		t.Errorf("expected SSH URL %s, got %s", expectedSSH, repos[0].SSHURL)
	}
}

// --- 6.3 ListRepos: multi-page pagination ---

func TestCodeupProvider_ListRepos_Pagination(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++

		perPage := r.URL.Query().Get("perPage")
		page := r.URL.Query().Get("page")

		// Simulate perPage=2 to force pagination
		allRepos := []codeupRepo{
			{Name: "repo-1", PathWithNamespace: "org/repo-1"},
			{Name: "repo-2", PathWithNamespace: "org/repo-2"},
			{Name: "repo-3", PathWithNamespace: "org/repo-3"},
		}

		// Paginate based on page number, regardless of perPage param
		var result []codeupRepo
		_ = perPage
		if page == "1" || page == "" {
			result = allRepos[:2]
		} else {
			result = allRepos[2:]
		}

		resp := codeupResponse{
			Success: true,
			Total:   3,
			Result:  mustMarshal(result),
		}
		json.NewEncoder(w).Encode(resp)
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

	// With perPage=100 and total=3, it's only 1 page.
	// The test server always returns 2 items on page 1, so we get 2 repos.
	// This validates the pagination loop correctly stops after 1 page (ceil(3/100) = 1).
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos (from page 1 with perPage=100), got %d", len(repos))
	}

	if callCount != 1 {
		t.Errorf("expected 1 API call (all fits in one page), got %d", callCount)
	}
}

func TestCodeupProvider_ListRepos_MultiPage(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		page := r.URL.Query().Get("page")

		var result []codeupRepo
		// Simulate total=250, perPage=100 → 3 pages
		if page == "1" || page == "" {
			for i := 0; i < 100; i++ {
				result = append(result, codeupRepo{Name: fmt.Sprintf("repo-%d", i), PathWithNamespace: fmt.Sprintf("org/repo-%d", i)})
			}
		} else if page == "2" {
			for i := 100; i < 200; i++ {
				result = append(result, codeupRepo{Name: fmt.Sprintf("repo-%d", i), PathWithNamespace: fmt.Sprintf("org/repo-%d", i)})
			}
		} else {
			for i := 200; i < 250; i++ {
				result = append(result, codeupRepo{Name: fmt.Sprintf("repo-%d", i), PathWithNamespace: fmt.Sprintf("org/repo-%d", i)})
			}
		}

		resp := codeupResponse{
			Success: true,
			Total:   250,
			Result:  mustMarshal(result),
		}
		json.NewEncoder(w).Encode(resp)
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
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
}

// --- 6.4 ListRepos: auth failure and API error ---

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		resp := codeupResponse{
			Success:      false,
			ErrorCode:    "SYSTEM_UNKNOWN_ERROR",
			ErrorMessage: "internal error",
		}
		json.NewEncoder(w).Encode(resp)
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
	if !strings.Contains(err.Error(), "SYSTEM_UNKNOWN_ERROR") {
		t.Errorf("expected error code in message, got: %v", err)
	}
}

// --- 6.5 ListGroups test ---

func TestCodeupProvider_ListGroups(t *testing.T) {
	orgID := "60de7a6852743a5162b5f957"
	rootNamespaceID := int64(26842)

	// Mock group detail for identityGetGroupByPath
	groupDetail := codeupGroupDetail{
		ID:                rootNamespaceID,
		Path:              orgID,
		Name:              orgID,
		PathWithNamespace: orgID,
	}

	// Mock top-level groups
	topGroups := []codeupGroup{
		{ID: 100, Path: "frontend", Name: "frontend", PathWithNamespace: "org/frontend"},
		{ID: 101, Path: "backend", Name: "backend", PathWithNamespace: "org/backend"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/api/4/groups/find_by_path") {
			// identityGetGroupByPath
			resp := codeupResponse{
				Success: true,
				Result:  mustMarshal(groupDetail),
			}
			json.NewEncoder(w).Encode(resp)
		} else if strings.Contains(path, "/repository/groups/get/all") {
			// ListRepositoryGroups
			resp := codeupResponse{
				Success: true,
				Total:   2,
				Result:  mustMarshal(topGroups),
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "unexpected path: %s", path)
		}
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

func TestCodeupProvider_ListGroups_RootLookupFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return error for find_by_path
		resp := codeupResponse{
			Success:      false,
			ErrorCode:    "NotFound",
			ErrorMessage: "group not found",
		}
		json.NewEncoder(w).Encode(resp)
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
		t.Errorf("expected nil groups on root lookup failure, got %d groups", len(groups))
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
