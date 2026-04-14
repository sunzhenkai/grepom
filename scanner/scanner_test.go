package scanner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo 创建一个包含模拟敏感信息的临时 git 仓库。
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// git init
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.email", "test@test.com").Run(); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.name", "Test").Run(); err != nil {
		t.Fatalf("git config name: %v", err)
	}

	// 创建包含 SSH 私钥的文件（gitleaks private-key 规则能可靠检出）
	keyFile := filepath.Join(dir, "id_rsa")
	keyContent := "-----BEGIN RSA PRIVATE KEY-----\n"
	keyContent += "MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7mSrORbsXxHNHYtML\n"
	keyContent += "qWPQQe0XMs4RkhdVtPYlY8zVSUDAJdyhHL2spKagpKvX7eBZLii3OQNz7PdmPhD\n"
	keyContent += "-----END RSA PRIVATE KEY-----\n"
	if err := os.WriteFile(keyFile, []byte(keyContent), 0644); err != nil {
		t.Fatalf("write id_rsa: %v", err)
	}

	// 创建一个安全文件（不应触发任何规则）
	safeFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(safeFile, []byte("# Hello World\nThis is a safe file.\n"), 0644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	// git commit
	if err := exec.Command("git", "-C", dir, "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", "initial commit").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	return dir
}

func TestScanDir(t *testing.T) {
	dir := setupTestRepo(t)

	s := NewScanner(Options{})
	findings, err := s.ScanDir(context.Background(), dir)
	if err != nil {
		t.Fatalf("ScanDir error: %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("ScanDir 应该发现至少一个敏感信息（SSH 私钥），但返回了 0 个")
	}

	// 验证至少发现了 private-key
	foundPrivateKey := false
	for _, f := range findings {
		if f.RuleID == "private-key" || f.RuleID == "ssh-private-key" {
			foundPrivateKey = true
		}
	}
	if !foundPrivateKey {
		t.Errorf("期望发现 private-key，但没有找到")
		t.Logf("发现的 findings:")
		for _, f := range findings {
			t.Logf("  RuleID=%s File=%s Line=%d Secret=%s", f.RuleID, f.File, f.Line, MaskSecret(f.Secret))
		}
	}

	// 验证 File 字段不是空的
	for _, f := range findings {
		if f.File == "" {
			t.Error("Finding.File 不应为空")
		}
	}
}

func TestScanGitHistory(t *testing.T) {
	dir := setupTestRepo(t)

	s := NewScanner(Options{})
	findings, err := s.ScanGitHistory(context.Background(), dir)
	if err != nil {
		t.Fatalf("ScanGitHistory error: %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("ScanGitHistory 应该发现至少一个敏感信息（SSH 私钥），但返回了 0 个")
	}

	// 验证至少发现了 private-key
	foundPrivateKey := false
	for _, f := range findings {
		if f.RuleID == "private-key" || f.RuleID == "ssh-private-key" {
			foundPrivateKey = true
		}
	}
	if !foundPrivateKey {
		t.Errorf("期望在 git 历史中发现 private-key，但没有找到")
	}
}

func TestScanDirWithGitignore(t *testing.T) {
	dir := t.TempDir()

	// git init
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.email", "test@test.com").Run(); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.name", "Test").Run(); err != nil {
		t.Fatalf("git config name: %v", err)
	}

	// 创建 .gitignore 排除 secrets/ 目录
	gitignoreContent := "secrets/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	// 创建一个应被发现的文件（SSH 私钥）
	keyFile := filepath.Join(dir, "id_rsa")
	keyContent := "-----BEGIN RSA PRIVATE KEY-----\n"
	keyContent += "MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7mSrORbsXxHNHYtML\n"
	keyContent += "-----END RSA PRIVATE KEY-----\n"
	if err := os.WriteFile(keyFile, []byte(keyContent), 0644); err != nil {
		t.Fatalf("write id_rsa: %v", err)
	}

	// 创建一个在 .gitignore 排除目录中的文件（也应包含敏感信息）
	secretsDir := filepath.Join(dir, "secrets")
	if err := os.MkdirAll(secretsDir, 0755); err != nil {
		t.Fatalf("mkdir secrets: %v", err)
	}
	ignoredFile := filepath.Join(secretsDir, "private.key")
	ignoredContent := "-----BEGIN RSA PRIVATE KEY-----\n"
	ignoredContent += "MIIEpAIBAAKCAQEA2Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7mSrORbsXxHNHYtML\n"
	ignoredContent += "-----END RSA PRIVATE KEY-----\n"
	if err := os.WriteFile(ignoredFile, []byte(ignoredContent), 0644); err != nil {
		t.Fatalf("write private.key: %v", err)
	}

	// git commit
	if err := exec.Command("git", "-C", dir, "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", "initial").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	s := NewScanner(Options{})
	findings, err := s.ScanDir(context.Background(), dir)
	if err != nil {
		t.Fatalf("ScanDir error: %v", err)
	}

	// 验证没有发现 secrets/ 目录下的文件
	for _, f := range findings {
		if containsStr(f.File, "secrets/") || containsStr(f.File, "secrets\\") {
			t.Errorf("不应该扫描 .gitignore 排除的文件: %s", f.File)
		}
	}
}

func TestScanDirWithGitleaksignore(t *testing.T) {
	dir := t.TempDir()

	// git init
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.email", "test@test.com").Run(); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.name", "Test").Run(); err != nil {
		t.Fatalf("git config name: %v", err)
	}

	// 创建包含 SSH 私钥的文件
	keyFile := filepath.Join(dir, "id_rsa")
	keyContent := "-----BEGIN RSA PRIVATE KEY-----\n"
	keyContent += "MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7mSrORbsXxHNHYtML\n"
	keyContent += "-----END RSA PRIVATE KEY-----\n"
	if err := os.WriteFile(keyFile, []byte(keyContent), 0644); err != nil {
		t.Fatalf("write id_rsa: %v", err)
	}

	// git commit
	if err := exec.Command("git", "-C", dir, "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", "initial").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// 先扫描获取 fingerprint
	s := NewScanner(Options{})
	findings, err := s.ScanDir(context.Background(), dir)
	if err != nil {
		t.Fatalf("ScanDir error: %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("初始扫描应该发现至少一个敏感信息")
	}

	// 记录第一个发现的 fingerprint（格式：file:rule:line）
	firstFinding := findings[0]
	fingerprint := firstFinding.File + ":" + firstFinding.RuleID + ":" + intToStr(firstFinding.Line)

	// 创建 .gitleaksignore 忽略该发现
	ignoreContent := fingerprint + "\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitleaksignore"), []byte(ignoreContent), 0644); err != nil {
		t.Fatalf("write .gitleaksignore: %v", err)
	}

	// 重新扫描（需要新的 Scanner/Detector 实例来重新加载 ignore）
	s2 := NewScanner(Options{})
	findings2, err := s2.ScanDir(context.Background(), dir)
	if err != nil {
		t.Fatalf("ScanDir with ignore error: %v", err)
	}

	// 验证被忽略的发现不再出现
	for _, f := range findings2 {
		if f.File == firstFinding.File && f.Line == firstFinding.Line && f.RuleID == firstFinding.RuleID {
			t.Errorf("发现应被 .gitleaksignore 过滤的项: %s:%d (%s)", f.File, f.Line, f.RuleID)
		}
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
