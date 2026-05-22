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

	// 新设计：不指定 --reference 时不触发 Step 3 跨组 name 去重
	// Step 1 组内去重：无重复（各组内 URL 都不同）
	// Step 2 跨组 URL 警告：只打印不删除
	// 所以 core-team 的 repos 应保持不变
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	if len(coreGroup.Repos) != 4 {
		t.Fatalf("expected 4 repos (unchanged, no --reference), got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	if len(coreGroup.ExcludeRepos) != 0 {
		t.Errorf("expected no exclude_repos, got %v", coreGroup.ExcludeRepos)
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

func TestDedupCommand_NoGroupNoReference(t *testing.T) {
	// 不指定 --group 和 --reference 时，应对所有 group 执行 Step 1+2
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = ""
	dedupReference = ""
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup without --group should not error: %v", err)
	}

	// dry-run 模式不修改 config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	// repos 应保持不变
	totalRepos := 0
	for _, g := range cfg.Groups {
		totalRepos += len(g.Repos)
	}
	if totalRepos != 9 {
		t.Errorf("expected 9 repos total (unchanged), got %d", totalRepos)
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
		t.Fatalf("expected 1 repo (core-only), got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	if coreGroup.Repos[0].Name != "core-only" {
		t.Errorf("expected core-only, got %s", coreGroup.Repos[0].Name)
	}
}

// ═══════════════════════════════════════════
// 新增测试：组内 URL 去重
// ═══════════════════════════════════════════

func createIntraDupTestConfig(t *testing.T, dir string) string {
	t.Helper()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: core-team
    resource: gl
    path: my-org/core
    local_path: ./core
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
      - name: api-lib-dup
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
      - name: shared-utils
        url: https://gitlab.com/my-org/core/shared-utils.git
        path: my-org/core/shared-utils
      - name: worker-ssh
        url: git@gitlab.com:my-org/core/worker.git
        path: my-org/core/worker
      - name: worker-https
        url: https://gitlab.com/my-org/core/worker.git
        path: my-org/core/worker
  - name: infra-team
    resource: gl
    path: my-org/infra
    local_path: ./infra
    repos:
      - name: deploy-tool
        url: https://gitlab.com/my-org/infra/deploy-tool.git
        path: my-org/infra/deploy-tool
`
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestDedupCommand_IntraGroup_DryRun(t *testing.T) {
	dir := t.TempDir()
	configPath := createIntraDupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = ""
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup dry-run failed: %v", err)
	}

	// dry-run 不应修改 config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	coreGroup := cfg.Groups[0]
	if len(coreGroup.Repos) != 5 {
		t.Errorf("dry-run should not modify config, expected 5 repos, got %d", len(coreGroup.Repos))
	}
}

func TestDedupCommand_IntraGroup_Apply(t *testing.T) {
	dir := t.TempDir()
	configPath := createIntraDupTestConfig(t, dir)

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

	coreGroup := cfg.Groups[0]
	// api-lib-dup (与 api-lib 同 URL) 和 worker-https (与 worker-ssh 同 URL) 应被删除
	if len(coreGroup.Repos) != 3 {
		t.Fatalf("expected 3 repos after dedup, got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	// 保留第一个出现的条目
	keptNames := make(map[string]bool)
	for _, r := range coreGroup.Repos {
		keptNames[r.Name] = true
	}
	if !keptNames["api-lib"] {
		t.Error("expected api-lib to be kept (first occurrence)")
	}
	if keptNames["api-lib-dup"] {
		t.Error("expected api-lib-dup to be removed (duplicate)")
	}
	if !keptNames["shared-utils"] {
		t.Error("expected shared-utils to be kept")
	}
	if !keptNames["worker-ssh"] {
		t.Error("expected worker-ssh to be kept (first occurrence)")
	}
	if keptNames["worker-https"] {
		t.Error("expected worker-https to be removed (duplicate)")
	}

	// 组内去重不应添加 exclude_repos
	if len(coreGroup.ExcludeRepos) != 0 {
		t.Errorf("expected no exclude_repos for intra-group dedup, got %v", coreGroup.ExcludeRepos)
	}
}

func TestDedupCommand_IntraGroup_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir) // 默认配置中各组内无 URL 重复

	configFile = configPath
	dedupGroup = "infra-team"
	dedupReference = ""
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	infraGroup := cfg.Groups[0]
	if len(infraGroup.Repos) != 3 {
		t.Errorf("expected 3 repos (unchanged, no intra-group dupes), got %d", len(infraGroup.Repos))
	}
}

func TestDedupCommand_IntraGroup_ThreeDuplicates(t *testing.T) {
	dir := t.TempDir()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: core-team
    resource: gl
    path: my-org/core
    local_path: ./core
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
      - name: api-lib-v2
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
      - name: api-lib-v3
        url: git@gitlab.com:my-org/core/api-lib.git
        path: my-org/core/api-lib
`
	configPath := filepath.Join(dir, "test.yml")
	os.WriteFile(configPath, []byte(content), 0644)

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

	coreGroup := cfg.Groups[0]
	if len(coreGroup.Repos) != 1 {
		t.Fatalf("expected 1 repo after dedup (keep first), got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	if coreGroup.Repos[0].Name != "api-lib" {
		t.Errorf("expected api-lib (first) to remain, got %s", coreGroup.Repos[0].Name)
	}
}

// ═══════════════════════════════════════════
// 新增测试：跨组 URL 警告
// ═══════════════════════════════════════════

func createCrossDupTestConfig(t *testing.T, dir string) string {
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
        url: https://gitlab.com/my-org/api-lib.git
        path: my-org/infra/api-lib
      - name: deploy-tool
        url: https://gitlab.com/my-org/infra/deploy-tool.git
        path: my-org/infra/deploy-tool
  - name: core-team
    resource: gl
    path: my-org/core
    local_path: ./core
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/api-lib.git
        path: my-org/core/api-lib
      - name: shared-utils
        url: https://gitlab.com/my-org/core/shared-utils.git
        path: my-org/core/shared-utils
  - name: legacy-team
    resource: gl
    path: my-org/legacy
    local_path: ./legacy
    repos:
      - name: api-lib-ssh
        url: git@gitlab.com:my-org/api-lib.git
        path: my-org/legacy/api-lib
      - name: old-app
        url: https://gitlab.com/my-org/legacy/old-app.git
        path: my-org/legacy/old-app
`
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestDedupCommand_CrossGroup_WarningOnly(t *testing.T) {
	// 跨组 URL 重复只应输出警告，不修改配置
	dir := t.TempDir()
	configPath := createCrossDupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = ""
	dedupReference = ""
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup failed: %v", err)
	}

	// 即使 --apply，跨组警告也不应修改配置
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	totalRepos := 0
	for _, g := range cfg.Groups {
		totalRepos += len(g.Repos)
	}
	if totalRepos != 6 {
		t.Errorf("cross-group warnings should not modify repos, expected 6, got %d", totalRepos)
	}
}

func TestDedupCommand_CrossGroup_ExitCodeZero(t *testing.T) {
	dir := t.TempDir()
	configPath := createCrossDupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = ""
	dedupReference = ""
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	// 跨组 URL 重复不应导致错误退出码
	if err != nil {
		t.Errorf("cross-group warnings should not cause error, got: %v", err)
	}
}

func TestDedupCommand_CrossGroup_WithGroupFilter(t *testing.T) {
	// 指定 --group 时只报告该组与其他组之间的跨组重复
	dir := t.TempDir()
	configPath := createCrossDupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = ""
	dedupApply = false

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup failed: %v", err)
	}
	// 仅验证不报错，输出内容通过人工观察
}

// ═══════════════════════════════════════════
// 新增测试：三步流程集成
// ═══════════════════════════════════════════

func TestDedupCommand_AllSteps(t *testing.T) {
	// 同时指定 --group 和 --reference 时，三步都执行
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = "infra-team"
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup all steps failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	// Step 3 应从 core-team 中排除 api-lib 和 worker
	if len(coreGroup.Repos) != 2 {
		t.Fatalf("expected 2 repos after all steps, got %d: %+v", len(coreGroup.Repos), coreGroup.Repos)
	}
	for _, r := range coreGroup.Repos {
		if r.Name == "api-lib" || r.Name == "worker" {
			t.Errorf("%s should have been removed from core-team repos", r.Name)
		}
	}
	if len(coreGroup.ExcludeRepos) != 2 {
		t.Errorf("expected 2 exclude_repos, got %d: %v", len(coreGroup.ExcludeRepos), coreGroup.ExcludeRepos)
	}
}

func TestDedupCommand_NoGroupNoReference_AllGroups(t *testing.T) {
	// 不指定 --group 时，应对所有 group 执行 Step 1+2
	dir := t.TempDir()
	configPath := createIntraDupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = ""
	dedupReference = ""
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup without --group failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	// core-team 应有组内去重生效
	coreGroup := cfg.Groups[0]
	if len(coreGroup.Repos) != 3 {
		t.Errorf("expected 3 repos in core-team (after intra dedup), got %d", len(coreGroup.Repos))
	}

	// infra-team 无组内重复，应不变
	infraGroup := cfg.Groups[1]
	if len(infraGroup.Repos) != 1 {
		t.Errorf("expected 1 repo in infra-team, got %d", len(infraGroup.Repos))
	}
}

func TestDedupCommand_GroupOnly_Step1And2(t *testing.T) {
	// 仅指定 --group 时，执行 Step 1+2，不执行 Step 3
	dir := t.TempDir()
	configPath := createDedupTestConfig(t, dir)

	configFile = configPath
	dedupGroup = "core-team"
	dedupReference = ""
	dedupApply = true

	err := dedupCmd.RunE(dedupCmd, []string{})
	if err != nil {
		t.Fatalf("dedup --group only failed: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	coreGroup := cfg.Groups[1]
	// 没有 Step 3，跨组 name 去重不触发
	if len(coreGroup.Repos) != 4 {
		t.Errorf("expected 4 repos (no Step 3), got %d", len(coreGroup.Repos))
	}
	if len(coreGroup.ExcludeRepos) != 0 {
		t.Errorf("expected no exclude_repos (no Step 3), got %v", coreGroup.ExcludeRepos)
	}
}
