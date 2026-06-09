package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wii/grepom/service"
)

func TestParseSvcRunArgs(t *testing.T) {
	isConfigured := func(name string) bool { return name == "api" }

	name, cmd, useConfig, err := parseSvcRunArgs([]string{"sleep", "8"}, isConfigured)
	if err != nil || useConfig || name == "" || len(cmd.Args) != 2 {
		t.Fatalf("default command parse: name=%q cmd=%#v useConfig=%v err=%v", name, cmd, useConfig, err)
	}

	name, cmd, useConfig, err = parseSvcRunArgs([]string{"api", "make", "dev"}, isConfigured)
	if err != nil || useConfig || name != "api" || len(cmd.Args) != 2 {
		t.Fatalf("named command parse: name=%q cmd=%#v useConfig=%v err=%v", name, cmd, useConfig, err)
	}

	name, _, useConfig, err = parseSvcRunArgs([]string{"api"}, isConfigured)
	if err != nil || !useConfig || name != "api" {
		t.Fatalf("config parse: name=%q useConfig=%v err=%v", name, useConfig, err)
	}
}

func TestPrintSvcTableIncludesPathAndStatus(t *testing.T) {
	var buf bytes.Buffer
	entries := []service.Entry{
		{
			Record: service.Record{
				Name:    "api",
				PID:     100,
				Cwd:     "/tmp/api",
				Command: "make dev",
				LogPath: "/tmp/logs/api.log",
			},
			Status: service.StatusRunning,
		},
	}
	if err := printSvcTable(&buf, entries); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"NAME", "STATUS", "PATH", "api", "running", "/tmp/api", "make dev"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}
