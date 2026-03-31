package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: ${TEST_TOKEN}
    groups:
      - path: my-org/frontend
        recursive: true
repos:
  - name: special
    url: https://gitlab.com/other/special.git
    path: ./special
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
	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Token != "glpat-xxx" {
		t.Errorf("token not expanded, got: %s", cfg.Sources[0].Token)
	}
	if len(cfg.Sources[0].Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(cfg.Sources[0].Groups))
	}
	if !cfg.Sources[0].Groups[0].Recursive {
		t.Error("expected recursive=true")
	}
	if len(cfg.Repos) != 1 || cfg.Repos[0].Name != "special" {
		t.Error("repo entry not parsed correctly")
	}
}

func TestLoad_MissingBase(t *testing.T) {
	content := `
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: test
    groups:
      - path: my-org
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
sources:
  - provider: bitbucket
    url: https://bitbucket.org
    token: test
    groups:
      - path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestLoad_UndefinedEnvVar(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: ${UNDEFINED_VAR}
    groups:
      - path: my-org
`
	os.Unsetenv("UNDEFINED_VAR")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for undefined env var in token")
	}
	if !strings.Contains(err.Error(), "UNDEFINED_VAR") {
		t.Errorf("error should mention the env var name, got: %v", err)
	}
}

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

func TestAddSource(t *testing.T) {
	os.Setenv("GL_TOKEN", "test-token-value")
	defer os.Unsetenv("GL_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	source := Source{
		Provider: "gitlab",
		URL:      "https://gitlab.com",
		Token:    "${GL_TOKEN}",
		Groups:   []GroupSource{{Path: "my-org", Recursive: true}},
	}

	err := AddSource(path, source)
	if err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after AddSource failed: %v", err)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Provider != "gitlab" {
		t.Error("provider mismatch")
	}
}

func TestAddRepo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	repo := RepoEntry{
		Name: "special",
		URL:  "https://gitlab.com/other/special.git",
		Path: "./special",
	}

	err := AddRepo(path, repo)
	if err != nil {
		t.Fatalf("AddRepo failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after AddRepo failed: %v", err)
	}
	if len(cfg.Repos) != 1 || cfg.Repos[0].Name != "special" {
		t.Error("repo not added correctly")
	}
}

func TestSyncGroups_AddNewGroups(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: test
    groups:
      - path: my-org/frontend
        recursive: true
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	newGroups := []GroupSource{
		{Path: "my-org/backend", Recursive: true},
		{Path: "my-org/mobile", Recursive: true},
	}

	err := SyncGroups(path, 0, newGroups)
	if err != nil {
		t.Fatalf("SyncGroups failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after SyncGroups failed: %v", err)
	}

	if len(cfg.Sources[0].Groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(cfg.Sources[0].Groups))
	}

	// Original group should be preserved
	if cfg.Sources[0].Groups[0].Path != "my-org/frontend" {
		t.Errorf("original group path wrong: %s", cfg.Sources[0].Groups[0].Path)
	}
}

func TestSyncGroups_NoDuplicates(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: test
    groups:
      - path: my-org/frontend
        recursive: true
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	newGroups := []GroupSource{
		{Path: "my-org/frontend", Recursive: true},
		{Path: "my-org/backend", Recursive: true},
	}

	err := SyncGroups(path, 0, newGroups)
	if err != nil {
		t.Fatalf("SyncGroups failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after SyncGroups failed: %v", err)
	}

	// Should only have 2 groups (no duplicate)
	if len(cfg.Sources[0].Groups) != 2 {
		t.Fatalf("expected 2 groups (no duplicates), got %d", len(cfg.Sources[0].Groups))
	}
}

func TestSyncGroups_NoChanges(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: test
    groups:
      - path: my-org/frontend
        recursive: true
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)
	originalContent := content

	newGroups := []GroupSource{
		{Path: "my-org/frontend", Recursive: true},
	}

	err := SyncGroups(path, 0, newGroups)
	if err != nil {
		t.Fatalf("SyncGroups failed: %v", err)
	}

	// File should not have been rewritten (no new groups)
	after, _ := os.ReadFile(path)
	if string(after) != originalContent {
		t.Error("config file should not be rewritten when no new groups")
	}
}

func TestSyncGroups_InvalidSourceIndex(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: test
    groups:
      - path: my-org/frontend
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	err := SyncGroups(path, 5, []GroupSource{{Path: "new-group"}})
	if err == nil {
		t.Fatal("expected error for invalid source index")
	}
}

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

// --- Token environment variable placeholder tests (Task 1.5) ---

func TestResolveToken_PlainText(t *testing.T) {
	result, err := resolveToken("glpat-xxxxxxxxxxxx")
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

	result, err := resolveToken("${MY_TOKEN}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "secret-value" {
		t.Errorf("expected resolved value, got: %s", result)
	}
}

func TestResolveToken_Empty(t *testing.T) {
	result, err := resolveToken("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got: %s", result)
	}
}

func TestResolveToken_UndefinedEnvVar(t *testing.T) {
	os.Unsetenv("DEFINITELY_NOT_SET_TOKEN")

	_, err := resolveToken("${DEFINITELY_NOT_SET_TOKEN}")
	if err == nil {
		t.Fatal("expected error for undefined env var")
	}
	if !strings.Contains(err.Error(), "DEFINITELY_NOT_SET_TOKEN") {
		t.Errorf("error should mention env var name, got: %v", err)
	}
}

func TestResolveToken_NotPlaceholder(t *testing.T) {
	// Strings that look like they might have vars but aren't full ${VAR} patterns
	result, err := resolveToken("prefix-${TOO_MANY_THINGS}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "prefix-${TOO_MANY_THINGS}" {
		t.Errorf("non-placeholder should be returned as-is, got: %s", result)
	}
}

func TestWriteConfig_PreservesTokenPlaceholder(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: ${TEST_WRITE_TOKEN}
    groups:
      - path: my-org
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

	// Token should be resolved at runtime
	if cfg.Sources[0].Token != "glpat-write-test" {
		t.Fatalf("expected resolved token, got: %s", cfg.Sources[0].Token)
	}

	// Add a repo and write back
	cfg.Repos = append(cfg.Repos, RepoEntry{Name: "test-repo", URL: "https://gitlab.com/test.git", Path: "test-repo"})
	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	// Read raw file and verify token placeholder is preserved
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

	// Token should still be resolved in memory
	if cfg.Sources[0].Token != "glpat-write-test" {
		t.Errorf("in-memory token should remain resolved, got: %s", cfg.Sources[0].Token)
	}
}

// --- Source name field tests (Task 2.6) ---

func TestLoad_SourceWithName(t *testing.T) {
	content := `
base: ~/projects
sources:
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.com
    token: test-token
    groups:
      - path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Sources[0].Name != "my-gitlab" {
		t.Errorf("expected source name 'my-gitlab', got: %s", cfg.Sources[0].Name)
	}
}

func TestLoad_DuplicateSourceNames(t *testing.T) {
	content := `
base: ~/projects
sources:
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.com
    token: test1
    groups:
      - path: org1
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.example.com
    token: test2
    groups:
      - path: org2
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate source names")
	}
	if !strings.Contains(err.Error(), "duplicate source name") {
		t.Errorf("error should mention duplicate name, got: %v", err)
	}
}

func TestFindSource_ByName(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Name: "my-gitlab", Provider: "gitlab", URL: "https://gitlab.com"},
			{Name: "my-github", Provider: "github", URL: "https://github.com"},
		},
	}

	idx, src, err := cfg.FindSource("my-github")
	if err != nil {
		t.Fatalf("FindSource failed: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if src.Provider != "github" {
		t.Errorf("expected github provider, got: %s", src.Provider)
	}
}

func TestFindSource_ByIndex(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Provider: "gitlab", URL: "https://gitlab.com"},
			{Provider: "github", URL: "https://github.com"},
		},
	}

	idx, src, err := cfg.FindSource("0")
	if err != nil {
		t.Fatalf("FindSource failed: %v", err)
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
	if src.Provider != "gitlab" {
		t.Errorf("expected gitlab provider, got: %s", src.Provider)
	}
}

func TestFindSource_NamePriorityOverIndex(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Name: "0", Provider: "github", URL: "https://github.com"},
			{Provider: "gitlab", URL: "https://gitlab.com"},
		},
	}

	idx, src, err := cfg.FindSource("0")
	if err != nil {
		t.Fatalf("FindSource failed: %v", err)
	}
	// Name "0" matches source 0 (github), not index 0
	if src.Provider != "github" {
		t.Errorf("name should take priority, expected github, got: %s", src.Provider)
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestFindSource_NotFound(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Provider: "gitlab", URL: "https://gitlab.com"},
		},
	}

	_, _, err := cfg.FindSource("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
}

func TestFindSource_IndexOutOfRange(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Provider: "gitlab", URL: "https://gitlab.com"},
		},
	}

	_, _, err := cfg.FindSource("5")
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestWriteConfig_SourceNamePreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	cfg := &Config{
		Base: "~/projects",
		Sources: []Source{
			{Name: "my-gitlab", Provider: "gitlab", URL: "https://gitlab.com", Token: "test"},
		},
	}
	cfg.rawTokens = map[int]string{0: "test"}

	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: my-gitlab") {
		t.Errorf("source name should be written to file, got:\n%s", content)
	}
}

func TestWriteConfig_SourceNameOmitted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	cfg := &Config{
		Base: "~/projects",
		Sources: []Source{
			{Provider: "gitlab", URL: "https://gitlab.com", Token: "test"},
		},
	}
	cfg.rawTokens = map[int]string{0: "test"}

	if err := writeConfig(path, cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "name:") {
		t.Errorf("name field should be omitted when empty, got:\n%s", content)
	}
}

// --- SyncRepos tests (Task 3.6) ---

func TestSyncRepos_AddNewRepos(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: test-token
    groups:
      - path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	newRepos := []RepoEntry{
		{Name: "repo1", URL: "https://gitlab.com/my-org/repo1.git", Path: "my-org/repo1"},
		{Name: "repo2", URL: "https://gitlab.com/my-org/repo2.git", Path: "my-org/repo2"},
	}

	added, err := SyncRepos(path, newRepos)
	if err != nil {
		t.Fatalf("SyncRepos failed: %v", err)
	}
	if added != 2 {
		t.Errorf("expected 2 repos added, got %d", added)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after SyncRepos failed: %v", err)
	}
	if len(cfg.Repos) != 2 {
		t.Fatalf("expected 2 repos in config, got %d", len(cfg.Repos))
	}
}

func TestSyncRepos_NoDuplicates(t *testing.T) {
	content := `
base: ~/projects
repos:
  - name: repo1
    url: https://gitlab.com/my-org/repo1.git
    path: my-org/repo1
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	newRepos := []RepoEntry{
		{Name: "repo1", URL: "https://gitlab.com/my-org/repo1.git", Path: "my-org/repo1"},
		{Name: "repo2", URL: "https://gitlab.com/my-org/repo2.git", Path: "my-org/repo2"},
	}

	added, err := SyncRepos(path, newRepos)
	if err != nil {
		t.Fatalf("SyncRepos failed: %v", err)
	}
	if added != 1 {
		t.Errorf("expected 1 repo added (dedup by URL), got %d", added)
	}
}

// --- Backward compatibility test (Task 4.4) ---

func TestLoad_PlainTokenBackwardCompat(t *testing.T) {
	content := `
base: ~/projects
sources:
  - provider: gitlab
    url: https://gitlab.com
    token: glpat-plain-text-token
    groups:
      - path: my-org
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Sources[0].Token != "glpat-plain-text-token" {
		t.Errorf("plain token should work, got: %s", cfg.Sources[0].Token)
	}
}

func TestAddSource_WithTokenPlaceholder(t *testing.T) {
	os.Setenv("ADD_SOURCE_TOKEN", "resolved-token")
	defer os.Unsetenv("ADD_SOURCE_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	source := Source{
		Name:     "test-source",
		Provider: "gitlab",
		URL:      "https://gitlab.com",
		Token:    "${ADD_SOURCE_TOKEN}",
		Groups:   []GroupSource{{Path: "my-org"}},
	}

	if err := AddSource(path, source); err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}

	// Verify file contains placeholder
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(data), "${ADD_SOURCE_TOKEN}") {
		t.Errorf("file should contain token placeholder, got:\n%s", string(data))
	}

	// Verify loading resolves the token
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Sources[0].Token != "resolved-token" {
		t.Errorf("token should be resolved, got: %s", cfg.Sources[0].Token)
	}
	if cfg.Sources[0].Name != "test-source" {
		t.Errorf("source name should be preserved, got: %s", cfg.Sources[0].Name)
	}
}
