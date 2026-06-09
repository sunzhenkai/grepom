package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func createPruneTestConfig(t *testing.T, dir string) string {
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
    exclude_repos:
      - old-app
      - deprecated-lib
    repos:
      - name: old-app
        url: https://gitlab.com/my-org/frontend/old-app.git
        path: my-org/frontend/old-app
      - name: deprecated-lib
        url: https://gitlab.com/my-org/frontend/deprecated-lib.git
        path: my-org/frontend/deprecated-lib
      - name: active-app
        url: https://gitlab.com/my-org/frontend/active-app.git
        path: my-org/frontend/active-app
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
      - name: new-api
        url: https://gitlab.com/my-org/backend/new-api.git
        path: my-org/backend/new-api
`
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func createFakeClonedRepo(t *testing.T, baseDir, groupLocalPath, repoSubpath string) string {
	t.Helper()
	repoDir := filepath.Join(baseDir, groupLocalPath, repoSubpath)
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	return repoDir
}

func TestPruneCommand_DryRun(t *testing.T) {
	dir := t.TempDir()
	configPath := createPruneTestConfig(t, dir)
	createFakeClonedRepo(t, dir, "frontend", "old-app")

	configFile = configPath
	pruneApply = false
	pruneForce = false
	pruneGroup = ""
	pruneVGroup = ""
	pruneResource = ""

	err := pruneCmd.RunE(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("prune dry-run failed: %v", err)
	}

	repoPath := filepath.Join(dir, "frontend", "old-app")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("dry-run should not delete repos")
	}
}

func TestPruneCommand_ApplyDeletesClean(t *testing.T) {
	dir := t.TempDir()
	configPath := createPruneTestConfig(t, dir)
	createFakeClonedRepo(t, dir, "frontend", "old-app")

	configFile = configPath
	pruneApply = true
	pruneForce = true
	pruneGroup = ""
	pruneVGroup = ""
	pruneResource = ""

	err := pruneCmd.RunE(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("prune --apply failed: %v", err)
	}

	repoPath := filepath.Join(dir, "frontend", "old-app")
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Error("repo should have been deleted with --apply --force")
	}
}

func TestPruneCommand_NoExcludedRepos(t *testing.T) {
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
    repos:
      - name: app
        url: https://gitlab.com/my-org/frontend/app.git
        path: my-org/frontend/app
`
	configPath := filepath.Join(dir, "test.yml")
	os.WriteFile(configPath, []byte(content), 0644)

	configFile = configPath
	pruneApply = false
	pruneForce = false
	pruneGroup = ""
	pruneVGroup = ""
	pruneResource = ""

	err := pruneCmd.RunE(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("prune should succeed: %v", err)
	}
}

func TestPruneCommand_GroupFilter(t *testing.T) {
	dir := t.TempDir()
	configPath := createPruneTestConfig(t, dir)
	createFakeClonedRepo(t, dir, "frontend", "old-app")
	createFakeClonedRepo(t, dir, "backend", "legacy-api")

	configFile = configPath
	pruneApply = true
	pruneForce = true
	pruneGroup = "frontend"
	pruneVGroup = ""
	pruneResource = ""

	err := pruneCmd.RunE(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("prune --group failed: %v", err)
	}

	frontendPath := filepath.Join(dir, "frontend", "old-app")
	if _, err := os.Stat(frontendPath); !os.IsNotExist(err) {
		t.Error("frontend excluded repo should have been deleted")
	}

	backendPath := filepath.Join(dir, "backend", "legacy-api")
	if _, err := os.Stat(backendPath); os.IsNotExist(err) {
		t.Error("backend excluded repo should NOT be deleted when filtering by frontend group")
	}
}

func TestPruneCommand_NotCloned(t *testing.T) {
	dir := t.TempDir()
	configPath := createPruneTestConfig(t, dir)

	configFile = configPath
	pruneApply = true
	pruneForce = false
	pruneGroup = ""
	pruneVGroup = ""
	pruneResource = ""

	err := pruneCmd.RunE(pruneCmd, []string{})
	if err != nil {
		t.Fatalf("prune should succeed with nothing cloned: %v", err)
	}
}
