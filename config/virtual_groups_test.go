package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_VirtualGroups(t *testing.T) {
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
  - name: backend
    resource: gl
    path: my-org/backend
virtual_groups:
  work:
    groups:
      - frontend
      - backend
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.VirtualGroups) != 1 {
		t.Fatalf("expected 1 virtual group, got %d", len(cfg.VirtualGroups))
	}
	vg, ok := cfg.VirtualGroups["work"]
	if !ok {
		t.Fatal("expected virtual group 'work'")
	}
	if len(vg.Groups) != 2 || vg.Groups[0] != "frontend" || vg.Groups[1] != "backend" {
		t.Errorf("unexpected virtual group members: %v", vg.Groups)
	}
}

func TestLoad_VirtualGroupsDefaultEmpty(t *testing.T) {
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
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.VirtualGroups == nil {
		t.Fatal("expected non-nil VirtualGroups map")
	}
	if len(cfg.VirtualGroups) != 0 {
		t.Errorf("expected empty virtual groups, got %d", len(cfg.VirtualGroups))
	}
}

func TestLoad_VirtualGroupSameNameAsRealGroup(t *testing.T) {
	content := `
base: ~/projects
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test
groups:
  - name: work
    resource: gl
    path: my-org/work
  - name: frontend
    resource: gl
    path: my-org/frontend
virtual_groups:
  work:
    groups:
      - frontend
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should allow same name for real group and virtual group: %v", err)
	}
	if _, _, err := cfg.FindGroup("work"); err != nil {
		t.Fatalf("real group work not found: %v", err)
	}
	if _, err := cfg.FindVirtualGroup("work"); err != nil {
		t.Fatalf("virtual group work not found: %v", err)
	}
}

func TestLoad_VirtualGroupMissingMember(t *testing.T) {
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
virtual_groups:
  work:
    groups:
      - missing-group
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing member reference")
	}
	if !strings.Contains(err.Error(), "missing-group") {
		t.Errorf("error should mention missing-group, got: %v", err)
	}
}

func TestResolveGroupSelection_UnionAndDedup(t *testing.T) {
	cfg := &Config{
		Groups: []Group{
			{Name: "frontend"},
			{Name: "backend"},
			{Name: "infra"},
		},
		VirtualGroups: map[string]VirtualGroup{
			"work": {Groups: []string{"frontend", "backend"}},
		},
	}

	selected, err := cfg.ResolveGroupSelection("frontend", "work")
	if err != nil {
		t.Fatalf("ResolveGroupSelection failed: %v", err)
	}
	if len(selected) != 2 {
		t.Fatalf("expected 2 groups after dedup, got %v", selected)
	}
	if selected[0] != "frontend" || selected[1] != "backend" {
		t.Errorf("unexpected selection order: %v", selected)
	}
}

func TestResolveGroupSelection_MissingVirtualGroup(t *testing.T) {
	cfg := &Config{
		Groups:        []Group{{Name: "frontend"}},
		VirtualGroups: map[string]VirtualGroup{},
	}

	_, err := cfg.ResolveGroupSelection("", "missing")
	if err == nil {
		t.Fatal("expected error for missing virtual group")
	}
	if !strings.Contains(err.Error(), "virtual group") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveGroupSelection_EmptyMeansAll(t *testing.T) {
	cfg := &Config{
		Groups: []Group{{Name: "frontend"}},
	}

	selected, err := cfg.ResolveGroupSelection("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != nil {
		t.Errorf("expected nil selection for no filters, got %v", selected)
	}
}
