package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wii/grepom/config"
	"github.com/wii/grepom/repo"
)

func createVirtualGroupTestConfig(t *testing.T, dir string) string {
	t.Helper()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: frontend
    resource: gl
    path: my-org/frontend
    local_path: ./frontend
    repos:
      - name: web-app
        url: https://gitlab.com/my-org/frontend/web-app.git
        path: my-org/frontend/web-app
      - name: ui-lib
        url: https://gitlab.com/my-org/frontend/ui-lib.git
        path: my-org/frontend/ui-lib
  - name: backend
    resource: gl
    path: my-org/backend
    local_path: ./backend
    repos:
      - name: web-api
        url: https://gitlab.com/my-org/backend/web-api.git
        path: my-org/backend/web-api
virtual_groups:
  work:
    groups:
      - frontend
      - backend
`
	configPath := filepath.Join(dir, ".grepom.yml")
	os.WriteFile(configPath, []byte(content), 0644)
	return configPath
}

func captureStdout(fn func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := fn()
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestListGroups_MixedRealAndVirtual(t *testing.T) {
	dir := t.TempDir()
	configPath := createVirtualGroupTestConfig(t, dir)

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	output, err := captureStdout(func() error {
		return runListGroups(cfg)
	})
	if err != nil {
		t.Fatalf("runListGroups failed: %v", err)
	}

	normalized := strings.Join(strings.Fields(output), " ")
	if !strings.Contains(normalized, "TYPE NAME RESOURCE PATH LOCAL_PATH RECURSIVE REPOS GROUPS") {
		t.Errorf("expected new table header, got:\n%s", output)
	}
	if !strings.Contains(normalized, "group frontend gl my-org/frontend") {
		t.Errorf("expected real group row, got:\n%s", output)
	}
	if !strings.Contains(normalized, "vgroup work - - - - 3 frontend,backend") {
		t.Errorf("expected virtual group row with repo total 3, got:\n%s", output)
	}
}

func TestListRepos_VGroupFilter(t *testing.T) {
	dir := t.TempDir()
	configPath := createVirtualGroupTestConfig(t, dir)

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	config.ResolveBasePath(cfg, dir)

	listGroup = ""
	listVGroup = "work"
	listResource = ""
	listAll = true

	filter, err := buildRepoFilter(cfg, listGroup, listVGroup, listResource, listAll)
	if err != nil {
		t.Fatalf("buildRepoFilter failed: %v", err)
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		t.Fatalf("ResolveAndFilter failed: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos from virtual group work, got %d", len(repos))
	}

	_ = configPath
}

func TestStatus_VGroupFilter(t *testing.T) {
	dir := t.TempDir()
	configPath := createVirtualGroupTestConfig(t, dir)

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	config.ResolveBasePath(cfg, dir)

	filter, err := buildRepoFilter(cfg, "", "work", "", false)
	if err != nil {
		t.Fatalf("buildRepoFilter failed: %v", err)
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		t.Fatalf("ResolveAndFilter failed: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}

	names := make(map[string]bool)
	for _, r := range repos {
		names[r.Name] = true
	}
	for _, want := range []string{"web-app", "ui-lib", "web-api"} {
		if !names[want] {
			t.Errorf("expected repo %q in vgroup filter result", want)
		}
	}

	_ = configPath
}

func TestSearch_VGroupFilter(t *testing.T) {
	dir := t.TempDir()
	configPath := createVirtualGroupTestConfig(t, dir)

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	config.ResolveBasePath(cfg, dir)

	resolver := repo.NewResolver(cfg)
	allRepos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	filter, err := buildRepoFilter(cfg, "", "work", "", false)
	if err != nil {
		t.Fatalf("buildRepoFilter failed: %v", err)
	}
	results := repo.ApplySearchFilter(allRepos, "web", filter)
	if len(results) != 2 {
		t.Fatalf("expected 2 search hits in vgroup work, got %d", len(results))
	}

	_ = configPath
}

func TestSync_VGroupSelection(t *testing.T) {
	dir := t.TempDir()
	configPath := createVirtualGroupTestConfig(t, dir)

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	selection, err := cfg.ResolveGroupSelection("", "work")
	if err != nil {
		t.Fatalf("ResolveGroupSelection failed: %v", err)
	}
	groups, err := cfg.FilterGroups("", "work")
	if err != nil {
		t.Fatalf("FilterGroups failed: %v", err)
	}
	if len(selection) != 2 || len(groups) != 2 {
		t.Fatalf("expected 2 groups for sync vgroup work, got selection=%v groups=%d", selection, len(groups))
	}

	_ = configPath
}

func TestDedup_VGroupLimitsScope(t *testing.T) {
	dir := t.TempDir()
	configPath := createVirtualGroupTestConfig(t, dir)

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	groups, err := cfg.FilterGroups("", "work")
	if err != nil {
		t.Fatalf("FilterGroups failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups in vgroup work, got %d", len(groups))
	}

	selection, err := cfg.ResolveGroupSelection("", "work")
	if err != nil {
		t.Fatalf("ResolveGroupSelection failed: %v", err)
	}
	dups := detectCrossGroupDups(cfg.Groups, selection)
	_ = dups
	_ = configPath
}

func TestPrune_VGroupFilter(t *testing.T) {
	t.Cleanup(func() {
		pruneGroup = ""
		pruneVGroup = ""
		pruneResource = ""
		pruneApply = false
		pruneForce = false
	})
	dir := t.TempDir()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: frontend
    resource: gl
    path: my-org/frontend
    local_path: ./frontend
    exclude_repos:
      - old-app
    repos:
      - name: old-app
        url: https://gitlab.com/my-org/frontend/old-app.git
        path: my-org/frontend/old-app
  - name: backend
    resource: gl
    path: my-org/backend
    local_path: ./backend
    exclude_repos:
      - legacy-api
    repos:
      - name: legacy-api
        url: https://gitlab.com/my-org/backend/legacy-api.git
        path: my-org/backend/legacy-api
virtual_groups:
  work:
    groups:
      - frontend
      - backend
`
	configPath := filepath.Join(dir, "test.yml")
	os.WriteFile(configPath, []byte(content), 0644)

	configFile = configPath
	pruneApply = false
	pruneForce = false
	pruneGroup = ""
	pruneVGroup = "work"
	pruneResource = ""

	err := pruneCmd.RunE(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("prune --vgroup failed: %v", err)
	}
}
