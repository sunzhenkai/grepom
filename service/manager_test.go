package service

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/wii/grepom/config"
)

func testManager(t *testing.T) *Manager {
	t.Helper()
	old := StateHomeFunc
	t.Cleanup(func() { StateHomeFunc = old })
	base := t.TempDir()
	StateHomeFunc = func() (string, error) { return base, nil }

	cfgPath := filepath.Join(t.TempDir(), ".grepom.yml")
	mgr, err := NewManager(cfgPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	return mgr
}

func TestEvaluateStatusExited(t *testing.T) {
	rec := Record{PID: 999999, LastStatus: StatusRunning}
	if got := evaluateStatus(rec); got != StatusExited {
		t.Fatalf("status = %q", got)
	}
}

func TestDuplicateRunProtection(t *testing.T) {
	mgr := testManager(t)
	rec, err := mgr.Run(RunOptions{
		Name:    "demo",
		Cwd:     t.TempDir(),
		Command: config.ServiceCommand{Args: []string{"sleep", "30"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = mgr.Kill("demo", true)
	})

	_, err = mgr.Run(RunOptions{
		Name:    "demo",
		Cwd:     t.TempDir(),
		Command: config.ServiceCommand{Args: []string{"sleep", "30"}},
	})
	if err == nil {
		t.Fatal("expected duplicate run error")
	}
	if rec.PID <= 0 {
		t.Fatalf("invalid pid %d", rec.PID)
	}
}

func TestListIncludesPathAndStatus(t *testing.T) {
	mgr := testManager(t)
	cwd := t.TempDir()
	_, err := mgr.Run(RunOptions{
		Name:    "api",
		Cwd:     cwd,
		Command: config.ServiceCommand{Args: []string{"sleep", "10"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mgr.Kill("api", true) })

	entries, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d", len(entries))
	}
	if entries[0].Status != StatusRunning {
		t.Fatalf("status = %q", entries[0].Status)
	}
	if entries[0].Record.Cwd != cwd {
		t.Fatalf("cwd = %q", entries[0].Record.Cwd)
	}
}

func TestRestartRunningService(t *testing.T) {
	mgr := testManager(t)
	cwd := t.TempDir()
	first, err := mgr.Run(RunOptions{
		Name:    "api",
		Cwd:     cwd,
		Command: config.ServiceCommand{Args: []string{"sleep", "30"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mgr.Kill("api", true) })

	second, err := mgr.Restart("api")
	if err != nil {
		t.Fatal(err)
	}
	if second.PID == first.PID {
		t.Fatalf("expected new pid, got same %d", second.PID)
	}
	entry, err := mgr.Status("api")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Status != StatusRunning {
		t.Fatalf("status = %q", entry.Status)
	}
}

func TestRestartExitedService(t *testing.T) {
	mgr := testManager(t)
	cwd := t.TempDir()
	err := WithRegistryLock(mgr.Registry, func(reg *Registry) error {
		reg.Services["worker"] = Record{
			Name:        "worker",
			PID:         999999,
			Cwd:         cwd,
			Command:     "sleep 30",
			CommandArgs: []string{"sleep", "30"},
			LogPath:     LogPathForService(mgr.StateDir, "worker"),
			StartedAt:   time.Now().UTC(),
			LastStatus:  StatusExited,
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	second, err := mgr.Restart("worker")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mgr.Kill("worker", true) })
	if second.PID <= 0 {
		t.Fatalf("invalid pid %d", second.PID)
	}
}

func TestCleanRemovesExited(t *testing.T) {
	mgr := testManager(t)
	err := WithRegistryLock(mgr.Registry, func(reg *Registry) error {
		reg.Services["old"] = Record{
			Name:       "old",
			PID:        999999,
			Cwd:        "/tmp",
			Command:    "echo",
			LogPath:    LogPathForService(mgr.StateDir, "old"),
			StartedAt:  time.Now().UTC(),
			LastStatus: StatusExited,
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	removed, err := mgr.Clean(CleanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d", removed)
	}
}
