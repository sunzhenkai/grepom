package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ServicesConfig(t *testing.T) {
	content := `
base: ~/projects
resources: {}
groups: []
repos: []
services:
  api:
    cwd: ./backend
    command: make dev
  web:
    cwd: ./frontend
    command:
      - pnpm
      - dev
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".grepom.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	api, ok := cfg.Services["api"]
	if !ok {
		t.Fatal("expected api service")
	}
	if api.Command.Shell != "make dev" {
		t.Fatalf("api command = %#v", api.Command)
	}
	web := cfg.Services["web"]
	if len(web.Command.Args) != 2 || web.Command.Args[0] != "pnpm" {
		t.Fatalf("web command = %#v", web.Command)
	}
}

func TestResolveServiceCwd(t *testing.T) {
	configDir := "/tmp/project"
	if got := ResolveServiceCwd(configDir, ""); got != configDir {
		t.Fatalf("empty cwd = %q", got)
	}
	if got := ResolveServiceCwd(configDir, "./backend"); got != filepath.Join(configDir, "backend") {
		t.Fatalf("relative cwd = %q", got)
	}
	if got := ResolveServiceCwd(configDir, "/abs/path"); got != "/abs/path" {
		t.Fatalf("absolute cwd = %q", got)
	}
}

func TestFindService(t *testing.T) {
	cfg := &Config{
		Services: map[string]ServiceDef{
			"api": {Command: ServiceCommand{Shell: "make dev"}},
		},
	}
	def, err := cfg.FindService("api")
	if err != nil || def.Command.Shell != "make dev" {
		t.Fatalf("FindService api: %v %#v", err, def)
	}
	if _, err := cfg.FindService("missing"); err == nil {
		t.Fatal("expected missing service error")
	}
}
