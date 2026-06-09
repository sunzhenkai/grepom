package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/service"
)

func TestCompleteSvcNamesMergesConfigAndRegistry(t *testing.T) {
	oldStateHome := service.StateHomeFunc
	t.Cleanup(func() { service.StateHomeFunc = oldStateHome })

	dir := t.TempDir()
	stateHome := filepath.Join(dir, "state")
	service.StateHomeFunc = func() (string, error) { return stateHome, nil }

	cfgPath := filepath.Join(dir, ".grepom.yml")
	cfgYAML := `
base: .
resources: {}
groups: []
repos: []
services:
  api:
    command: [sleep, "30"]
  web:
    command: [sleep, "30"]
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	mgr, err := service.NewManager(cfgPath, cfg.Services)
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.Run(service.RunOptions{
		Name:    "worker",
		Cwd:     dir,
		Command: config.ServiceCommand{Args: []string{"sleep", "30"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mgr.Kill("worker", true) })

	oldConfigFile := configFile
	t.Cleanup(func() { configFile = oldConfigFile })
	configFile = cfgPath

	names, directive := completeSvcNames(nil, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v", directive)
	}

	want := map[string]bool{"api": true, "web": true, "worker": true}
	if len(names) != len(want) {
		t.Fatalf("names = %v, want %d unique entries", names, len(want))
	}
	for _, name := range names {
		if !want[name] {
			t.Fatalf("unexpected name %q in %v", name, names)
		}
	}
}