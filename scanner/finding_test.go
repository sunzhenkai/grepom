package scanner

import (
	"testing"
)

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		maxLen int
		want   string
	}{
		{
			name:   "短路径不截断",
			path:   "config/settings.json",
			maxLen: 40,
			want:   "config/settings.json",
		},
		{
			name:   "长路径截断保留文件名",
			path:   "repos/github/sunzhenkai/notes/computer science/cryptography/rsa.md",
			maxLen: 40,
			want:   "repos/github/sunzhenkai/notes/.../rsa.md",
		},
		{
			name:   "刚好等于 maxLen",
			path:   "a/b/c/d/e.txt",
			maxLen: 14,
			want:   "a/b/c/d/e.txt",
		},
		{
			name:   "无路径分隔符的长路径",
			path:   "verylongfilenamethatiswaytoolong.txt",
			maxLen: 20,
			want:   "verylongfilenamet...",
		},
		{
			name:   "文件名超长时截断整体",
			path:   "short/verylongfilenamethatiswaywaywaytoolong.txt",
			maxLen: 20,
			want:   "short/verylongfil...",
		},
		{
			name:   "maxLen 为 0 时使用默认 40",
			path:   "config/app.yml",
			maxLen: 0,
			want:   "config/app.yml",
		},
		{
			name:   "空路径",
			path:   "",
			maxLen: 40,
			want:   "",
		},
		{
			name:   "单层路径不截断",
			path:   "README.md",
			maxLen: 40,
			want:   "README.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncatePath(tt.path, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncatePath(%q, %d) = %q, want %q", tt.path, tt.maxLen, got, tt.want)
			}
			// 验证截断后长度不超过 maxLen（有意义的场景）
			if tt.maxLen > 0 && len(tt.path) > tt.maxLen && len(got) > tt.maxLen {
				t.Errorf("TruncatePath result length %d exceeds maxLen %d: %q", len(got), tt.maxLen, got)
			}
		})
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		want   string
	}{
		{
			name:   "长 secret 显示前 8 字符",
			secret: "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
			want:   "ghp_ABCD...",
		},
		{
			name:   "刚好 8 字符",
			secret: "12345678",
			want:   "1234...",
		},
		{
			name:   "短 secret 3 字符",
			secret: "abc",
			want:   "a...",
		},
		{
			name:   "短 secret 5 字符",
			secret: "abcde",
			want:   "ab...",
		},
		{
			name:   "空字符串",
			secret: "",
			want:   "",
		},
		{
			name:   "带前后空白",
			secret: "  secret-value-here  ",
			want:   "secret-v...",
		},
		{
			name:   "AK/SK 类型",
			secret: "AKIAIOSFODNN7EXAMPLE",
			want:   "AKIAIOSF...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSecret(tt.secret)
			if got != tt.want {
				t.Errorf("MaskSecret(%q) = %q, want %q", tt.secret, got, tt.want)
			}
		})
	}
}

func TestFindingFields(t *testing.T) {
	f := Finding{
		Repo:        "web-app",
		File:        "config/settings.yml",
		Line:        42,
		RuleID:      "generic-api-key",
		Description: "Generic API Key",
		Secret:      "super-secret-key-12345",
		Severity:    SeverityHigh,
	}

	if f.Repo != "web-app" {
		t.Errorf("Repo = %q, want %q", f.Repo, "web-app")
	}
	if f.File != "config/settings.yml" {
		t.Errorf("File = %q, want %q", f.File, "config/settings.yml")
	}
	if f.Line != 42 {
		t.Errorf("Line = %d, want %d", f.Line, 42)
	}
	if f.RuleID != "generic-api-key" {
		t.Errorf("RuleID = %q, want %q", f.RuleID, "generic-api-key")
	}
	if f.Severity != SeverityHigh {
		t.Errorf("Severity = %q, want %q", f.Severity, SeverityHigh)
	}
}

func TestSeverityConstants(t *testing.T) {
	if SeverityCritical != "CRITICAL" {
		t.Errorf("SeverityCritical = %q, want %q", SeverityCritical, "CRITICAL")
	}
	if SeverityHigh != "HIGH" {
		t.Errorf("SeverityHigh = %q, want %q", SeverityHigh, "HIGH")
	}
	if SeverityMedium != "MEDIUM" {
		t.Errorf("SeverityMedium = %q, want %q", SeverityMedium, "MEDIUM")
	}
	if SeverityLow != "LOW" {
		t.Errorf("SeverityLow = %q, want %q", SeverityLow, "LOW")
	}
}

func TestFindingsToJSON(t *testing.T) {
	findings := []Finding{
		{
			Repo:        "test-repo",
			File:        ".env",
			Line:        5,
			RuleID:      "generic-api-key",
			Description: "Generic API Key",
			Secret:      "sk-abc123",
			Severity:    SeverityHigh,
		},
	}

	data, err := FindingsToJSON(findings)
	if err != nil {
		t.Fatalf("FindingsToJSON error: %v", err)
	}

	// 验证 JSON 包含预期字段
	jsonStr := string(data)
	expectedFields := []string{`"repo"`, `"test-repo"`, `"file"`, `".env"`, `"rule_id"`, `"generic-api-key"`, `"severity"`, `"HIGH"`}
	for _, field := range expectedFields {
		if !containsString(jsonStr, field) {
			t.Errorf("JSON output missing expected field: %q\nGot: %s", field, jsonStr)
		}
	}
}

func containsString(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
