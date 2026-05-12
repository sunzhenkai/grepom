package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubProvider_ListRepos(t *testing.T) {
	repos := []githubRepo{
		{Name: "api-server", FullName: "my-org/api-server", CloneURL: "https://github.com/my-org/api-server.git", SSHURL: "git@github.com:my-org/api-server.git"},
		{Name: "web-app", FullName: "my-org/web-app", CloneURL: "https://github.com/my-org/web-app.git", SSHURL: "git@github.com:my-org/web-app.git"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/users/my-org" {
			// 返回 Organization 类型，触发 /orgs/ 端点
			json.NewEncoder(w).Encode(githubUserInfo{Type: "Organization"})
		} else if r.URL.Path == "/orgs/my-org/repos" {
			json.NewEncoder(w).Encode(repos)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitHubProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Orgs:      []string{"my-org"},
	}

	result, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(result))
	}

	if result[0].Path != "my-org/api-server" {
		t.Errorf("expected path my-org/api-server, got %s", result[0].Path)
	}
}

func TestGitHubProvider_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer ts.Close()

	p := &GitHubProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "bad-token",
		Orgs:      []string{"my-org"},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestGitHubProvider_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-Remaining", "0")
		w.Header().Set("X-Ratelimit-Reset", "9999999999")
		w.WriteHeader(403)
		w.Write([]byte(`{"message": "rate limit exceeded"}`))
	}))
	defer ts.Close()

	p := &GitHubProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Orgs:      []string{"my-org"},
	}

	_, err := p.ListRepos(context.Background(), params)
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestGitHubProvider_EntityType_Organization(t *testing.T) {
	// Organization 类型应使用 /orgs/ 端点，能发现私有仓库
	repos := []githubRepo{
		{Name: "public-repo", FullName: "my-org/public-repo", CloneURL: "https://github.com/my-org/public-repo.git", Private: false},
		{Name: "private-repo", FullName: "my-org/private-repo", CloneURL: "https://github.com/my-org/private-repo.git", Private: true},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/users/my-org" {
			json.NewEncoder(w).Encode(githubUserInfo{Type: "Organization"})
		} else if r.URL.Path == "/orgs/my-org/repos" {
			json.NewEncoder(w).Encode(repos)
		} else {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitHubProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Orgs:      []string{"my-org"},
	}

	result, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 repos (including private), got %d", len(result))
	}
}

func TestGitHubProvider_EntityType_User(t *testing.T) {
	// User 类型应使用 /users/ 端点
	repos := []githubRepo{
		{Name: "my-project", FullName: "my-user/my-project", CloneURL: "https://github.com/my-user/my-project.git"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/users/my-user" {
			json.NewEncoder(w).Encode(githubUserInfo{Type: "User"})
		} else if r.URL.Path == "/users/my-user/repos" {
			json.NewEncoder(w).Encode(repos)
		} else {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitHubProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Orgs:      []string{"my-user"},
	}

	result, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(result))
	}
	if result[0].Name != "my-project" {
		t.Errorf("expected my-project, got %s", result[0].Name)
	}
}

func TestGitHubProvider_EntityType_NotFound(t *testing.T) {
	// /users/{name} 返回 404 时，应回退到 /orgs/ 端点
	repos := []githubRepo{
		{Name: "fallback-repo", FullName: "unknown/fallback-repo", CloneURL: "https://github.com/unknown/fallback-repo.git"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/users/unknown" {
			w.WriteHeader(404)
		} else if r.URL.Path == "/orgs/unknown/repos" {
			json.NewEncoder(w).Encode(repos)
		} else {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	p := &GitHubProvider{}
	params := ListReposParams{
		ServerURL: ts.URL,
		Token:     "test-token",
		Orgs:      []string{"unknown"},
	}

	result, err := p.ListRepos(context.Background(), params)
	if err != nil {
		t.Fatalf("ListRepos failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 repo from fallback, got %d", len(result))
	}
}
