package service

import (
	"path/filepath"
	"testing"
)

func TestScopeFromPathStable(t *testing.T) {
	a, err := ScopeFromPath(filepath.Join("/tmp", "work", ".grepom.yml"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := ScopeFromPath(filepath.Join("/tmp", "work", ".grepom.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("scope not stable: %q vs %q", a, b)
	}
}

func TestStateDirUsesStateHome(t *testing.T) {
	old := StateHomeFunc
	t.Cleanup(func() { StateHomeFunc = old })
	stateHome := t.TempDir()
	StateHomeFunc = func() (string, error) {
		return stateHome, nil
	}
	scope, err := ScopeFromPath("/tmp/demo/.grepom.yml")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := StateDir(scope)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(stateHome, "grepom", "services", scope)
	if dir != want {
		t.Fatalf("state dir = %q, want %q", dir, want)
	}
}
