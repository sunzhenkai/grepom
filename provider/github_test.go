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

		if r.URL.Path == "/orgs/my-org/repos" {
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
