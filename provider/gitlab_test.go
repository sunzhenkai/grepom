package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGitLabProvider_ListRepos_Recursive(t *testing.T) {
	group := gitlabGroup{ID: 123, Path: "frontend", FullPath: "my-org/frontend"}
	projects := []gitlabProject{
		{Name: "web-app", PathWithNamespace: "my-org/frontend/web-app", HTTPURLToRepo: "https://gitlab.com/my-org/frontend/web-app.git", SSHURLToRepo: "git@gitlab.com:my-org/frontend/web-app.git"},
		{Name: "api", PathWithNamespace: "my-org/frontend/api", HTTPURLToRepo: "https://gitlab.com/my-org/frontend/api.git", SSHURLToRepo: "git@gitlab.com:my-org/frontend/api.git"},
	}
	subgroups := []gitlabGroup{
		{ID: 456, Path: "components", FullPath: "my-org/frontend/components"},
	}
	subProjects := []gitlabProject{
		{Name: "ui-lib", PathWithNamespace: "my-org/frontend/components/ui-lib", HTTPURLToRepo: "https://gitlab.com/my-org/frontend/components/ui-lib.git", SSHURLToRepo: "git@gitlab.com:my-org/frontend/components/ui-lib.git"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/groups/") && !strings.Contains(path, "/projects") && !strings.Contains(path, "/subgroups") {
			json.NewEncoder(w).Encode(group)
		} else if path == "/api/v4/groups/123/projects" {
			json.NewEncoder(w).Encode(projects)
		} else if path == "/api/v4/groups/123/subgroups" {
			json.NewEncoder(w).Encode(subgroups)
		} else if path == "/api/v4/groups/456/projects" {
			json.NewEncoder(w).Encode(subProjects)
		} else if path == "/api/v4/groups/456/subgroups" {
			json.NewEncoder(w).Encode([]gitlabGroup{})
		} else {
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Groups:    []GroupQuery{{Path: "my-org/frontend", Recursive: true}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}

	names := map[string]bool{}
	for _, r := range repos {
		names[r.Name] = true
	}
	for _, name := range []string{"web-app", "api", "ui-lib"} {
		if !names[name] {
			t.Errorf("missing repo %s", name)
		}
	}

	// Verify path hierarchy preserved
	for _, r := range repos {
		if r.Name == "ui-lib" {
			expected := "my-org/frontend/components/ui-lib"
			if r.Path != expected {
				t.Errorf("expected path %s, got %s", expected, r.Path)
			}
		}
	}
}

func TestGitLabProvider_ListRepos_FiltersSharedProjects(t *testing.T) {
	group := gitlabGroup{ID: 123, Path: "topon-bidder", FullPath: "topon-bidder"}
	projects := []gitlabProject{
		{Name: "ad-platform-wiki", PathWithNamespace: "topon-bidder/ad-platform/ad-platform-wiki", HTTPURLToRepo: "https://gitlab.com/topon-bidder/ad-platform/ad-platform-wiki.git"},
		{Name: "Ai Coding", PathWithNamespace: "zhangfeixiang/ai-coding", HTTPURLToRepo: "https://gitlab.com/zhangfeixiang/ai-coding.git"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/groups/") && !strings.Contains(path, "/projects") && !strings.Contains(path, "/subgroups") {
			json.NewEncoder(w).Encode(group)
		} else if path == "/api/v4/groups/123/projects" {
			json.NewEncoder(w).Encode(projects)
		} else if path == "/api/v4/groups/123/subgroups" {
			json.NewEncoder(w).Encode([]gitlabGroup{})
		} else {
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Groups:    []GroupQuery{{Path: "topon-bidder", Recursive: true}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo (shared project filtered), got %d", len(repos))
	}
	if repos[0].Name != "ad-platform-wiki" {
		t.Errorf("expected ad-platform-wiki, got %q", repos[0].Name)
	}
}

func TestGitLabProvider_NonRecursive(t *testing.T) {
	group := gitlabGroup{ID: 123, Path: "frontend", FullPath: "my-org/frontend"}
	projects := []gitlabProject{
		{Name: "web-app", PathWithNamespace: "my-org/frontend/web-app"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(path, "/groups/") && !strings.Contains(path, "/projects") && !strings.Contains(path, "/subgroups") {
			json.NewEncoder(w).Encode(group)
		} else if path == "/api/v4/groups/123/projects" {
			json.NewEncoder(w).Encode(projects)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Groups:    []GroupQuery{{Path: "my-org/frontend", Recursive: false}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo (non-recursive), got %d", len(repos))
	}
}

func TestGitLabProvider_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message": "401 Unauthorized"}`))
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "bad-token",
		Groups:    []GroupQuery{{Path: "my-org"}},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestGitLabProvider_ListRepos_FallbackToUserNamespace(t *testing.T) {
	groupPath := "sunzhenkai"
	user := gitlabUser{ID: 99, Username: "sunzhenkai", Name: "Sun Zhenkai"}
	projects := []gitlabProject{
		{Name: "grepom", PathWithNamespace: "sunzhenkai/grepom", HTTPURLToRepo: "https://gitlab.com/sunzhenkai/grepom.git"},
		{Name: "shared", PathWithNamespace: "other/shared", HTTPURLToRepo: "https://gitlab.com/other/shared.git"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v4/groups/sunzhenkai":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"404 Group Not Found"}`))
		case r.URL.Path == "/api/v4/users":
			if r.URL.Query().Get("username") != groupPath {
				t.Fatalf("unexpected username query: %q", r.URL.RawQuery)
			}
			json.NewEncoder(w).Encode([]gitlabUser{user})
		case r.URL.Path == "/api/v4/users/99/projects":
			json.NewEncoder(w).Encode(projects)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Groups:    []GroupQuery{{Path: groupPath}},
	}

	repos, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo after namespace filtering, got %d", len(repos))
	}
	if repos[0].Path != "sunzhenkai/grepom" {
		t.Fatalf("expected user namespace path, got %q", repos[0].Path)
	}
}

func TestGitLabProvider_ListRepos_PathNotFoundOrInaccessible(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v4/groups/missing":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"404 Group Not Found"}`))
		case "/api/v4/users":
			json.NewEncoder(w).Encode([]gitlabUser{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Groups:    []GroupQuery{{Path: "missing"}},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(err.Error(), "not found or inaccessible") {
		t.Fatalf("expected normalized path error, got %v", err)
	}
}

func TestGitLabProvider_ListRepos_GroupAuthFailureNoFallback(t *testing.T) {
	userAPICalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v4/groups/private":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"401 Unauthorized"}`))
		case "/api/v4/users":
			userAPICalled = true
			json.NewEncoder(w).Encode([]gitlabUser{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "bad-token",
		Groups:    []GroupQuery{{Path: "private"}},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected auth error")
	}
	if userAPICalled {
		t.Fatal("unexpected user fallback for non-404 group error")
	}
}

func TestGitLabProvider_getUserByUsername_UsesEscapedQuery(t *testing.T) {
	receivedRawQuery := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v4/users" {
			receivedRawQuery = r.URL.RawQuery
			json.NewEncoder(w).Encode([]gitlabUser{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	p := &GitLabProvider{}
	_, _ = p.getUserByUsername(context.Background(), ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
	}, "sun zhenkai")

	if !strings.Contains(receivedRawQuery, "username="+url.QueryEscape("sun zhenkai")) {
		t.Fatalf("expected escaped username query, got %q", receivedRawQuery)
	}
}
