package cicd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubPipelineProvider_ListPipelines(t *testing.T) {
	response := githubWorkflowRunsResponse{
		TotalCount: 2,
		WorkflowRuns: []githubWorkflowRun{
			{
				ID:           5678,
				Name:         "CI",
				HeadBranch:   "main",
				HeadSHA:      "abc1234567890def",
				Status:       "completed",
				Conclusion:   "success",
				CreatedAt:    "2026-04-21T10:00:00Z",
				UpdatedAt:    "2026-04-21T10:02:34Z",
				RunStartedAt: "2026-04-21T10:00:00Z",
				HTMLURL:      "https://github.com/owner/repo/actions/runs/5678",
			},
			{
				ID:           5677,
				Name:         "CI",
				HeadBranch:   "feat-x",
				HeadSHA:      "def5678901234abc",
				Status:       "in_progress",
				Conclusion:   "",
				CreatedAt:    "2026-04-21T09:30:00Z",
				UpdatedAt:    "2026-04-21T09:31:23Z",
				RunStartedAt: "2026-04-21T09:30:00Z",
				HTMLURL:      "https://github.com/owner/repo/actions/runs/5677",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer token, got %v", r.Header.Get("Authorization"))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/repos/owner/repo/actions/runs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := &GitHubPipelineProvider{}
	pipelines, err := provider.ListPipelines(context.Background(), ListPipelinesParams{
		ServerURL: server.URL,
		Token:     "test-token",
		RepoPath:  "owner/repo",
		Limit:     5,
	})
	if err != nil {
		t.Fatalf("ListPipelines returned error: %v", err)
	}

	if len(pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(pipelines))
	}

	// First pipeline: completed+success
	if pipelines[0].ID != 5678 {
		t.Errorf("pipeline[0].ID = %d, want 5678", pipelines[0].ID)
	}
	if pipelines[0].Status != StatusSuccess {
		t.Errorf("pipeline[0].Status = %s, want success", pipelines[0].Status)
	}
	if pipelines[0].SHA != "abc1234" {
		t.Errorf("pipeline[0].SHA = %s, want abc1234", pipelines[0].SHA)
	}
	if pipelines[0].Duration != 154*time.Second {
		t.Errorf("pipeline[0].Duration = %v, want 154s", pipelines[0].Duration)
	}

	// Second pipeline: in_progress
	if pipelines[1].Status != StatusRunning {
		t.Errorf("pipeline[1].Status = %s, want running", pipelines[1].Status)
	}
	if pipelines[1].Branch != "feat-x" {
		t.Errorf("pipeline[1].Branch = %s, want feat-x", pipelines[1].Branch)
	}
}

func TestGitHubPipelineProvider_GetPipeline(t *testing.T) {
	response := githubWorkflowRun{
		ID:           5678,
		Name:         "CI",
		HeadBranch:   "main",
		HeadSHA:      "abc1234567890def",
		Status:       "in_progress",
		Conclusion:   "",
		RunStartedAt: "2026-04-21T10:00:00Z",
		HTMLURL:      "https://github.com/owner/repo/actions/runs/5678",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/actions/runs/5678" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := &GitHubPipelineProvider{}
	pipeline, err := provider.GetPipeline(context.Background(), GetPipelineParams{
		ServerURL:  server.URL,
		Token:      "test-token",
		RepoPath:   "owner/repo",
		PipelineID: 5678,
	})
	if err != nil {
		t.Fatalf("GetPipeline returned error: %v", err)
	}

	if pipeline.Status != StatusRunning {
		t.Errorf("Status = %s, want running", pipeline.Status)
	}
}

func TestGitHubPipelineProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	provider := &GitHubPipelineProvider{}
	_, err := provider.ListPipelines(context.Background(), ListPipelinesParams{
		ServerURL: server.URL,
		Token:     "test-token",
		RepoPath:  "nonexistent/repo",
		Limit:     5,
	})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestMapGitHubStatus(t *testing.T) {
	tests := []struct {
		status     string
		conclusion string
		want       PipelineStatus
	}{
		{"in_progress", "", StatusRunning},
		{"queued", "", StatusPending},
		{"waiting", "", StatusPending},
		{"completed", "success", StatusSuccess},
		{"completed", "failure", StatusFailed},
		{"completed", "cancelled", StatusCanceled},
		{"completed", "timed_out", StatusFailed},
		{"completed", "neutral", StatusCanceled},
	}
	for _, tt := range tests {
		if got := mapGitHubStatus(tt.status, tt.conclusion); got != tt.want {
			t.Errorf("mapGitHubStatus(%q, %q) = %q, want %q", tt.status, tt.conclusion, got, tt.want)
		}
	}
}

func TestGithubAPIURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com", "https://api.github.com"},
		{"http://github.com", "https://api.github.com"},
		{"https://github.com/", "https://api.github.com"},
		{"https://ghes.example.com", "https://ghes.example.com"},
	}
	for _, tt := range tests {
		if got := githubAPIURL(tt.input); got != tt.want {
			t.Errorf("githubAPIURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
