package mergerequest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitLabMRProvider_BuildWebURL(t *testing.T) {
	p := &GitLabMRProvider{}

	tests := []struct {
		name     string
		params   WebURLParams
		expected string
	}{
		{
			name: "standard gitlab.com",
			params: WebURLParams{
				ServerURL:    "https://gitlab.com",
				RepoPath:     "myorg/myrepo",
				SourceBranch: "feature-x",
				TargetBranch: "main",
			},
			expected: "https://gitlab.com/myorg/myrepo/-/merge_requests/new?merge_request[source_branch]=feature-x&merge_request[target_branch]=main",
		},
		{
			name: "with title and draft",
			params: WebURLParams{
				ServerURL:    "https://gitlab.com",
				RepoPath:     "myorg/myrepo",
				SourceBranch: "feature-x",
				TargetBranch: "main",
				Title:        "feat: add dark mode",
				Draft:        true,
			},
			expected: "https://gitlab.com/myorg/myrepo/-/merge_requests/new?merge_request[source_branch]=feature-x&merge_request[target_branch]=main&merge_request[title]=feat%3A+add+dark+mode&merge_request[draft]=true",
		},
		{
			name: "self-hosted gitlab",
			params: WebURLParams{
				ServerURL:    "https://gitlab.mycompany.com",
				RepoPath:     "team/project",
				SourceBranch: "feat",
				TargetBranch: "develop",
			},
			expected: "https://gitlab.mycompany.com/team/project/-/merge_requests/new?merge_request[source_branch]=feat&merge_request[target_branch]=develop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.BuildWebURL(tt.params)
			if got != tt.expected {
				t.Errorf("expected %s\ngot      %s", tt.expected, got)
			}
		})
	}
}

func TestGitLabMRProvider_CreateMergeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("PRIVATE-TOKEN"))
		}

		var req gitlabCreateMRRequest
		json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(gitlabMRResponse{
			ID:           123,
			IID:          45,
			Title:        req.Title,
			Description:  req.Description,
			State:        "opened",
			WebURL:       "https://gitlab.com/myorg/myrepo/-/merge_requests/45",
			SourceBranch: req.SourceBranch,
			TargetBranch: req.TargetBranch,
		})
	}))
	defer server.Close()

	p := &GitLabMRProvider{}
	mr, err := p.CreateMergeRequest(t.Context(), CreateMergeRequestParams{
		ServerURL:    server.URL,
		Token:        "test-token",
		RepoPath:     "myorg/myrepo",
		Title:        "feat: add dark mode",
		Description:  "Implement dark mode",
		SourceBranch: "feature-x",
		TargetBranch: "main",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.Number != 45 {
		t.Errorf("expected number 45, got %d", mr.Number)
	}
	if mr.Title != "feat: add dark mode" {
		t.Errorf("unexpected title: %s", mr.Title)
	}
	if mr.URL != "https://gitlab.com/myorg/myrepo/-/merge_requests/45" {
		t.Errorf("unexpected URL: %s", mr.URL)
	}
}

func TestGitLabMRProvider_CreateMergeRequest_Draft(t *testing.T) {
	var receivedTitle string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gitlabCreateMRRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedTitle = req.Title

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(gitlabMRResponse{
			ID:           124,
			IID:          46,
			Title:        req.Title,
			State:        "opened",
			WebURL:       "https://gitlab.com/myorg/myrepo/-/merge_requests/46",
			SourceBranch: "feature-x",
			TargetBranch: "main",
		})
	}))
	defer server.Close()

	p := &GitLabMRProvider{}
	mr, err := p.CreateMergeRequest(t.Context(), CreateMergeRequestParams{
		ServerURL:    server.URL,
		Token:        "test-token",
		RepoPath:     "myorg/myrepo",
		Title:        "feat: add dark mode",
		SourceBranch: "feature-x",
		TargetBranch: "main",
		Draft:        true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(receivedTitle, "Draft: ") {
		t.Errorf("expected title to start with 'Draft: ', got %q", receivedTitle)
	}
	if !mr.Draft {
		t.Error("expected draft=true")
	}
}

func TestGitLabMRProvider_CreateMergeRequest_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"source_branch is missing"}`))
	}))
	defer server.Close()

	p := &GitLabMRProvider{}
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

func TestGitLabMRProvider_URLEncodedPath(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(gitlabMRResponse{
			IID:    1, Title: "test", State: "opened",
			WebURL: "https://gitlab.com/test", SourceBranch: "f", TargetBranch: "m",
		})
	}))
	defer server.Close()

	p := &GitLabMRProvider{}
	_, err := p.CreateMergeRequest(t.Context(), CreateMergeRequestParams{
		ServerURL:    server.URL,
		Token:        "test-token",
		RepoPath:     "myorg/team/project",
		Title:        "test",
		SourceBranch: "f",
		TargetBranch: "m",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(receivedPath, "/merge_requests") {
		t.Errorf("expected merge_requests in path: %s", receivedPath)
	}
	// url.PathEscape("myorg/team/project") = "myorg%2Fteam%2Fproject"
	if !strings.Contains(receivedPath, "myorg") {
		t.Errorf("expected project path in URL: %s", receivedPath)
	}
}
