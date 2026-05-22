package config

import "testing"

func TestNormalizeRepoURL_HTTPS(t *testing.T) {
	got := NormalizeRepoURL("https://gitlab.com/my-org/infra/api-lib.git")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_HTTP(t *testing.T) {
	got := NormalizeRepoURL("http://gitlab.com/my-org/infra/api-lib.git")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_SSH(t *testing.T) {
	got := NormalizeRepoURL("git@gitlab.com:my-org/infra/api-lib.git")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_NoGitSuffix(t *testing.T) {
	got := NormalizeRepoURL("https://gitlab.com/my-org/infra/api-lib")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_WithPort(t *testing.T) {
	got := NormalizeRepoURL("https://gitlab.com:8443/my-org/infra/api-lib.git")
	want := "gitlab.com:8443/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_TrailingSlash(t *testing.T) {
	got := NormalizeRepoURL("https://gitlab.com/my-org/infra/api-lib/")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_HostCaseInsensitive(t *testing.T) {
	got := NormalizeRepoURL("https://GitLab.com/my-org/infra/api-lib.git")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_PathCaseSensitive(t *testing.T) {
	got := NormalizeRepoURL("https://gitlab.com/My-Org/Infra/api-lib.git")
	want := "gitlab.com/My-Org/Infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_Empty(t *testing.T) {
	got := NormalizeRepoURL("")
	want := ""
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_SSHNoGitSuffix(t *testing.T) {
	got := NormalizeRepoURL("git@gitlab.com:my-org/infra/api-lib")
	want := "gitlab.com/my-org/infra/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeRepoURL_HTTPSWithTrailingSlashAndGit(t *testing.T) {
	got := NormalizeRepoURL("https://gitlab.com/my-org/api-lib.git/")
	want := "gitlab.com/my-org/api-lib"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
