package service

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestReadTailLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "svc.log")
	content := strings.Repeat("line\n", 200)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := ReadTailLines(path, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
}

func TestFollowLogStopsOnCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "svc.log")
	if err := os.WriteFile(path, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- FollowLog(ctx, path, os.Stdout)
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("follow returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("follow did not stop")
	}
}

func TestOpenLogPrintsPathWhenNoEditor(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fallback path printing is only exercised when xdg-open is unavailable")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "svc.log")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")
	var buf strings.Builder
	if err := OpenLog(path, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), path) {
		t.Fatalf("expected path printed, got %q", buf.String())
	}
}
