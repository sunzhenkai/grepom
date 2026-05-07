package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- Load tests ---

func TestLoad_ValidConfig(t *testing.T) {
	content := `
base: ~/projects
resources:
  work-gl:
    provider: gitlab
    url: https://gitlab.mycompany.com
    token: ${TEST_TOKEN}
groups:
  - name: frontend
    resource: work-gl
    path: my-org/frontend
    local_path: ./frontend
    recursive: true
    repos:
      - name: shared-utils
        url: https://gitlab.mycompany.com/my-org/frontend/shared-utils.git
        path: my-org/frontend/shared-utils
repos:
  - name: dotfiles
    resource: work-gl
    url: https://gitlab.mycompany.com/me/dotfiles.git
    local_path: ./dotfiles
`
	os.Setenv("TEST_TOKEN", "glpat-xxx")
	defer os.Unsetenv("TEST_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Base == "" {
		t.Error("base should not be empty")
	}
	if !filepath.IsAbs(cfg.Base) {
		t.Errorf("base should be absolute after tilde expansion, got: %s", cfg.Base)
	}
	if len(cfg.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(cfg.Resources))
	}
	res, ok := cfg.Resources["work-gl"]
	if !ok {
		t.Fatal("expected resource 'work-gl'")
	}
	if res.Token != "${TEST_TOKEN}" {
		t.Errorf("token should remain as placeholder (lazy resolution), got: %s", res.Token)
	}
	if res.Provider != "gitlab" {
		t.Errorf("expected provider gitlab, got: %s", res.Provider)
	}
	if len(cfg.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Name != "frontend" {
		t.Errorf("expected group name 'frontend', got: %s", cfg.Groups[0].Name)
	}
	if cfg.Groups[0].Recursive != true {
		t.Error("expected recursive=true")
	}
	if len(cfg.Groups[0].Repos) != 1 {
		t.Fatalf("expected 1 repo in group, got %d", len(cfg.Groups[0].Repos))
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 standalone repo, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0].Name != "dotfiles" {
		t.Errorf("expected standalone repo 'dotfiles', got: %s", cfg.Repos[0].Name)
	}
}

func TestLoad_MissingBase(t *testing.T) {
	content := `
resources:
  work-gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing base")
	}
}

func TestLoad_UnsupportedProvider(t *testing.T) {
	content := `
base: ~/projects
resources:
  my-res:
    provider: bitbucket
    url: https://bitbucket.org
    token: test
groups:
  - name: test
    resource: my-res
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestLoad_CodeupResource(t *testing.T) {
	content := `
base: ~/projects
resources:
  my-codeup:
    provider: codeup
    url: codeup.aliyun.com
    token: test-token
    organization_id: "60de7a6852743a5162b5f957"
groups:
  - name: solo
    resource: my-codeup
    path: wii/solo
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	res, ok := cfg.Resources["my-codeup"]
	if !ok {
		t.Fatal("resource my-codeup not found")
	}
	if res.Provider != "codeup" {
		t.Errorf("expected provider codeup, got %s", res.Provider)
	}
	if res.OrganizationID != "60de7a6852743a5162b5f957" {
		t.Errorf("expected organization_id '60de7a6852743a5162b5f957', got %s", res.OrganizationID)
	}
	if res.URL != "codeup.aliyun.com" {
		t.Errorf("expected url 'codeup.aliyun.com', got %s", res.URL)
	}
}

func TestLoad_CodeupMissingOrganizationID(t *testing.T) {
	content := `
base: ~/projects
resources:
  my-codeup:
    provider: codeup
    url: codeup.aliyun.com
    token: test-token
groups:
  - name: solo
    resource: my-codeup
    path: wii/solo
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for codeup without organization_id")
	}
	if !strings.Contains(err.Error(), "organization_id") {
		t.Errorf("expected organization_id error, got: %v", err)
	}
}

func TestLoad_UndefinedEnvVar(t *testing.T) {
	content := `
base: ~/projects
resources:
  my-res:
    provider: gitlab
    url: https://gitlab.com
    token: ${UNDEFINED_VAR}
groups:
  - name: test
    resource: my-res
    path: my-org
`
	os.Unsetenv("UNDEFINED_VAR")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed with unresolved env var (deferred resolution), got: %v", err)
	}
	// Token should remain as placeholder string
	if cfg.Resources["my-res"].Token != "${UNDEFINED_VAR}" {
		t.Errorf("expected token to remain as placeholder '${UNDEFINED_VAR}', got: %s", cfg.Resources["my-res"].Token)
	}
}

func TestLoad_MultipleResourcesPartialEnvVar(t *testing.T) {
	content := `
base: ~/projects
resources:
  github:
    provider: github
    url: github.com
    token: ${GITHUB_TEST_TOKEN}
  gitlab:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_UNDEF_TOKEN}
groups:
  - name: fe
    resource: github
    path: my-org/frontend
`
	os.Setenv("GITHUB_TEST_TOKEN", "ghp-xxx")
	defer os.Unsetenv("GITHUB_TEST_TOKEN")
	os.Unsetenv("GITLAB_UNDEF_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed even with partial env vars, got: %v", err)
	}
	// Both tokens remain as placeholders
	if cfg.Resources["github"].Token != "${GITHUB_TEST_TOKEN}" {
		t.Errorf("github token should remain as placeholder, got: %s", cfg.Resources["github"].Token)
	}
	if cfg.Resources["gitlab"].Token != "${GITLAB_UNDEF_TOKEN}" {
		t.Errorf("gitlab token should remain as placeholder, got: %s", cfg.Resources["gitlab"].Token)
	}
}

func TestLoad_DisabledResourceNoTokenResolve(t *testing.T) {
	content := `
base: ~/projects
resources:
  disabled-gl:
    provider: gitlab
    url: https://gitlab.com
    token: ${NEVER_SET_TOKEN}
    enabled: false
  github:
    provider: github
    url: github.com
    token: ${GITHUB_TOKEN_SET}
groups:
  - name: fe
    resource: github
    path: my-org/frontend
`
	os.Setenv("GITHUB_TOKEN_SET", "ghp-xxx")
	defer os.Unsetenv("GITHUB_TOKEN_SET")
	os.Unsetenv("NEVER_SET_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed with disabled resource, got: %v", err)
	}
	if cfg.Resources["disabled-gl"].Token != "${NEVER_SET_TOKEN}" {
		t.Errorf("disabled resource token should remain as placeholder, got: %s", cfg.Resources["disabled-gl"].Token)
	}
}

func TestLoad_GroupReferencesMissingResource(t *testing.T) {
	content := `
base: ~/projects
resources:
  work-gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: test
    resource: nonexistent
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing resource reference")
	}
}

func TestLoad_DuplicateGroupNames(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: frontend
    resource: gl
    path: org1/frontend
  - name: frontend
    resource: gl
    path: org2/frontend
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate group names")
	}
	if !strings.Contains(err.Error(), "duplicate group name") {
		t.Errorf("error should mention duplicate, got: %v", err)
	}
}

func TestLoad_GroupRepoPathMismatch(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: frontend
    resource: gl
    path: my-org/frontend
    repos:
      - name: wrong
        url: https://gitlab.com/other/repo.git
        path: other-org/backend
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for repo path not matching group path")
	}
}

func TestLoad_DefaultLocalPaths(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: frontend
    resource: gl
    path: my-org/frontend
repos:
  - name: dotfiles
    resource: gl
    url: https://gitlab.com/me/dotfiles.git
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Groups[0].LocalPath != "./frontend" {
		t.Errorf("expected default local_path './frontend', got: %s", cfg.Groups[0].LocalPath)
	}
	if cfg.Repos[0].LocalPath != "./dotfiles" {
		t.Errorf("expected default local_path './dotfiles', got: %s", cfg.Repos[0].LocalPath)
	}
}

// --- Token placeholder tests ---

func TestResolveToken_PlainText(t *testing.T) {
	result, err := ResolveToken("glpat-xxxxxxxxxxxx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "glpat-xxxxxxxxxxxx" {
		t.Errorf("expected plain text token unchanged, got: %s", result)
	}
}

func TestResolveToken_EnvVar(t *testing.T) {
	os.Setenv("MY_TOKEN", "secret-value")
	defer os.Unsetenv("MY_TOKEN")

	result, err := ResolveToken("${MY_TOKEN}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "secret-value" {
		t.Errorf("expected resolved value, got: %s", result)
	}
}

func TestResolveToken_Empty(t *testing.T) {
	result, err := ResolveToken("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got: %s", result)
	}
}

func TestResolveToken_UndefinedEnvVar(t *testing.T) {
	os.Unsetenv("DEFINITELY_NOT_SET_TOKEN")

	_, err := ResolveToken("${DEFINITELY_NOT_SET_TOKEN}")
	if err == nil {
		t.Fatal("expected error for undefined env var")
	}
}

func TestResolveToken_SingleQuotedEnvVar(t *testing.T) {
	os.Setenv("MY_TOKEN", "secret-value")
	defer os.Unsetenv("MY_TOKEN")

	result, err := ResolveToken("'${MY_TOKEN}'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "secret-value" {
		t.Errorf("expected resolved value from single-quoted placeholder, got: %s", result)
	}
}

func TestResolveToken_DoubleQuotedEnvVar(t *testing.T) {
	os.Setenv("MY_TOKEN", "secret-value")
	defer os.Unsetenv("MY_TOKEN")

	result, err := ResolveToken("\"${MY_TOKEN}\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "secret-value" {
		t.Errorf("expected resolved value from double-quoted placeholder, got: %s", result)
	}
}

func TestResolveToken_QuotedPlainText(t *testing.T) {
	result, err := ResolveToken("\"glpat-xxx\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "glpat-xxx" {
		t.Errorf("expected quotes stripped from plain text, got: %s", result)
	}
}

func TestResolveToken_MismatchedQuotes(t *testing.T) {
	result, err := ResolveToken("'${MY_TOKEN}\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mismatched quotes should not be stripped, so it's not a valid placeholder
	if result != "'${MY_TOKEN}\"" {
		t.Errorf("expected mismatched quotes preserved, got: %s", result)
	}
}

// --- stripQuotes tests ---

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"'${GITLAB_TOKEN}'", "${GITLAB_TOKEN}"},
		{"\"${GITLAB_TOKEN}\"", "${GITLAB_TOKEN}"},
		{"'hello'", "hello"},
		{"\"world\"", "world"},
		{"${GITLAB_TOKEN}", "${GITLAB_TOKEN}"},
		{"glpat-xxx", "glpat-xxx"},
		{"", ""},
		{"'", "'"},
		{"\"\"", ""},
		{"''", ""},
		{"'mixed\"", "'mixed\""},
		{"\"mixed'", "\"mixed'"},
		{"a", "a"},
		{"ab", "ab"},
		{"'a", "'a"},
		{"a'", "a'"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- ResolvedToken tests ---

func TestResolvedToken_EnvVar(t *testing.T) {
	os.Setenv("RES_RESOLVE_TEST_TOKEN", "resolved-value")
	defer os.Unsetenv("RES_RESOLVE_TEST_TOKEN")

	res := Resource{Token: "${RES_RESOLVE_TEST_TOKEN}"}
	token, err := res.ResolvedToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "resolved-value" {
		t.Errorf("expected 'resolved-value', got: %s", token)
	}
}

func TestResolvedToken_PlainText(t *testing.T) {
	res := Resource{Token: "glpat-xxxxxxxxxxxx"}
	token, err := res.ResolvedToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "glpat-xxxxxxxxxxxx" {
		t.Errorf("expected 'glpat-xxxxxxxxxxxx', got: %s", token)
	}
}

func TestResolvedToken_UndefinedEnvVar(t *testing.T) {
	os.Unsetenv("RES_UNDEF_TOKEN")

	res := Resource{Token: "${RES_UNDEF_TOKEN}"}
	_, err := res.ResolvedToken()
	if err == nil {
		t.Fatal("expected error for undefined env var")
	}
}

func TestResolvedToken_Empty(t *testing.T) {
	res := Resource{Token: ""}
	token, err := res.ResolvedToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty string, got: %s", token)
	}
}

func TestResolvedToken_QuotedEnvVar(t *testing.T) {
	os.Setenv("RES_QUOTED_TOKEN", "quoted-value")
	defer os.Unsetenv("RES_QUOTED_TOKEN")

	res := Resource{Token: "'${RES_QUOTED_TOKEN}'"}
	token, err := res.ResolvedToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "quoted-value" {
		t.Errorf("expected 'quoted-value', got: %s", token)
	}
}

// --- Write config tests ---

func TestWriteConfig_PreservesTokenPlaceholder(t *testing.T) {
	content := `
base: ~/projects
resources:
  work-gl:
    provider: gitlab
    url: https://gitlab.com
    token: ${TEST_WRITE_TOKEN}
groups:
  - name: test
    resource: work-gl
    path: my-org
`
	os.Setenv("TEST_WRITE_TOKEN", "glpat-write-test")
	defer os.Unsetenv("TEST_WRITE_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Token remains as placeholder (lazy resolution)
	res := cfg.Resources["work-gl"]
	if res.Token != "${TEST_WRITE_TOKEN}" {
		t.Fatalf("expected token to remain as placeholder, got: %s", res.Token)
	}

	// Add a repo and write back
	cfg.Repos = append(cfg.Repos, Repo{Name: "test-repo", Resource: "work-gl", URL: "https://gitlab.com/test.git", LocalPath: "./test-repo"})
	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	rawContent := string(data)
	if !strings.Contains(rawContent, "${TEST_WRITE_TOKEN}") {
		t.Errorf("token placeholder should be preserved in file, got:\n%s", rawContent)
	}
	if strings.Contains(rawContent, "glpat-write-test") {
		t.Errorf("resolved token should NOT appear in file, got:\n%s", rawContent)
	}
}

// --- findConfigUpward tests ---

func TestFindConfigUpward_CurrentDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".grepom.yml")
	os.WriteFile(configPath, []byte("base: ~/projects"), 0644)

	found, err := findConfigUpward(dir)
	if err != nil {
		t.Fatalf("findConfigUpward failed: %v", err)
	}
	if found != configPath {
		t.Errorf("expected %s, got %s", configPath, found)
	}
}

func TestFindConfigUpward_ParentDir(t *testing.T) {
	parent := t.TempDir()
	configPath := filepath.Join(parent, ".grepom.yml")
	os.WriteFile(configPath, []byte("base: ~/projects"), 0644)

	child := filepath.Join(parent, "subdir")
	os.MkdirAll(child, 0755)

	found, err := findConfigUpward(child)
	if err != nil {
		t.Fatalf("findConfigUpward failed: %v", err)
	}
	if found != configPath {
		t.Errorf("expected %s, got %s", configPath, found)
	}
}

func TestFindConfigUpward_MultiLevel(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".grepom.yml")
	os.WriteFile(configPath, []byte("base: ~/projects"), 0644)

	deep := filepath.Join(root, "a", "b", "c", "d")
	os.MkdirAll(deep, 0755)

	found, err := findConfigUpward(deep)
	if err != nil {
		t.Fatalf("findConfigUpward failed: %v", err)
	}
	if found != configPath {
		t.Errorf("expected %s, got %s", configPath, found)
	}
}

func TestFindConfigUpward_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := findConfigUpward(dir)
	if err == nil {
		t.Fatal("expected error when no config found upward")
	}
	if !IsConfigNotFound(err) {
		t.Errorf("expected ErrConfigNotFound, got: %v", err)
	}
}

func TestFindConfigUpward_NestedConfig_NearestWins(t *testing.T) {
	parent := t.TempDir()
	parentConfig := filepath.Join(parent, ".grepom.yml")
	os.WriteFile(parentConfig, []byte("base: ~/projects"), 0644)

	child := filepath.Join(parent, "personal")
	os.MkdirAll(child, 0755)
	childConfig := filepath.Join(child, ".grepom.yml")
	os.WriteFile(childConfig, []byte("base: ~/personal"), 0644)

	grandchild := filepath.Join(child, "blog")
	os.MkdirAll(grandchild, 0755)

	// 从 grandchild 向上查找，应该找到最近的 child 中的配置
	found, err := findConfigUpward(grandchild)
	if err != nil {
		t.Fatalf("findConfigUpward failed: %v", err)
	}
	if found != childConfig {
		t.Errorf("expected nearest config %s, got %s", childConfig, found)
	}
}

// --- FindConfig integration tests (upward search) ---

func TestFindConfig_UpwardFromSubdirectory(t *testing.T) {
	parent := t.TempDir()
	configPath := filepath.Join(parent, ".grepom.yml")
	os.WriteFile(configPath, []byte("base: ~/projects"), 0644)

	child := filepath.Join(parent, "my-org", "web-app")
	os.MkdirAll(child, 0755)

	oldDir, _ := os.Getwd()
	os.Chdir(child)
	defer os.Chdir(oldDir)

	found, err := FindConfig("")
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}
	// 应该找到父目录的配置文件（绝对路径）
	if filepath.Clean(found) != configPath {
		t.Errorf("expected %s, got %s", configPath, filepath.Clean(found))
	}
}

func TestFindConfig_CurrentDirTakesPrecedence(t *testing.T) {
	parent := t.TempDir()
	parentConfig := filepath.Join(parent, ".grepom.yml")
	os.WriteFile(parentConfig, []byte("base: ~/parent"), 0644)

	child := filepath.Join(parent, "child")
	os.MkdirAll(child, 0755)
	childConfig := filepath.Join(child, ".grepom.yml")
	os.WriteFile(childConfig, []byte("base: ~/child"), 0644)

	oldDir, _ := os.Getwd()
	os.Chdir(child)
	defer os.Chdir(oldDir)

	found, err := FindConfig("")
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}
	if found != ".grepom.yml" {
		t.Errorf("expected .grepom.yml (current dir), got %s", found)
	}
}

// --- FindConfig tests ---

func TestFindConfig_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yml")
	os.WriteFile(path, []byte("base: ~/projects"), 0644)

	found, err := FindConfig(path)
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}
	if found != path {
		t.Errorf("expected %s, got %s", path, found)
	}
}

func TestFindConfig_DefaultDotGrepom(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	os.WriteFile(".grepom.yml", []byte("base: ~/projects"), 0644)

	found, err := FindConfig("")
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}
	if found != ".grepom.yml" {
		t.Errorf("expected .grepom.yml, got %s", found)
	}
}

func TestFindConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	_, err := FindConfig("")
	if err == nil {
		t.Fatal("expected error when no config found")
	}
}

func TestFindConfig_NotFound_ReturnsErrConfigNotFound(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	_, err := FindConfig("")
	if err == nil {
		t.Fatal("expected error when no config found")
	}
	if !IsConfigNotFound(err) {
		t.Errorf("expected IsConfigNotFound to be true, got error: %v", err)
	}
}

func TestIsConfigNotFound_Nil(t *testing.T) {
	if IsConfigNotFound(nil) {
		t.Error("expected IsConfigNotFound(nil) to be false")
	}
}

func TestIsConfigNotFound_OtherError(t *testing.T) {
	if IsConfigNotFound(fmt.Errorf("some other error")) {
		t.Error("expected IsConfigNotFound to be false for non-config error")
	}
}

func TestFindConfig_ExplicitPathNotFound_ReturnsErrConfigNotFound(t *testing.T) {
	_, err := FindConfig("/nonexistent/path/config.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent explicit path")
	}
	if !IsConfigNotFound(err) {
		t.Errorf("expected IsConfigNotFound to be true for explicit path not found, got: %v", err)
	}
}

// --- Path resolution tests ---

func TestResolveGroupRepoPath_DirectChild(t *testing.T) {
	result := ResolveGroupRepoPath("/home/user/projects", "./frontend", "my-org/frontend", "my-org/frontend/shared-utils")
	expected := filepath.Join("/home/user/projects", "frontend", "shared-utils")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveGroupRepoPath_NestedSubgroup(t *testing.T) {
	result := ResolveGroupRepoPath("/home/user/projects", "./frontend", "my-org/frontend", "my-org/frontend/ui/design-system")
	expected := filepath.Join("/home/user/projects", "frontend", "ui", "design-system")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveGroupRepoPath_ThreeLevelsDeep(t *testing.T) {
	result := ResolveGroupRepoPath("/home/user/projects", "./frontend", "my-org/frontend", "my-org/frontend/ui/components/button-lib")
	expected := filepath.Join("/home/user/projects", "frontend", "ui", "components", "button-lib")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveRepoPath(t *testing.T) {
	result := ResolveRepoPath("/home/user/projects", "./dotfiles")
	expected := filepath.Join("/home/user/projects", "dotfiles")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// --- ResolveBasePath tests ---

func TestResolveBasePath_RelativePath(t *testing.T) {
	cfg := &Config{Base: "repos/my-org"}
	ResolveBasePath(cfg, "/home/user/projects")
	expected := filepath.Join("/home/user/projects", "repos/my-org")
	if cfg.Base != expected {
		t.Errorf("expected %s, got %s", expected, cfg.Base)
	}
}

func TestResolveBasePath_DotSlashPrefix(t *testing.T) {
	cfg := &Config{Base: "./repos"}
	ResolveBasePath(cfg, "/home/user/projects")
	expected := filepath.Join("/home/user/projects", "repos")
	if cfg.Base != expected {
		t.Errorf("expected %s, got %s", expected, cfg.Base)
	}
}

func TestResolveBasePath_AbsolutePath_Unchanged(t *testing.T) {
	cfg := &Config{Base: "/opt/repos"}
	ResolveBasePath(cfg, "/home/user/projects")
	if cfg.Base != "/opt/repos" {
		t.Errorf("expected /opt/repos, got %s", cfg.Base)
	}
}

func TestResolveBasePath_TildePath_Unchanged(t *testing.T) {
	// expandTilde 已经在 Load 中处理了 ~/，这里测试已是绝对路径后的情况
	home, _ := os.UserHomeDir()
	expanded := filepath.Join(home, "projects")
	cfg := &Config{Base: expanded}
	ResolveBasePath(cfg, "/some/other/dir")
	if cfg.Base != expanded {
		t.Errorf("expected %s, got %s", expanded, cfg.Base)
	}
}

func TestResolveBasePath_EmptyBase_NoPanic(t *testing.T) {
	cfg := &Config{Base: ""}
	ResolveBasePath(cfg, "/home/user/projects")
	if cfg.Base != "" {
		t.Errorf("expected empty, got %s", cfg.Base)
	}
}

// --- Path conflict detection tests ---

func TestDetectPathConflicts_NoConflict(t *testing.T) {
	cfg := &Config{
		Base: "/home/user/projects",
		Resources: map[string]Resource{
			"gl": {Provider: "gitlab", URL: "https://gitlab.com", Token: "test"},
		},
		Groups: []Group{
			{
				Name: "frontend", Resource: "gl", Path: "my-org/frontend", LocalPath: "./frontend",
				Repos: []GroupRepo{
					{Name: "app", Path: "my-org/frontend/app", URL: "https://gitlab.com/my-org/frontend/app.git"},
				},
			},
		},
		Repos: []Repo{
			{Name: "dotfiles", Resource: "gl", URL: "https://gitlab.com/me/dotfiles.git", LocalPath: "./dotfiles"},
		},
	}

	if err := cfg.DetectPathConflicts(); err != nil {
		t.Errorf("expected no conflict, got: %v", err)
	}
}

func TestDetectPathConflicts_DuplicatePaths(t *testing.T) {
	cfg := &Config{
		Base: "/home/user/projects",
		Resources: map[string]Resource{
			"gl": {Provider: "gitlab", URL: "https://gitlab.com", Token: "test"},
		},
		Groups: []Group{
			{
				Name: "frontend", Resource: "gl", Path: "my-org/frontend", LocalPath: "./frontend",
				Repos: []GroupRepo{
					{Name: "app", Path: "my-org/frontend/app", URL: "https://gitlab.com/my-org/frontend/app.git"},
				},
			},
		},
		Repos: []Repo{
			{Name: "app", Resource: "gl", URL: "https://gitlab.com/me/app.git", LocalPath: "./frontend/app"},
		},
	}

	err := cfg.DetectPathConflicts()
	if err == nil {
		t.Fatal("expected path conflict error")
	}
	if !strings.Contains(err.Error(), "path conflict") {
		t.Errorf("expected path conflict message, got: %v", err)
	}
}

// --- Add operations tests ---

func TestInitConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	if err := InitConfig(path, ""); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after InitConfig failed: %v", err)
	}
	if cfg.Base == "" {
		t.Error("base should not be empty")
	}
	if cfg.Resources == nil {
		t.Error("resources should be initialized")
	}
}

func TestInitConfig_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte("base: ~/projects"), 0644)

	err := InitConfig(path, "")
	if err == nil {
		t.Fatal("expected error when config already exists")
	}
}

func TestAddResource(t *testing.T) {
	os.Setenv("GL_TOKEN", "test-token-value")
	defer os.Unsetenv("GL_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	res := Resource{
		Provider: "gitlab",
		URL:      "https://gitlab.com",
		Token:    "${GL_TOKEN}",
	}

	if err := AddResource(path, "work-gl", res); err != nil {
		t.Fatalf("AddResource failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after AddResource failed: %v", err)
	}
	if len(cfg.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(cfg.Resources))
	}
	if cfg.Resources["work-gl"].Provider != "gitlab" {
		t.Error("provider mismatch")
	}

	// Check file contains placeholder
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "${GL_TOKEN}") {
		t.Error("token placeholder should be preserved in file")
	}
}

func TestAddGroup(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	group := Group{
		Name:     "frontend",
		Resource: "gl",
		Path:     "my-org/frontend",
	}

	if err := AddGroup(path, group); err != nil {
		t.Fatalf("AddGroup failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after AddGroup failed: %v", err)
	}
	if len(cfg.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Name != "frontend" {
		t.Error("group name mismatch")
	}
}

func TestSyncGroupRepos_AddNewRepos(t *testing.T) {
	content := `
base: ~/projects
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
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	newRepos := []GroupRepo{
		{Name: "repo1", URL: "https://gitlab.com/my-org/frontend/repo1.git", Path: "my-org/frontend/repo1"},
		{Name: "repo2", URL: "https://gitlab.com/my-org/frontend/repo2.git", Path: "my-org/frontend/repo2"},
	}

	added, err := SyncGroupRepos(path, "frontend", newRepos)
	if err != nil {
		t.Fatalf("SyncGroupRepos failed: %v", err)
	}
	if added != 2 {
		t.Errorf("expected 2 repos added, got %d", added)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after SyncGroupRepos failed: %v", err)
	}
	if len(cfg.Groups[0].Repos) != 2 {
		t.Fatalf("expected 2 repos in group, got %d", len(cfg.Groups[0].Repos))
	}
}

func TestSyncGroupRepos_NoDuplicates(t *testing.T) {
	content := `
base: ~/projects
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
      - name: repo1
        url: https://gitlab.com/my-org/frontend/repo1.git
        path: my-org/frontend/repo1
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	newRepos := []GroupRepo{
		{Name: "repo1", URL: "https://gitlab.com/my-org/frontend/repo1.git", Path: "my-org/frontend/repo1"},
		{Name: "repo2", URL: "https://gitlab.com/my-org/frontend/repo2.git", Path: "my-org/frontend/repo2"},
	}

	added, err := SyncGroupRepos(path, "frontend", newRepos)
	if err != nil {
		t.Fatalf("SyncGroupRepos failed: %v", err)
	}
	if added != 1 {
		t.Errorf("expected 1 repo added (dedup by URL), got %d", added)
	}
}

// --- File lock test ---

func TestWithFileLock_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte("base: ~/projects"), 0644)

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = WithFileLock(path, 5*time.Second, func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			})
		}(i)
	}

	wg.Wait()

	for i, err := range results {
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}
}

// --- FindGroup / FindResource tests ---

func TestFindGroup(t *testing.T) {
	cfg := &Config{
		Groups: []Group{
			{Name: "frontend", Resource: "gl", Path: "org/frontend"},
			{Name: "backend", Resource: "gl", Path: "org/backend"},
		},
	}

	idx, group, err := cfg.FindGroup("backend")
	if err != nil {
		t.Fatalf("FindGroup failed: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if group.Name != "backend" {
		t.Errorf("expected backend, got: %s", group.Name)
	}
}

func TestFindGroup_NotFound(t *testing.T) {
	cfg := &Config{
		Groups: []Group{
			{Name: "frontend", Resource: "gl", Path: "org/frontend"},
		},
	}

	_, _, err := cfg.FindGroup("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestFindResource(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"work-gl": {Provider: "gitlab", URL: "https://gitlab.com"},
		},
	}

	res, err := cfg.FindResource("work-gl")
	if err != nil {
		t.Fatalf("FindResource failed: %v", err)
	}
	if res.Provider != "gitlab" {
		t.Errorf("expected gitlab, got: %s", res.Provider)
	}
}

func TestFindResource_NotFound(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{},
	}

	_, err := cfg.FindResource("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent resource")
	}
}

// --- URL parse scheme tests ---

func TestParseResourceURL_HTTPS(t *testing.T) {
	host, scheme := parseResourceURL("https://gitlab.mycompany.com")
	if host != "gitlab.mycompany.com" || scheme != "https" {
		t.Errorf("expected (gitlab.mycompany.com, https), got (%s, %s)", host, scheme)
	}
}

func TestParseResourceURL_HTTP(t *testing.T) {
	host, scheme := parseResourceURL("http://gitlab.mycompany.com:8080")
	if host != "gitlab.mycompany.com:8080" || scheme != "http" {
		t.Errorf("expected (gitlab.mycompany.com:8080, http), got (%s, %s)", host, scheme)
	}
}

func TestParseResourceURL_NoScheme(t *testing.T) {
	host, scheme := parseResourceURL("gitlab.mycompany.com")
	if host != "gitlab.mycompany.com" || scheme != "" {
		t.Errorf("expected (gitlab.mycompany.com, ), got (%s, %s)", host, scheme)
	}
}

func TestParseResourceURL_Empty(t *testing.T) {
	host, scheme := parseResourceURL("")
	if host != "" || scheme != "" {
		t.Errorf("expected (, ), got (%s, %s)", host, scheme)
	}
}

func TestParseResourceURL_NoSchemeWithPort(t *testing.T) {
	host, scheme := parseResourceURL("gitlab.mycompany.com:8080")
	if host != "gitlab.mycompany.com:8080" || scheme != "" {
		t.Errorf("expected (gitlab.mycompany.com:8080, ), got (%s, %s)", host, scheme)
	}
}

// --- Resource URL derivation tests ---

func TestResource_APIURL_Auto(t *testing.T) {
	r := Resource{URL: "gitlab.mycompany.com"}
	if got := r.APIURL(); got != "https://gitlab.mycompany.com" {
		t.Errorf("expected https://gitlab.mycompany.com, got: %s", got)
	}
}

func TestResource_APIURL_AutoWithPort(t *testing.T) {
	r := Resource{URL: "gitlab.mycompany.com:8443"}
	if got := r.APIURL(); got != "https://gitlab.mycompany.com:8443" {
		t.Errorf("expected https://gitlab.mycompany.com:8443, got: %s", got)
	}
}

func TestResource_APIURL_HTTPScheme(t *testing.T) {
	r := Resource{URL: "gitlab.mycompany.com", scheme: "http"}
	if got := r.APIURL(); got != "http://gitlab.mycompany.com" {
		t.Errorf("expected http://gitlab.mycompany.com, got: %s", got)
	}
}

func TestResource_APIURL_HTTPSScheme(t *testing.T) {
	r := Resource{URL: "gitlab.mycompany.com", scheme: "https"}
	if got := r.APIURL(); got != "https://gitlab.mycompany.com" {
		t.Errorf("expected https://gitlab.mycompany.com, got: %s", got)
	}
}

func TestResource_Scheme(t *testing.T) {
	tests := []struct {
		name   string
		res    Resource
		expect string
	}{
		{"auto (empty)", Resource{URL: "g.wii.pub"}, ""},
		{"http", Resource{URL: "g.wii.pub", scheme: "http"}, "http"},
		{"https", Resource{URL: "g.wii.pub", scheme: "https"}, "https"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.Scheme(); got != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, got)
			}
		})
	}
}

func TestResource_SSHURL(t *testing.T) {
	r := Resource{URL: "g.wii.pub"}
	if got := r.SSHURL("my-org/my-repo"); got != "git@g.wii.pub:my-org/my-repo.git" {
		t.Errorf("expected git@g.wii.pub:my-org/my-repo.git, got: %s", got)
	}
}

func TestResource_SSHURL_WithPort(t *testing.T) {
	r := Resource{URL: "g.wii.pub:8022"}
	if got := r.SSHURL("my-org/my-repo"); got != "git@g.wii.pub:8022:my-org/my-repo.git" {
		t.Errorf("expected git@g.wii.pub:8022:my-org/my-repo.git, got: %s", got)
	}
}

func TestResource_HTTPSURL(t *testing.T) {
	r := Resource{URL: "g.wii.pub"}
	if got := r.HTTPSURL("my-org/my-repo"); got != "https://g.wii.pub/my-org/my-repo.git" {
		t.Errorf("expected https://g.wii.pub/my-org/my-repo.git, got: %s", got)
	}
}

func TestResource_HTTPSURL_WithPort(t *testing.T) {
	r := Resource{URL: "g.wii.pub:8443"}
	if got := r.HTTPSURL("my-org/my-repo"); got != "https://g.wii.pub:8443/my-org/my-repo.git" {
		t.Errorf("expected https://g.wii.pub:8443/my-org/my-repo.git, got: %s", got)
	}
}

func TestResource_HTTPURL(t *testing.T) {
	r := Resource{URL: "g.wii.pub"}
	if got := r.HTTPURL("my-org/my-repo"); got != "http://g.wii.pub/my-org/my-repo.git" {
		t.Errorf("expected http://g.wii.pub/my-org/my-repo.git, got: %s", got)
	}
}

func TestResource_HTTPURL_WithPort(t *testing.T) {
	r := Resource{URL: "g.wii.pub:8080"}
	if got := r.HTTPURL("my-org/my-repo"); got != "http://g.wii.pub:8080/my-org/my-repo.git" {
		t.Errorf("expected http://g.wii.pub:8080/my-org/my-repo.git, got: %s", got)
	}
}

// --- URL auto-prefix test (now strips scheme instead of adding) ---

func TestLoad_ResourceURLStripsScheme(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.mycompany.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	res := cfg.Resources["gl"]
	if res.URL != "gitlab.mycompany.com" {
		t.Errorf("expected host-only URL, got: %s", res.URL)
	}
}

func TestLoad_ResourceURLStripsHTTPSPrefix(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.mycompany.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	res := cfg.Resources["gl"]
	if res.URL != "gitlab.mycompany.com" {
		t.Errorf("expected stripped URL, got: %s", res.URL)
	}
	if res.Scheme() != "https" {
		t.Errorf("expected scheme https, got: %s", res.Scheme())
	}
}

func TestLoad_ResourceURLStripsHTTPPrefixWithPort(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: http://gitlab.mycompany.com:8080
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	res := cfg.Resources["gl"]
	if res.URL != "gitlab.mycompany.com:8080" {
		t.Errorf("expected stripped URL with port, got: %s", res.URL)
	}
	if res.Scheme() != "http" {
		t.Errorf("expected scheme http, got: %s", res.Scheme())
	}
}

func TestLoad_ResourceURLNoPrefixIsAuto(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.mycompany.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	res := cfg.Resources["gl"]
	if res.Scheme() != "" {
		t.Errorf("expected empty scheme (auto), got: %s", res.Scheme())
	}
	if res.APIURL() != "https://gitlab.mycompany.com" {
		t.Errorf("expected https://gitlab.mycompany.com, got: %s", res.APIURL())
	}
}

// --- Enabled field tests ---

// helper for creating *bool pointers in tests
func boolPtr(b bool) *bool {
	return &b
}

func TestLoad_ResourceEnabled(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantBool bool
		wantNil  bool
	}{
		{
			name: "enabled true",
			yaml: `base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.com
    token: test
    enabled: true
groups:
  - name: test
    resource: gl
    path: my-org`,
			wantBool: true,
			wantNil:  false,
		},
		{
			name: "enabled false",
			yaml: `base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.com
    token: test
    enabled: false
groups:
  - name: test
    resource: gl
    path: my-org`,
			wantBool: false,
			wantNil:  false,
		},
		{
			name: "enabled omitted defaults to true",
			yaml: `base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.yml")
			os.WriteFile(path, []byte(tt.yaml), 0644)

			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			res := cfg.Resources["gl"]
			if tt.wantNil {
				if res.Enabled != nil {
					t.Errorf("expected Enabled to be nil, got %v", *res.Enabled)
				}
				if !res.IsEnabled() {
					t.Error("IsEnabled() should return true when Enabled is nil")
				}
			} else {
				if res.Enabled == nil {
					t.Fatal("expected Enabled to be non-nil")
				}
				if *res.Enabled != tt.wantBool {
					t.Errorf("expected Enabled=%v, got %v", tt.wantBool, *res.Enabled)
				}
				if res.IsEnabled() != tt.wantBool {
					t.Errorf("expected IsEnabled()=%v", tt.wantBool)
				}
			}
		})
	}
}

func TestLoad_GroupEnabledAndExcludeRepos(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.com
    token: test
groups:
  - name: frontend
    resource: gl
    path: my-org/frontend
    enabled: false
    exclude_repos:
      - deprecated-app
      - temp-repo
  - name: backend
    resource: gl
    path: my-org/backend
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check frontend group
	if len(cfg.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(cfg.Groups))
	}
	fe := cfg.Groups[0]
	if fe.Name != "frontend" {
		t.Fatalf("expected group 'frontend', got %s", fe.Name)
	}
	if fe.Enabled == nil || *fe.Enabled != false {
		t.Error("expected frontend Enabled=false")
	}
	if fe.IsEnabled() {
		t.Error("expected frontend IsEnabled()=false")
	}
	if len(fe.ExcludeRepos) != 2 {
		t.Fatalf("expected 2 exclude_repos, got %d", len(fe.ExcludeRepos))
	}
	if fe.ExcludeRepos[0] != "deprecated-app" || fe.ExcludeRepos[1] != "temp-repo" {
		t.Errorf("unexpected exclude_repos: %v", fe.ExcludeRepos)
	}

	// Check backend group (defaults)
	be := cfg.Groups[1]
	if be.Name != "backend" {
		t.Fatalf("expected group 'backend', got %s", be.Name)
	}
	if be.Enabled != nil {
		t.Error("expected backend Enabled to be nil (default)")
	}
	if !be.IsEnabled() {
		t.Error("expected backend IsEnabled()=true (default)")
	}
	if len(be.ExcludeRepos) != 0 {
		t.Errorf("expected no exclude_repos for backend, got %v", be.ExcludeRepos)
	}
}

func TestLoad_GroupNoResource_ManualMode(t *testing.T) {
	content := `
base: ~/projects
groups:
  - name: local-tools
    local_path: ./tools
    repos:
      - name: my-script
        url: git@github.com:user/my-script.git
        path: my-script
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(cfg.Groups))
	}
	g := cfg.Groups[0]
	if g.Name != "local-tools" {
		t.Errorf("expected group name 'local-tools', got: %s", g.Name)
	}
	if g.Resource != "" {
		t.Errorf("expected empty resource, got: %s", g.Resource)
	}
	if len(g.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(g.Repos))
	}
	if g.Repos[0].URL != "git@github.com:user/my-script.git" {
		t.Errorf("unexpected repo URL: %s", g.Repos[0].URL)
	}
}

func TestLoad_GroupNoResourceNoPath_Ok(t *testing.T) {
	content := `
base: ~/projects
groups:
  - name: my-group
    repos:
      - name: repo1
        url: https://github.com/org/repo1.git
        path: repo1
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed without resource and path: %v", err)
	}
	if cfg.Groups[0].LocalPath != "./my-group" {
		t.Errorf("expected default local_path './my-group', got: %s", cfg.Groups[0].LocalPath)
	}
}

func TestLoad_GroupWithResourceMissingPath_Error(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: frontend
    resource: gl
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for group with resource but missing path")
	}
	if !strings.Contains(err.Error(), "path") {
		t.Errorf("error should mention path, got: %v", err)
	}
}

func TestLoad_RepoNoResourceWithURL_Ok(t *testing.T) {
	content := `
base: ~/projects
repos:
  - name: my-repo
    url: git@github.com:user/my-repo.git
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should succeed with url but no resource: %v", err)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0].URL != "git@github.com:user/my-repo.git" {
		t.Errorf("unexpected URL: %s", cfg.Repos[0].URL)
	}
	if cfg.Repos[0].LocalPath != "./my-repo" {
		t.Errorf("expected default local_path './my-repo', got: %s", cfg.Repos[0].LocalPath)
	}
}

func TestLoad_RepoNoResourceNoURL_Error(t *testing.T) {
	content := `
base: ~/projects
repos:
  - name: my-repo
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for repo without resource and url")
	}
	if !strings.Contains(err.Error(), "resource") || !strings.Contains(err.Error(), "url") {
		t.Errorf("error should mention 'resource or url', got: %v", err)
	}
}

func TestLoad_RepoEnabled(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: gitlab.com
    token: test
repos:
  - name: dotfiles
    resource: gl
    url: https://gitlab.com/me/dotfiles.git
    enabled: false
  - name: notes
    resource: gl
    url: https://gitlab.com/me/notes.git
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(cfg.Repos))
	}

	// dotfiles: explicitly disabled
	dot := cfg.Repos[0]
	if dot.Name != "dotfiles" {
		t.Fatalf("expected repo 'dotfiles', got %s", dot.Name)
	}
	if dot.Enabled == nil || *dot.Enabled != false {
		t.Error("expected dotfiles Enabled=false")
	}
	if dot.IsEnabled() {
		t.Error("expected dotfiles IsEnabled()=false")
	}

	// notes: default (nil → enabled)
	notes := cfg.Repos[1]
	if notes.Name != "notes" {
		t.Fatalf("expected repo 'notes', got %s", notes.Name)
	}
	if notes.Enabled != nil {
		t.Error("expected notes Enabled to be nil (default)")
	}
	if !notes.IsEnabled() {
		t.Error("expected notes IsEnabled()=true (default)")
	}
}

// --- YAML indent tests ---

func TestWriteConfig_DefaultIndent(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Write back and verify default 2-space indent
	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	rawContent := string(data)
	// Default indent should be 2 spaces (e.g., "  provider:")
	if !strings.Contains(rawContent, "  provider:") {
		t.Errorf("expected 2-space indent for default, got:\n%s", rawContent)
	}
}

func TestWriteConfig_CustomIndent(t *testing.T) {
	content := `
base: ~/projects
yaml_indent: 4
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.YAMLIndent != 4 {
		t.Fatalf("expected YAMLIndent=4, got %d", cfg.YAMLIndent)
	}

	// Write back and verify 4-space indent
	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	rawContent := string(data)
	// Should use 4-space indent (e.g., "    provider:")
	if !strings.Contains(rawContent, "    provider:") {
		t.Errorf("expected 4-space indent, got:\n%s", rawContent)
	}

	// yaml_indent field itself should NOT appear in the output file
	if strings.Contains(rawContent, "yaml_indent") {
		t.Errorf("yaml_indent field should not appear in output file, got:\n%s", rawContent)
	}
}

// --- DedupGroupRepos tests ---

func TestDedupGroupRepos_Basic(t *testing.T) {
	content := `
base: ~/projects
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
      - name: shared-utils
        url: https://gitlab.com/my-org/core/shared-utils.git
        path: my-org/core/shared-utils
      - name: core-only
        url: https://gitlab.com/my-org/core/core-only.git
        path: my-org/core/core-only
  - name: infra-team
    resource: gl
    path: my-org/infra
    local_path: ./infra
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/infra/api-lib.git
        path: my-org/infra/api-lib
      - name: deploy-tool
        url: https://gitlab.com/my-org/infra/deploy-tool.git
        path: my-org/infra/deploy-tool
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	excluded, err := DedupGroupRepos(path, "core-team", []string{"api-lib"})
	if err != nil {
		t.Fatalf("DedupGroupRepos failed: %v", err)
	}
	if len(excluded) != 1 || excluded[0] != "api-lib" {
		t.Errorf("expected excluded [api-lib], got %v", excluded)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after DedupGroupRepos failed: %v", err)
	}

	coreGroup := cfg.Groups[0]
	if len(coreGroup.Repos) != 2 {
		t.Fatalf("expected 2 repos in core-team, got %d", len(coreGroup.Repos))
	}
	for _, r := range coreGroup.Repos {
		if r.Name == "api-lib" {
			t.Error("api-lib should have been removed from repos")
		}
	}

	if len(coreGroup.ExcludeRepos) != 1 || coreGroup.ExcludeRepos[0] != "api-lib" {
		t.Errorf("expected exclude_repos [api-lib], got %v", coreGroup.ExcludeRepos)
	}

	infraGroup := cfg.Groups[1]
	if len(infraGroup.Repos) != 2 {
		t.Errorf("expected 2 repos in infra-team (unchanged), got %d", len(infraGroup.Repos))
	}
	if len(infraGroup.ExcludeRepos) != 0 {
		t.Errorf("expected no exclude_repos for infra-team, got %v", infraGroup.ExcludeRepos)
	}
}

func TestDedupGroupRepos_NoDuplicates(t *testing.T) {
	content := `
base: ~/projects
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
    exclude_repos:
      - api-lib
    repos:
      - name: api-lib
        url: https://gitlab.com/my-org/core/api-lib.git
        path: my-org/core/api-lib
      - name: other
        url: https://gitlab.com/my-org/core/other.git
        path: my-org/core/other
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	excluded, err := DedupGroupRepos(path, "core-team", []string{"api-lib"})
	if err != nil {
		t.Fatalf("DedupGroupRepos failed: %v", err)
	}
	if len(excluded) != 1 || excluded[0] != "api-lib" {
		t.Errorf("expected excluded [api-lib], got %v", excluded)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	coreGroup := cfg.Groups[0]
	if len(coreGroup.ExcludeRepos) != 1 || coreGroup.ExcludeRepos[0] != "api-lib" {
		t.Errorf("expected single exclude_repos entry [api-lib], got %v", coreGroup.ExcludeRepos)
	}
	if len(coreGroup.Repos) != 1 {
		t.Errorf("expected 1 repo (api-lib removed), got %d", len(coreGroup.Repos))
	}
}

func TestDedupGroupRepos_GlobCovered(t *testing.T) {
	content := `
base: ~/projects
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
    exclude_repos:
      - "test-*"
    repos:
      - name: test-integration
        url: https://gitlab.com/my-org/core/test-integration.git
        path: my-org/core/test-integration
      - name: other
        url: https://gitlab.com/my-org/core/other.git
        path: my-org/core/other
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	excluded, err := DedupGroupRepos(path, "core-team", []string{"test-integration"})
	if err != nil {
		t.Fatalf("DedupGroupRepos failed: %v", err)
	}
	if len(excluded) != 1 {
		t.Errorf("expected 1 excluded, got %v", excluded)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	coreGroup := cfg.Groups[0]
	if len(coreGroup.ExcludeRepos) != 1 || coreGroup.ExcludeRepos[0] != "test-*" {
		t.Errorf("expected exclude_repos unchanged [test-*], got %v", coreGroup.ExcludeRepos)
	}
	if len(coreGroup.Repos) != 1 || coreGroup.Repos[0].Name != "other" {
		t.Errorf("expected only 'other' repo remaining, got %v", coreGroup.Repos)
	}
}

func TestDedupGroupRepos_NoMatches(t *testing.T) {
	content := `
base: ~/projects
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
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	excluded, err := DedupGroupRepos(path, "core-team", []string{"nonexistent"})
	if err != nil {
		t.Fatalf("DedupGroupRepos failed: %v", err)
	}
	if len(excluded) != 0 {
		t.Errorf("expected 0 excluded, got %v", excluded)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Groups[0].Repos) != 1 {
		t.Error("repos should be unchanged")
	}
}

func TestDedupGroupRepos_GroupNotFound(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups: []
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := DedupGroupRepos(path, "nonexistent", []string{"repo1"})
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestIsExcludedName(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		repoName string
		want     bool
	}{
		{"exact match", []string{"app"}, "app", true},
		{"no match", []string{"other"}, "app", false},
		{"glob match", []string{"api-*"}, "api-lib", true},
		{"glob no match", []string{"api-*"}, "web-app", false},
		{"empty patterns", nil, "app", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExcludedName(tt.patterns, tt.repoName); got != tt.want {
				t.Errorf("isExcludedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteConfig_YAMLIndentPreservedAfterWrite(t *testing.T) {
	content := `
base: ~/projects
yaml_indent: 4
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: test
    resource: gl
    path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Write back
	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	// YAMLIndent should be preserved on the in-memory config
	if cfg.YAMLIndent != 4 {
		t.Errorf("YAMLIndent should be preserved in memory after write, got %d", cfg.YAMLIndent)
	}

	// Load again — yaml_indent was stripped from file, so should default to 0
	cfg2, err := Load(path)
	if err != nil {
		t.Fatalf("Load after write failed: %v", err)
	}
	if cfg2.YAMLIndent != 0 {
		t.Errorf("yaml_indent stripped from file, expected 0 on reload, got %d", cfg2.YAMLIndent)
	}
}
