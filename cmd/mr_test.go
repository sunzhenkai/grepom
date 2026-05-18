package cmd

import "testing"

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
