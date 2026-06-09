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

func TestStateDirUsesUserConfig(t *testing.T) {
	old := UserConfigDirFunc
	t.Cleanup(func() { UserConfigDirFunc = old })
	UserConfigDirFunc = func() (string, error) {
		return t.TempDir(), nil
	}
	scope, err := ScopeFromPath("/tmp/demo/.grepom.yml")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := StateDir(scope)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(filepath.Dir(dir)) != "services" {
		t.Fatalf("unexpected state dir: %s", dir)
	}
}
