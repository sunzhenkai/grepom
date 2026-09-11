package scanner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixture 在 dir 下写一个待扫描文件。
func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// gitCommit 把 dir 变成 git 仓库并提交全部文件。
// 生产扫描对象都是 git 克隆;gitleaks 的 Files 源在测试进程里扫裸目录
// 会静默零检出(二进制模式不复现),故统一按仓库形态构造夹具。
func gitCommit(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args[1:], err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	run("add", ".")
	run("commit", "-m", "fixture")
}

// scanFindings 扫描 dir 并按文件名归集 RuleID。
func scanFindings(t *testing.T, dir string) map[string][]string {
	t.Helper()
	s := NewScanner(Options{})
	findings, err := s.ScanDir(context.Background(), dir)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	byFile := map[string][]string{}
	for _, f := range findings {
		byFile[filepath.Base(f.File)] = append(byFile[filepath.Base(f.File)], f.RuleID)
	}
	return byFile
}

// TestAllowlistSuppressesKnownFalsePositives 内置放行层必须吞掉这些已确认误报,
// 样本与 scanner/gitleaks.toml 注释一一对应;新增放行规则时在此补对应样本。
func TestAllowlistSuppressesKnownFalsePositives(t *testing.T) {
	dir := t.TempDir()
	fixtures := map[string]string{
		// RSA 公钥体(X.509 SubjectPublicKeyInfo)——公钥不是秘密
		"public_key.txt": `publicKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCOhQ81fJEKI/b0+iKtK0/WN7wcMJKuGL5Qhzu0WvGvh` + strings.Repeat("Aa9Z", 30) + `"`,
		// 文档示例:纯顺序数字 sk- 占位
		"doc_example.md": "api_key = sk-1234567890abcdef\n",
		// 测试夹具自明假 key
		"client_test.go": `cfg := "sk-integration-test-key-12345678"`,
		// README 示例读写 token
		"readme_curl.md": "curl -H \"Authorization: Bearer write123\" --upload-file f.txt http://localhost:9200/user\n",
		// PEM 格式名字面量
		"rsa.js": "var key = NodeRSA(privateKey, 'pkcs8-private-pem');\n",
		// 完整性指纹 JSON 字段(值为 sha256,非秘密)
		"fingerprint.json": "{\"sha256_fingerprint\": \"" + strings.Repeat("3f79bb7b435b05321651daefd374cdc6", 2) + "\",\n \"fingerprint\": \"" + strings.Repeat("7e0e9d1d1a2b3c4d5e6f0a1b2c3d4e5f", 2) + "\"}\n",
		// k8s 证书哈希行
		"k8s_install.md": "  --discovery-token-ca-cert-hash sha256:" + strings.Repeat("a1b2c3d4", 8) + "\n",
		// MATLAB jsonlab 取值语句
		"jsonopt.m": "function value = jsonopt(key, val, varargin)\n  v = getfield(json, key);\n",
		// sage 本地烟测 dev 夹具(base64 解码为可读英文)
		"smoke.mjs": "process.env.SAGE_SECRET_MASTER_KEY = 'c2FnZS1sb2NhbC1kZXYtc2VjcmV0LW1hc3Rlci1reSE=';\n",
		// 测试夹具值级放行样本
		"dotfiles_test.py": "token = \"actual-env-secret-123\"\n",
		"sage_admission.ts": "invocation: { idempotencyKey: 'admission-key-1' }\n",
	}
	for name, content := range fixtures {
		writeFixture(t, dir, name, content)
	}
	gitCommit(t, dir)

	byFile := scanFindings(t, dir)
	for name := range fixtures {
		if rules, ok := byFile[name]; ok {
			t.Errorf("误报未被放行: %s => %v", name, rules)
		}
	}
}

// TestAllowlistDoesNotWeakenDetection 放行层不得削弱默认规则的真实检测能力:
// 真私钥(PEM)与高熵裸 token 仍必须上报。
func TestAllowlistDoesNotWeakenDetection(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "id_rsa",
		"-----BEGIN RSA PRIVATE KEY-----\n"+
			"MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7mSrORbsXxHNHYtML\n"+
			"qWPQQe0XMs4RkhdVtPYlY8zVSUDAJdyhHL2spKagpKvX7eBZLii3OQNz7PdmPhD\n"+
			"-----END RSA PRIVATE KEY-----\n")
	writeFixture(t, dir, "config.py", "api_key = \"gsk_9fX2mQ7vT4wZ8bN1cJ6hR5tY3uE0oP2aS4dF7gL1kIj\"\n")
	gitCommit(t, dir)

	byFile := scanFindings(t, dir)
	if rules, ok := byFile["id_rsa"]; !ok {
		t.Error("真私钥 PEM 未被检出——放行层误伤 private-key 规则")
	} else {
		t.Logf("id_rsa => %v", rules)
	}
	if rules, ok := byFile["config.py"]; !ok {
		t.Error("高熵 api_key 未被检出——放行层误伤 generic-api-key 规则")
	} else {
		t.Logf("config.py => %v", rules)
	}
}
