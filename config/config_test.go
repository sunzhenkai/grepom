package config

import (
	"os"
	"path/filepath"
	"testing"
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

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// os.ExpandEnv silently replaces undefined vars with empty string
	if cfg.Sources[0].Token != "" {
		t.Errorf("expected empty token for undefined env var, got: %q", cfg.Sources[0].Token)
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
