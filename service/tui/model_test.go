package tui

import (
	"path/filepath"
	"testing"

	"github.com/wii/grepom/service"
)

func TestModelListViewShowsPathAndStatus(t *testing.T) {
	m := model{
		entries: []service.Entry{
			{
				Record: service.Record{Name: "api", PID: 42, Cwd: "/tmp/api", Command: "make dev"},
				Status: service.StatusRunning,
			},
		},
	}
	out := m.listView()
	if !containsAll(out, "api", "running", "42", "/tmp/api") {
		t.Fatalf("list view missing fields: %s", out)
	}
}

func TestModelServicePath(t *testing.T) {
	m := model{
		cursor: 0,
		entries: []service.Entry{
			{Record: service.Record{Cwd: "/tmp/web"}},
		},
	}
	if got := m.servicePath(); got != "/tmp/web" {
		t.Fatalf("servicePath = %q", got)
	}
}

func TestScopeIntegrationWithManager(t *testing.T) {
	dir := t.TempDir()
	scope, err := service.ScopeFromPath(filepath.Join(dir, ".grepom.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if scope == "" {
		t.Fatal("expected scope id")
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
