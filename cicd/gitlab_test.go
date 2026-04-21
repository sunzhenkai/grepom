package cicd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitLabPipelineProvider_ListPipelines(t *testing.T) {
	response := []gitlabPipeline{
		{
			ID:        1234,
			SHA:       "abc1234567890",
			Ref:       "main",
			Status:    "success",
			StartedAt: "2026-04-21T10:00:00.000Z",
			Duration:  154.0,
			WebURL:    "https://gitlab.com/org/repo/-/pipelines/1234",
		},
		{
			ID:        1233,
			SHA:       "def5678901234",
			Ref:       "main",
			Status:    "failed",
			StartedAt: "2026-04-21T09:00:00.000Z",
			Duration:  312.0,
			WebURL:    "https://gitlab.com/org/repo/-/pipelines/1233",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
			t.Errorf("expected PRIVATE-TOKEN header, got %v", r.Header.Get("PRIVATE-TOKEN"))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/api/v4/projects/org/repo/pipelines" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.URL.Query().Get("per_page") != "5" {
			t.Errorf("unexpected per_page: %s", r.URL.Query().Get("per_page"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := &GitLabPipelineProvider{}
	pipelines, err := provider.ListPipelines(context.Background(), ListPipelinesParams{
		ServerURL: server.URL,
		Token:     "test-token",
		RepoPath:  "org/repo",
		Limit:     5,
	})
	if err != nil {
		t.Fatalf("ListPipelines returned error: %v", err)
	}

	if len(pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(pipelines))
	}

	if pipelines[0].ID != 1234 {
		t.Errorf("pipeline[0].ID = %d, want 1234", pipelines[0].ID)
	}
	if pipelines[0].Status != StatusSuccess {
		t.Errorf("pipeline[0].Status = %s, want success", pipelines[0].Status)
	}
	if pipelines[0].SHA != "abc1234" {
		t.Errorf("pipeline[0].SHA = %s, want abc1234", pipelines[0].SHA)
	}
	if pipelines[0].Branch != "main" {
		t.Errorf("pipeline[0].Branch = %s, want main", pipelines[0].Branch)
	}
	if pipelines[0].Duration != 154*time.Second {
		t.Errorf("pipeline[0].Duration = %v, want 154s", pipelines[0].Duration)
	}

	if pipelines[1].Status != StatusFailed {
		t.Errorf("pipeline[1].Status = %s, want failed", pipelines[1].Status)
	}
}

func TestGitLabPipelineProvider_GetPipeline(t *testing.T) {
	response := gitlabPipeline{
		ID:        1234,
		SHA:       "abc1234567890",
		Ref:       "main",
		Status:    "running",
		StartedAt: "2026-04-21T10:00:00.000Z",
		Duration:  0,
		WebURL:    "https://gitlab.com/org/repo/-/pipelines/1234",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/projects/org/repo/pipelines/1234" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := &GitLabPipelineProvider{}
	pipeline, err := provider.GetPipeline(context.Background(), GetPipelineParams{
		ServerURL:  server.URL,
		Token:      "test-token",
		RepoPath:   "org/repo",
		PipelineID: 1234,
	})
	if err != nil {
		t.Fatalf("GetPipeline returned error: %v", err)
	}

	if pipeline.Status != StatusRunning {
		t.Errorf("Status = %s, want running", pipeline.Status)
	}
	if pipeline.ID != 1234 {
		t.Errorf("ID = %d, want 1234", pipeline.ID)
	}
}

func TestGitLabPipelineProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"404 Project Not Found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	provider := &GitLabPipelineProvider{}
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

func TestMapGitLabStatus(t *testing.T) {
	tests := []struct {
		input string
		want  PipelineStatus
	}{
		{"running", StatusRunning},
		{"pending", StatusPending},
		{"success", StatusSuccess},
		{"failed", StatusFailed},
		{"canceled", StatusCanceled},
		{"canceled_for_retry", StatusFailed},
	}
	for _, tt := range tests {
		if got := mapGitLabStatus(tt.input); got != tt.want {
			t.Errorf("mapGitLabStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
