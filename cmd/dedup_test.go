package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wii/grepom/config"
)

func createDedupTestConfig(t *testing.T, dir string) string {
	t.Helper()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: infra-team
    resource: gl
    path: my-org/infra
    local_path: ./infra
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/infra/api-lib.git
        path: my-org/infra/api-lib
      - name: worker
        url: https://gitlab.com/my-org/infra/worker.git
        path: my-org/infra/worker
      - name: deploy-tool
        url: https://gitlab.com/my-org/infra/deploy-tool.git
        path: my-org/infra/deploy-tool
  - name: core-team
    resource: gl
    path: my-org/core
    local_path: ./core
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
      - name: shared-utils
        url: https://gitlab.com/my-org/core/shared-utils.git
        path: my-org/core/shared-utils
      - name: worker
        url: https://gitlab.com/my-org/core/worker.git
        path: my-org/core/worker
      - name: core-only
        url: https://gitlab.com/my-org/core/core-only.git
        path: my-org/core/core-only
  - name: legacy-team
    resource: gl
    path: my-org/legacy
    local_path: ./legacy
    repos:
      - name: shared-utils
        url: https://gitlab.com/my-org/legacy/shared-utils.git
        path: my-org/legacy/shared-utils
      - name: old-app
        url: https://gitlab.com/my-org/legacy/old-app.git
        path: my-org/legacy/old-app
`
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestDedupCommand_DryRun(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "infra-team"
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup dry-run failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 4 {
		t.Errorf("dry-run should not modify config, expected 4 repos, got %d", len(coreGroup.Repos))
	}
}

func TestDedupCommand_ApplyWithReference(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "infra-team"
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup --apply failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 2 {
		t.Fatalf("expected 2 repos in core-team, got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	for _, r := range coreGroup.Repos {
		if r.Name == "api-lib" || r.Name == "worker" {
			t.Errorf("%s should have been removed from core-team repos", r.Name)
		}
	}

	if len(coreGroup.ExcludeRepos) != 2 {
		t.Fatalf("expected 2 exclude_repos, got %d: %v", len(coreGroup.ExcludeRepos), coreGroup.ExcludeRepos)
	}
	excluded := map[string]bool{}
	for _, e := range coreGroup.ExcludeRepos {
		excluded[e] = true
	}
	if !excluded["api-lib"] || !excluded["worker"] {
		t.Errorf("expected exclude_repos to contain api-lib and worker, got %v", coreGroup.ExcludeRepos)
	}

	infraGroup := cfg.Groups[0]
	if len(infraGroup.Repos) != 3 {
		t.Errorf("expected 3 repos in infra-team (unchanged), got %d", len(infraGroup.Repos))
	}
	if len(infraGroup.ExcludeRepos) != 0 {
		t.Errorf("expected no exclude_repos for infra-team, got %v", infraGroup.ExcludeRepos)
	}
}

func TestDedupCommand_NoReference_CompareAll(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = ""
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup --apply failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 1 {
		t.Fatalf("expected 1 repo remaining (core-only), got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	if coreGroup.Repos[0].Name != "core-only" {
		t.Errorf("expected core-only to remain, got %s", coreGroup.Repos[0].Name)
	}
	if len(coreGroup.ExcludeRepos) != 3 {
		t.Errorf("expected 3 exclude_repos, got %d: %v", len(coreGroup.ExcludeRepos), coreGroup.ExcludeRepos)
	}
}

func TestDedupCommand_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: infra-team
    resource: gl
    path: my-org/infra
    local_path: ./infra
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/infra/api-lib.git
        path: my-org/infra/api-lib
  - name: core-team
    resource: gl
    path: my-org/core
    local_path: ./core
    exclude_repos:
      - api-lib
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
`
	configPath := filepath.Join(dir, "test.yml")
	os.WriteFile(configPath, []byte(content), 0644)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "infra-team"
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(coreGroup.Repos))
	}
	if len(coreGroup.ExcludeRepos) != 1 || coreGroup.ExcludeRepos[0] != "api-lib" {
		t.Errorf("expected single exclude_repos [api-lib], got %v", coreGroup.ExcludeRepos)
	}
}

func TestDedupCommand_NoDuplicatesFound(t *testing.T) {
	dir := t.TempDir()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: infra-team
    resource: gl
    path: my-org/infra
    local_path: ./infra
    repos:
      - name: deploy-tool
        url: https://gitlab.com/my-org/infra/deploy-tool.git
        path: my-org/infra/deploy-tool
  - name: core-team
    resource: gl
    path: my-org/core
    local_path: ./core
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
`
	configPath := filepath.Join(dir, "test.yml")
	os.WriteFile(configPath, []byte(content), 0644)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "infra-team"
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 1 {
		t.Errorf("expected 1 repo (unchanged), got %d", len(coreGroup.Repos))
	}
}

func TestDedupCommand_GroupNotFound(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "nonexistent"
	dedupReference = ""
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestDedupCommand_ReferenceNotFound(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "nonexistent"
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err == nil {
		t.Fatal("expected error for nonexistent reference group")
	}
}

func TestDedupCommand_MultipleReferences(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "infra-team,legacy-team"
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup with multiple references failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 1 {
		t.Fatalf("expected 1 repo (core-only), got %d", len(coreGroup.Repos))
	}
	if coreGroup.Repos[0].Name != "core-only" {
		t.Errorf("expected core-only, got %s", coreGroup.Repos[0].Name)
	}
}
