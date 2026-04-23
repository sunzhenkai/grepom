package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createDirTestConfig(t *testing.T, dir string) string {
	t.Helper()
	content := `
base: ` + dir + `
resources:
  gl:
    provider: gitlab
    url: https://gitlab.com
    token: test-token
groups:
  - name: frontend
    resource: gl
    path: my-org/frontend
    local_path: ./frontend
    repos:
      - name: web-app
        url: https://gitlab.com/my-org/frontend/web-app.git
        path: my-org/frontend/web-app
      - name: ui-lib
        url: https://gitlab.com/my-org/frontend/ui-lib.git
        path: my-org/frontend/ui-lib
  - name: backend
    resource: gl
    path: my-org/backend
    local_path: ./backend
    repos:
      - name: web-api
        url: https://gitlab.com/my-org/backend/web-api.git
        path: my-org/backend/web-api
repos:
  - name: dotfiles
    resource: gl
    url: https://gitlab.com/me/dotfiles.git
    local_path: ./dotfiles
`
	configPath := filepath.Join(dir, ".grepom.yml")
	os.WriteFile(configPath, []byte(content), 0644)
	return configPath
}

func TestDirCommand_NoArgs_PrintsBase(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != dir {
		t.Errorf("expected base dir %q, got %q", dir, output)
	}
}

func TestDirCommand_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{"web-app"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	expected := filepath.Join(dir, "frontend", "web-app")
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestDirCommand_FuzzySingleResult(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{"dotfiles"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	expected := filepath.Join(dir, "dotfiles")
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestDirCommand_FuzzyCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{"WEB-APP"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	expected := filepath.Join(dir, "frontend", "web-app")
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestDirCommand_NoMatch_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{"nonexistent"})

	if err == nil {
		t.Fatal("expected error for no match")
	}
}

func TestDirCommand_MultipleMatch_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	// "web" 匹配 web-app 和 web-api
	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{"web"})

	if err == nil {
		t.Fatal("expected error for multiple matches")
	}
}

func TestDirCommand_SubstringMatch(t *testing.T) {
	dir := t.TempDir()
	configPath := createDirTestConfig(t, dir)

	// "ui" 匹配 ui-lib
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	configFile = configPath
	err := dirCmd.RunE(dirCmd, []string{"ui"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	expected := filepath.Join(dir, "frontend", "ui-lib")
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestDirCommand_UpwardConfigSearch(t *testing.T) {
	// 创建目录结构: parent/.grepom.yml → parent/child/grandchild
	parent := t.TempDir()
	configPath := createDirTestConfig(t, parent)

	grandchild := filepath.Join(parent, "child", "grandchild")
	os.MkdirAll(grandchild, 0755)

	// 在子目录中执行，configFile 使用空值（默认），触发向上查找
	oldDir, _ := os.Getwd()
	os.Chdir(grandchild)
	defer os.Chdir(oldDir)

	// 使用空 configFile，让 FindConfig 向上查找
	configFile = ""

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := dirCmd.RunE(dirCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd should find config upward: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != parent {
		t.Errorf("expected base dir %q, got %q", parent, output)
	}

	// 确认配置文件确实不在当前目录
	if _, err := os.Stat(".grepom.yml"); !os.IsNotExist(err) {
		t.Error("expected .grepom.yml to NOT exist in grandchild dir")
	}

	_ = configPath // suppress unused warning
}

func TestDirCommand_ShellPrintsFunction(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	dirShell = true
	defer func() { dirShell = false }()

	err := dirCmd.RunE(dirCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("dirCmd --shell failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "gcd()") {
		t.Errorf("expected gcd() function in output, got:\n%s", output)
	}
	if !strings.Contains(output, "grepom dir") {
		t.Errorf("expected 'grepom dir' reference in output, got:\n%s", output)
	}
}
