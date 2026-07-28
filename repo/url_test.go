package repo

import (
	"testing"

	"github.com/wii/grepom/config"
)

func TestClassifyRepoURL(t *testing.T) {
	cases := []struct {
		in   string
		want URLKind
	}{
		{"org/app.git", URLRelative},
		{"https://git.example.com/org/app.git", URLAbsoluteHTTP},
		{"http://git.example.com/org/app.git", URLAbsoluteHTTP},
		{"git@git.example.com:org/app.git", URLAbsoluteSSH},
		{"ssh://git@git.example.com/org/app.git", URLAbsoluteSSH},
	}
	for _, tc := range cases {
		if got := ClassifyRepoURL(tc.in); got != tc.want {
			t.Errorf("ClassifyRepoURL(%q)=%v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestResolveBoundRepoURLs_Relative(t *testing.T) {
	res := config.Resource{URL: "git.example.com", Provider: "generic"}
	got := ResolveBoundRepoURLs("tools/internal-tool.git", res)
	if got.CloneURL != "https://git.example.com/tools/internal-tool.git" {
		t.Errorf("CloneURL=%s", got.CloneURL)
	}
	if got.SSHURL != "git@git.example.com:tools/internal-tool.git" {
		t.Errorf("SSHURL=%s", got.SSHURL)
	}
	if got.PreferHTTPS {
		t.Error("relative path should prefer SSH")
	}
}

func TestResolveBoundRepoURLs_AbsoluteHTTPS(t *testing.T) {
	res := config.Resource{URL: "other.example.com", Provider: "generic"}
	raw := "https://git.example.com/tools/internal-tool.git"
	got := ResolveBoundRepoURLs(raw, res)
	if got.CloneURL != raw {
		t.Errorf("CloneURL should be unchanged, got %s", got.CloneURL)
	}
	if got.SSHURL != "git@git.example.com:tools/internal-tool.git" {
		t.Errorf("SSHURL=%s", got.SSHURL)
	}
	if !got.PreferHTTPS {
		t.Error("absolute HTTPS should PreferHTTPS")
	}
}

func TestResolveBoundRepoURLs_AbsoluteGitAt(t *testing.T) {
	res := config.Resource{URL: "other.example.com", Provider: "generic"}
	raw := "git@git.example.com:tools/internal-tool.git"
	got := ResolveBoundRepoURLs(raw, res)
	if got.SSHURL != raw {
		t.Errorf("SSHURL should be unchanged, got %s", got.SSHURL)
	}
	if got.PreferHTTPS {
		t.Error("absolute SSH should not PreferHTTPS")
	}
	if got.CloneURL != "https://git.example.com/tools/internal-tool.git" {
		t.Errorf("CloneURL=%s", got.CloneURL)
	}
}

func TestResolveBoundRepoURLs_AbsoluteSSHScheme(t *testing.T) {
	res := config.Resource{URL: "other.example.com", Provider: "generic"}
	raw := "ssh://git@git.example.com/tools/internal-tool.git"
	got := ResolveBoundRepoURLs(raw, res)
	if got.SSHURL != raw {
		t.Errorf("SSHURL should be unchanged, got %s", got.SSHURL)
	}
	if stringsHasPrefix(got.SSHURL, "git@other.example.com:") {
		t.Errorf("must not re-concatenate onto resource host: %s", got.SSHURL)
	}
}

func TestExtractRemotePath_SSHScheme(t *testing.T) {
	got := ExtractRemotePath("ssh://git@git.example.com/tools/internal-tool.git")
	if got != "tools/internal-tool" {
		t.Errorf("got %q", got)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
