package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ResourcesMapFormat(t *testing.T) {
	content := `
base: ~/projects
resources:
  my-gitlab:
    provider: gitlab
    url: gitlab.example.com
    token: ${GITLAB_TOKEN}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	res, ok := cfg.Resources["my-gitlab"]
	if !ok {
		t.Fatal("expected resource my-gitlab")
	}
	if res.Provider != "gitlab" || res.URL != "gitlab.example.com" {
		t.Errorf("unexpected resource: %+v", res)
	}
}

func TestLoad_ResourcesListFormat(t *testing.T) {
	content := `
base: ~/projects
resources:
  - name: my-gitlab
    provider: gitlab
    url: gitlab.example.com
    token: ${GITLAB_TOKEN}
  - name: my-github
    provider: github
    url: github.com
    token: ${GITHUB_TOKEN}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load list format failed: %v", err)
	}
	if len(cfg.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(cfg.Resources))
	}
	if cfg.Resources["my-gitlab"].Provider != "gitlab" {
		t.Errorf("expected gitlab provider, got %s", cfg.Resources["my-gitlab"].Provider)
	}
	if cfg.Resources["my-github"].Provider != "github" {
		t.Errorf("expected github provider, got %s", cfg.Resources["my-github"].Provider)
	}
}

func TestLoad_ResourcesListMissingName(t *testing.T) {
	content := `
base: ~/projects
resources:
  - provider: gitlab
    url: gitlab.example.com
    token: x
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention name, got: %v", err)
	}
}

func TestLoad_ResourcesListDuplicateName(t *testing.T) {
	content := `
base: ~/projects
resources:
  - name: dup
    provider: gitlab
    url: gitlab.example.com
    token: a
  - name: dup
    provider: github
    url: github.com
    token: b
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention duplicate, got: %v", err)
	}
}

func TestLoad_ResourcesInvalidType(t *testing.T) {
	content := `
base: ~/projects
resources: "not-a-map-or-list"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid resources type")
	}
	if !strings.Contains(err.Error(), "resources") {
		t.Errorf("error should mention resources, got: %v", err)
	}
}
