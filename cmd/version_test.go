package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	old := Version
	t.Cleanup(func() { Version = old })
	Version = "v0.2.0"

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	versionCmd.SetErr(&buf)
	versionCmd.SetArgs(nil)

	versionCmd.Run(versionCmd, nil)
	if got := strings.TrimSpace(buf.String()); got != "v0.2.0" {
		t.Fatalf("got %q, want %q", got, "v0.2.0")
	}
}
