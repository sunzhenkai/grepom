package cmd

import (
	"os"
	"testing"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		// HTTPS with oauth2 token
		{"HTTPS with oauth2 token", "https://oauth2:glpat-xxx@gitlab.company.com/myorg/repo.git", "gitlab.company.com"},
		// HTTPS with username:password
		{"HTTPS with username:password", "https://user:pass@host.example.com/org/repo.git", "host.example.com"},
		// HTTPS with username only (token)
		{"HTTPS with username only", "https://token@github.com/user/repo.git", "github.com"},
		// HTTPS without userinfo (backward compat)
		{"HTTPS without userinfo", "https://gitlab.com/myorg/repo.git", "gitlab.com"},
		// HTTP without userinfo
		{"HTTP without userinfo", "http://gitlab.company.com:8080/org/repo.git", "gitlab.company.com:8080"},
		// HTTPS without userinfo, no path
		{"HTTPS no path", "https://github.com", "github.com"},
		// SSH via ssh:// with username
		{"ssh:// with username", "ssh://git@gitlab.company.com:2222/org/repo.git", "gitlab.company.com:2222"},
		// SSH via ssh:// without username
		{"ssh:// without username", "ssh://gitlab.company.com/org/repo.git", "gitlab.company.com"},
		// SCP style git@host:path
		{"git@ SCP style", "git@github.com:user/repo.git", "github.com"},
		// SCP style self-hosted git@
		{"git@ self-hosted SCP", "git@gitlab.mycompany.com:org/repo.git", "gitlab.mycompany.com"},
		// HTTPS with @ in path (not userinfo) - @ comes after first /
		{"HTTPS @ in path (not userinfo)", "https://example.com/org/repo@v1.git", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHost(tt.url)
			if got != tt.expected {
				t.Errorf("extractHost(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

// --- detectProvider Codeup 用例 ---

// writeTempConfig 在临时目录写入 .grepom.yml 并返回路径。
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/.grepom.yml"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDetectProvider_CodeupFromConfig(t *testing.T) {
	t.Setenv("CODEUP_TOKEN_TEST", "tok-abc")
	cfgPath := writeTempConfig(t, `
base: ~/projects
resources:
  my-codeup:
    provider: codeup
    url: codeup.aliyun.com
    token: ${CODEUP_TOKEN_TEST}
    organization_id: "org-999"
`)

	old := configFile
	configFile = cfgPath
	t.Cleanup(func() { configFile = old })

	provider, serverURL, token, orgID, err := detectProvider("https://codeup.aliyun.com/wii/solo/grepom.git")
	if err != nil {
		t.Fatalf("detectProvider: %v", err)
	}
	if provider != "codeup" {
		t.Errorf("provider = %q", provider)
	}
	if serverURL != "https://codeup.aliyun.com" {
		t.Errorf("serverURL = %q", serverURL)
	}
	if token != "tok-abc" {
		t.Errorf("token = %q", token)
	}
	if orgID != "org-999" {
		t.Errorf("organizationID = %q, want org-999", orgID)
	}
}

func TestDetectProvider_CodeupKnownDomainEnvToken(t *testing.T) {
	// 无 config 命中（configFile 指向不存在的路径时 tryLoadConfig 失败，走知名域名分支）
	t.Setenv("GREPOM_CODEUP_TOKEN", "tok-env")

	old := configFile
	configFile = t.TempDir() + "/nonexistent.yml"
	t.Cleanup(func() { configFile = old })

	// 知名域名分支：token 来自环境变量，organizationID 为空
	provider, _, token, orgID, err := detectProvider("https://codeup.aliyun.com/wii/solo/grepom.git")
	if err != nil {
		t.Fatalf("detectProvider: %v", err)
	}
	if provider != "codeup" {
		t.Errorf("provider = %q", provider)
	}
	if token != "tok-env" {
		t.Errorf("token = %q, want tok-env (GREPOM_CODEUP_TOKEN)", token)
	}
	if orgID != "" {
		t.Errorf("organizationID = %q, want empty (no config resource)", orgID)
	}
}

func TestDetectProvider_CodeupMissingOrgIDMessage(t *testing.T) {
	// 命中 config 中 codeup resource 但无 organization_id：config.Load 校验即报错
	cfgPath := writeTempConfig(t, `
resources:
  my-codeup:
    provider: codeup
    url: codeup.aliyun.com
    token: plain-tok
`)

	old := configFile
	configFile = cfgPath
	t.Cleanup(func() { configFile = old })

	// tryLoadConfig 失败（config 校验强制要求 organization_id），回落知名域名分支
	provider, _, _, orgID, err := detectProvider("https://codeup.aliyun.com/wii/solo/grepom.git")
	if err != nil {
		t.Fatalf("detectProvider: %v", err)
	}
	if provider != "codeup" {
		t.Errorf("provider = %q", provider)
	}
	if orgID != "" {
		t.Errorf("organizationID = %q, want empty", orgID)
	}
}
