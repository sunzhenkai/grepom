package mergerequest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubMRProvider_BuildWebURL(t *testing.T) {
	p := &GitHubMRProvider{}

	tests := []struct {
		name     string
		params   WebURLParams
		expected string
	}{
		{
			name: "standard github.com",
			params: WebURLParams{
				ServerURL:    "https://github.com",
				RepoPath:     "myorg/myrepo",
				SourceBranch: "feature-x",
				TargetBranch: "main",
			},
			expected: "https://github.com/myorg/myrepo/compare/main...feature-x?expand=1",
		},
		{
			name: "with draft",
			params: WebURLParams{
				ServerURL:    "https://github.com",
				RepoPath:     "myorg/myrepo",
				SourceBranch: "feature-x",
				TargetBranch: "main",
				Draft:        true,
			},
			expected: "https://github.com/myorg/myrepo/compare/main...feature-x?expand=1&draft=1",
		},
		{
			name: "GHE server",
			params: WebURLParams{
				ServerURL:    "https://ghe.example.com/api/v3",
				RepoPath:     "myorg/myrepo",
				SourceBranch: "feat",
				TargetBranch: "develop",
			},
			expected: "https://ghe.example.com/myorg/myrepo/compare/develop...feat?expand=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.BuildWebURL(tt.params)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestGitHubMRProvider_CreateMergeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/repos/myorg/myrepo/pulls" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var req githubCreatePRRequest
		json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(githubPRResponse{
			ID:     42,
			Number: 7,
			Title:  req.Title,
			Body:   req.Body,
			State:  "open",
			HTMLURL: "https://github.com/myorg/myrepo/pull/7",
			Head:   githubPRBranch{Ref: req.Head},
			Base:   githubPRBranch{Ref: req.Base},
			Draft:  req.Draft,
		})
	}))
	defer server.Close()

	p := &GitHubMRProvider{}
	mr, err := p.CreateMergeRequest(t.Context(), CreateMergeRequestParams{
		ServerURL:    server.URL,
		Token:        "test-token",
		RepoPath:     "myorg/myrepo",
		Title:        "feat: add dark mode",
		Description:  "Implements dark mode toggle",
		SourceBranch: "feature-x",
		TargetBranch: "main",
		Draft:        true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.Number != 7 {
		t.Errorf("expected number 7, got %d", mr.Number)
	}
	if mr.Title != "feat: add dark mode" {
		t.Errorf("unexpected title: %s", mr.Title)
	}
	if mr.URL != "https://github.com/myorg/myrepo/pull/7" {
		t.Errorf("unexpected URL: %s", mr.URL)
	}
	if !mr.Draft {
		t.Error("expected draft=true")
	}
}

func TestGitHubMRProvider_CreateMergeRequest_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(githubErrorResponse{
			Message: "Validation Failed: No commits between main and feature-x",
		})
	}))
	defer server.Close()

	p := &GitHubMRProvider{}
	_, err := p.CreateMergeRequest(t.Context(), CreateMergeRequestParams{
		ServerURL:    server.URL,
		Token:        "test-token",
		RepoPath:     "myorg/myrepo",
		Title:        "test",
		SourceBranch: "feature-x",
		TargetBranch: "main",
	})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGitHubMRProvider_GHEURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com", "https://api.github.com"},
		{"https://ghe.example.com", "https://ghe.example.com"},
		{"https://ghe.example.com/api/v3", "https://ghe.example.com/api/v3"},
	}

	for _, tt := range tests {
		got := githubAPIURL(tt.input)
		if got != tt.expected {
			t.Errorf("githubAPIURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildGitHubWebBaseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://api.github.com", "https://github.com"},
		{"https://ghe.example.com/api/v3", "https://ghe.example.com"},
		{"https://ghe.example.com", "https://ghe.example.com"},
	}

	for _, tt := range tests {
		got := buildGitHubWebBaseURL(tt.input)
		if got != tt.expected {
			t.Errorf("buildGitHubWebBaseURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
