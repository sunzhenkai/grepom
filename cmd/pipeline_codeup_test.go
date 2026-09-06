package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wii/grepom/config"
)

// createCodeupPipelineTestConfig 构造含 codeup resource + repo 的测试配置。
func createCodeupPipelineTestConfig(t *testing.T, dir string, withOrgID bool) *config.Config {
	t.Helper()
	orgLine := ""
	if withOrgID {
		orgLine = "\n    organization_id: \"org-777\""
	}
	content := `
base: ` + dir + `
resources:
  cu:
    provider: codeup
    url: codeup.aliyun.com
    token: cu-token` + orgLine + `
groups:
  - name: solo
    resource: cu
    path: wii/solo
    local_path: ./solo
    repos:
      - name: grepom
        url: https://codeup.aliyun.com/wii/solo/grepom.git
        path: wii/solo/grepom
`
	configPath := filepath.Join(dir, ".grepom.yml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	config.ResolveBasePath(cfg, dir)
	return cfg
}

func TestResolvePipelineInput_CodeupPassesOrgID(t *testing.T) {
	cfg := createCodeupPipelineTestConfig(t, t.TempDir(), true)

	provider, serverURL, remotePath, token, orgID, err := resolvePipelineInput(cfg, "grepom")
	if err != nil {
		t.Fatalf("resolvePipelineInput: %v", err)
	}

	// provider 应已注册（cicd.Get("codeup") 成功）
	if provider == nil {
		t.Fatal("provider is nil: codeup cicd provider 未注册")
	}
	if serverURL != "https://codeup.aliyun.com" {
		t.Errorf("serverURL = %q", serverURL)
	}
	if remotePath != "wii/solo/grepom" {
		t.Errorf("remotePath = %q", remotePath)
	}
	if token != "cu-token" {
		t.Errorf("token = %q", token)
	}
	if orgID != "org-777" {
		t.Errorf("organizationID = %q, want org-777", orgID)
	}
}

func TestResolvePipelineInput_CodeupMissingOrgIDConfigRejected(t *testing.T) {
	// config 校验层面：codeup resource 缺 organization_id 时 Load 即报错
	dir := t.TempDir()
	content := `
base: ` + dir + `
resources:
  cu:
    provider: codeup
    url: codeup.aliyun.com
    token: cu-token
`
	configPath := filepath.Join(dir, ".grepom.yml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(configPath)
	if err == nil {
		t.Fatal("expected config validation error")
	}
	if !strings.Contains(err.Error(), "organization_id") {
		t.Errorf("err = %v", err)
	}
}
