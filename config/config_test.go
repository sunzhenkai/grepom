package config

import (
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
	if res.Token != "glpat-xxx" {
		t.Errorf("token not expanded, got: %s", res.Token)
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

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for undefined env var in token")
	}
	if !strings.Contains(err.Error(), "UNDEFINED_VAR") {
		t.Errorf("error should mention the env var name, got: %v", err)
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

	res := cfg.Resources["work-gl"]
	if res.Token != "glpat-write-test" {
		t.Fatalf("expected resolved token, got: %s", res.Token)
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

// --- URL auto-prefix test ---

func TestLoad_ResourceURLAutoHTTPS(t *testing.T) {
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
	if !strings.HasPrefix(res.URL, "https://") {
		t.Errorf("expected URL to have https:// prefix, got: %s", res.URL)
	}
}
