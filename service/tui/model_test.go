package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/wii/grepom/config"
	"github.com/wii/grepom/service"
)

func TestModelListViewSelectedMarkerSpacing(t *testing.T) {
	m := model{
		cursor: 0,
		entries: []service.Entry{
			{
				Record: service.Record{Name: "api", PID: 42, Cwd: "/tmp/api"},
				Status: service.StatusRunning,
			},
		},
	}
	out := m.listView()
	if !contains(out, "> api") {
		t.Fatalf("expected space between marker and name: %q", out)
	}
}

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

func testManager(t *testing.T) *service.Manager {
	t.Helper()
	old := service.StateHomeFunc
	t.Cleanup(func() { service.StateHomeFunc = old })
	base := t.TempDir()
	service.StateHomeFunc = func() (string, error) { return base, nil }

	cfgPath := filepath.Join(t.TempDir(), ".grepom.yml")
	mgr, err := service.NewManager(cfgPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	return mgr
}

func TestRestartNoServiceSelected(t *testing.T) {
	m := model{mgr: testManager(t)}
	err := m.restart()
	if err == nil {
		t.Fatal("expected error when no service selected")
	}
	if !contains(err.Error(), "no service selected") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRestartSuccess(t *testing.T) {
	mgr := testManager(t)
	cwd := t.TempDir()
	rec, err := mgr.Run(service.RunOptions{
		Name:    "api",
		Cwd:     cwd,
		Command: config.ServiceCommand{Args: []string{"sleep", "30"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mgr.Kill("api", true) })

	entries, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}

	m := model{
		mgr:     mgr,
		cursor:  0,
		entries: entries,
		mode:    viewList,
	}

	err = m.restart()
	if err != nil {
		t.Fatalf("restart failed: %v", err)
	}
	if !contains(m.message, "restarted api") {
		t.Fatalf("message = %q", m.message)
	}
	// PID should have changed after restart
	if m.entries[0].Record.PID == rec.PID {
		t.Fatalf("PID did not change after restart: %d", rec.PID)
	}
	t.Cleanup(func() { _ = mgr.Kill("api", true) })
}

func TestRestartFailureNoCommand(t *testing.T) {
	mgr := testManager(t)

	// Manually create a registry entry with no command info
	entries := []service.Entry{
		{
			Record: service.Record{Name: "broken", PID: 999999, Cwd: "/tmp"},
			Status: service.StatusExited,
		},
	}

	m := model{
		mgr:     mgr,
		cursor:  0,
		entries: entries,
		mode:    viewList,
	}

	err := m.restart()
	if err == nil {
		t.Fatal("expected error for service with no command")
	}
}

func TestRestartMessageOnSuccess(t *testing.T) {
	mgr := testManager(t)
	cwd := t.TempDir()
	_, err := mgr.Run(service.RunOptions{
		Name:    "web",
		Cwd:     cwd,
		Command: config.ServiceCommand{Args: []string{"sleep", "30"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mgr.Kill("web", true) })

	entries, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}

	m := model{
		mgr:     mgr,
		cursor:  0,
		entries: entries,
		mode:    viewList,
	}

	err = m.restart()
	if err != nil {
		t.Fatalf("restart failed: %v", err)
	}
	if !contains(m.message, "restarted web") || !contains(m.message, "pid") {
		t.Fatalf("expected restart message with pid, got: %q", m.message)
	}
	t.Cleanup(func() { _ = mgr.Kill("web", true) })
}

func TestListViewLongNameAlignment(t *testing.T) {
	m := model{
		cursor: 0,
		entries: []service.Entry{
			{
				Record: service.Record{Name: "api", PID: 42, Cwd: "/tmp/api"},
				Status: service.StatusRunning,
			},
			{
				Record: service.Record{Name: "very-long-service-name-that-exceeds-default", PID: 100, Cwd: "/tmp/long"},
				Status: service.StatusRunning,
			},
			{
				Record: service.Record{Name: "web", PID: 0, Cwd: "/tmp/web"},
				Status: service.StatusExited,
			},
		},
	}
	out := m.listView()
	lines := strings.Split(out, "\n")

	// Find the header line (contains "NAME")
	var headerIdx int
	for i, l := range lines {
		if contains(l, "NAME") && contains(l, "STATUS") {
			headerIdx = i
			break
		}
	}

	// Collect all data lines (lines starting with " " or ">")
	var dataLines []string
	for _, l := range lines {
		if len(l) > 0 && (l[0] == ' ' || l[0] == '>') && !contains(l, "NAME") {
			dataLines = append(dataLines, l)
		}
	}

	if len(dataLines) != 3 {
		t.Fatalf("expected 3 data lines, got %d: %v", len(dataLines), dataLines)
	}

	// Find where STATUS column starts in each line
	// All data lines should have STATUS at the same column position
	statusCol := -1
	for _, dl := range dataLines {
		// Find "running" or "exited" in the line
		for _, status := range []string{"running", "exited"} {
			idx := indexOf(dl, status)
			if idx >= 0 {
				if statusCol == -1 {
					statusCol = idx
				} else if idx != statusCol {
					t.Fatalf("STATUS column misaligned in line %q (col %d, expected %d)", dl, idx, statusCol)
				}
			}
		}
	}

	// Header STATUS should also align
	headerStatusCol := indexOf(lines[headerIdx], "STATUS")
	if headerStatusCol != statusCol {
		t.Fatalf("header STATUS at col %d but data STATUS at col %d\nheader: %q\ndata[0]: %q",
			headerStatusCol, statusCol, lines[headerIdx], dataLines[0])
	}
}

func TestListViewContainsRestartKeyHint(t *testing.T) {
	m := model{
		entries: []service.Entry{
			{
				Record: service.Record{Name: "api", PID: 42, Cwd: "/tmp/api"},
				Status: service.StatusRunning,
			},
		},
	}
	out := m.listView()
	if !contains(out, "R restart") {
		t.Fatalf("list view missing 'R restart' hint: %q", out)
	}
}
