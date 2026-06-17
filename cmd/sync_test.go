package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wii/grepom/config"
)

func TestSync_UserNamespaceUsesExistingFilterChain(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "sync.yml")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v4/groups/sunzhenkai":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"404 Group Not Found"}`))
		case r.URL.Path == "/api/v4/users":
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": 99, "username": "sunzhenkai", "name": "Sun"},
			})
		case r.URL.Path == "/api/v4/users/99/projects":
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"name": "grepom", "path_with_namespace": "sunzhenkai/grepom", "http_url_to_repo": "https://gitlab.example.com/sunzhenkai/grepom.git"},
				{"name": "private-repo", "path_with_namespace": "sunzhenkai/private-repo", "http_url_to_repo": "https://gitlab.example.com/sunzhenkai/private-repo.git"},
				{"name": "dup", "path_with_namespace": "sunzhenkai/grepom", "http_url_to_repo": "https://gitlab.example.com/sunzhenkai/grepom.git"},
				{"name": "other", "path_with_namespace": "other/repo", "http_url_to_repo": "https://gitlab.example.com/other/repo.git"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := strings.TrimPrefix(ts.URL, "http://")

	content := `
base: ` + dir + `
resources:
  fake:
    provider: gitlab
    url: http://` + host + `
    token: test-token
groups:
  - name: personal
    resource: fake
    path: sunzhenkai
    exclude_repos:
      - private-repo
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	configFile = configPath
	syncGroup = ""
	syncVGroup = ""
	syncResource = ""

	if _, err := captureStdout(func() error {
		return syncCmd.RunE(syncCmd, []string{})
	}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if len(cfg.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(cfg.Groups))
	}
	repos := cfg.Groups[0].Repos
	if len(repos) != 1 {
		t.Fatalf("expected only 1 repo after path/exclude/dedup filters, got %d", len(repos))
	}
	if repos[0].Path != "sunzhenkai/grepom" {
		t.Fatalf("expected user namespace repo, got %q", repos[0].Path)
	}
}

func TestSync_GroupPathBehaviorStillWorks(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "sync-group.yml")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v4/groups/my-org/team" || r.URL.Path == "/api/v4/groups/my-org%2Fteam":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 101, "path": "team", "full_path": "my-org/team",
			})
		case r.URL.Path == "/api/v4/groups/101/projects":
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"name": "web", "path_with_namespace": "my-org/team/web", "http_url_to_repo": "https://gitlab.example.com/my-org/team/web.git"},
				{"name": "api", "path_with_namespace": "my-org/team/sub/api", "http_url_to_repo": "https://gitlab.example.com/my-org/team/sub/api.git"},
			})
		case r.URL.Path == "/api/v4/groups/101/subgroups":
			json.NewEncoder(w).Encode([]map[string]interface{}{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := strings.TrimPrefix(ts.URL, "http://")

	content := `
base: ` + dir + `
resources:
  fake:
    provider: gitlab
    url: http://` + host + `
    token: test-token
groups:
  - name: org-group
    resource: fake
    path: my-org/team
    recursive: true
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	configFile = configPath
	syncGroup = ""
	syncVGroup = ""
	syncResource = ""

	if _, err := captureStdout(func() error {
		return syncCmd.RunE(syncCmd, []string{})
	}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if len(cfg.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(cfg.Groups))
	}
	if len(cfg.Groups[0].Repos) != 2 {
		t.Fatalf("expected 2 group repos, got %d", len(cfg.Groups[0].Repos))
	}
}
