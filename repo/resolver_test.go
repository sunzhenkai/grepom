package repo

import (
	"path/filepath"
	"testing"

	"github.com/wii/grepom/config"
	"github.com/wii/grepom/provider"
)

func TestResolve_GroupRepos(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"work-gl": {Provider: "gitlab", URL: "gitlab.com", Token: "test"},
		},
		Groups: []config.Group{
			{
				Name:      "frontend",
				Resource:  "work-gl",
				Path:      "my-org/frontend",
				LocalPath: "./frontend",
				Repos: []config.GroupRepo{
					{Name: "app", URL: "https://gitlab.com/my-org/frontend/app.git", Path: "my-org/frontend/app"},
					{Name: "design-system", URL: "https://gitlab.com/my-org/frontend/ui/design-system.git", Path: "my-org/frontend/ui/design-system"},
				},
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}

	// Check first repo (direct child)
	r1 := repos[0]
	if r1.Name != "app" {
		t.Errorf("expected name 'app', got: %s", r1.Name)
	}
	expected1 := filepath.Join("/home/user/projects", "frontend", "app")
	if r1.Path != expected1 {
		t.Errorf("expected path %s, got: %s", expected1, r1.Path)
	}
	if r1.GroupName != "frontend" {
		t.Errorf("expected group 'frontend', got: %s", r1.GroupName)
	}
	if r1.Provider != "gitlab" {
		t.Errorf("expected provider 'gitlab', got: %s", r1.Provider)
	}
	// Verify URL derivation from host:port
	if r1.CloneURL != "https://gitlab.com/my-org/frontend/app.git" {
		t.Errorf("expected CloneURL https://gitlab.com/my-org/frontend/app.git, got: %s", r1.CloneURL)
	}
	if r1.SSHURL != "git@gitlab.com:my-org/frontend/app.git" {
		t.Errorf("expected SSHURL git@gitlab.com:my-org/frontend/app.git, got: %s", r1.SSHURL)
	}

	// Check second repo (subgroup child)
	r2 := repos[1]
	if r2.Name != "design-system" {
		t.Errorf("expected name 'design-system', got: %s", r2.Name)
	}
	expected2 := filepath.Join("/home/user/projects", "frontend", "ui", "design-system")
	if r2.Path != expected2 {
		t.Errorf("expected path %s, got: %s", expected2, r2.Path)
	}
}

func TestResolve_StandaloneRepos(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"github": {Provider: "github", URL: "github.com", Token: "test"},
		},
		Groups: []config.Group{},
		Repos: []config.Repo{
			{Name: "dotfiles", Resource: "github", URL: "https://github.com/me/dotfiles.git", LocalPath: "./dotfiles"},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	r := repos[0]
	if r.Name != "dotfiles" {
		t.Errorf("expected name 'dotfiles', got: %s", r.Name)
	}
	expected := filepath.Join("/home/user/projects", "dotfiles")
	if r.Path != expected {
		t.Errorf("expected path %s, got: %s", expected, r.Path)
	}
	if r.GroupName != "" {
		t.Errorf("standalone repo should have empty GroupName, got: %s", r.GroupName)
	}
	// Verify URL derivation
	if r.CloneURL != "https://github.com/me/dotfiles.git" {
		t.Errorf("expected CloneURL https://github.com/me/dotfiles.git, got: %s", r.CloneURL)
	}
	if r.SSHURL != "git@github.com:me/dotfiles.git" {
		t.Errorf("expected SSHURL git@github.com:me/dotfiles.git, got: %s", r.SSHURL)
	}
}

func TestResolveAndFilter_ByGroup(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gl": {Provider: "gitlab", URL: "gitlab.com", Token: "test"},
		},
		Groups: []config.Group{
			{
				Name: "frontend", Resource: "gl", Path: "org/frontend", LocalPath: "./frontend",
				Repos: []config.GroupRepo{
					{Name: "app", URL: "https://gitlab.com/org/frontend/app.git", Path: "org/frontend/app"},
				},
			},
			{
				Name: "backend", Resource: "gl", Path: "org/backend", LocalPath: "./backend",
				Repos: []config.GroupRepo{
					{Name: "api", URL: "https://gitlab.com/org/backend/api.git", Path: "org/backend/api"},
				},
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(Filter{Group: "frontend"})
	if err != nil {
		t.Fatalf("ResolveAndFilter failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "app" {
		t.Errorf("expected 'app', got: %s", repos[0].Name)
	}
}

func TestResolveAndFilter_ByResource(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gl":   {Provider: "gitlab", URL: "https://gitlab.com", Token: "test"},
			"ghub": {Provider: "github", URL: "github.com", Token: "test"},
		},
		Groups: []config.Group{
			{
				Name: "frontend", Resource: "gl", Path: "org/frontend", LocalPath: "./frontend",
				Repos: []config.GroupRepo{
					{Name: "app", URL: "https://gitlab.com/org/frontend/app.git", Path: "org/frontend/app"},
				},
			},
			{
				Name: "oss", Resource: "ghub", Path: "my-oss", LocalPath: "./oss",
				Repos: []config.GroupRepo{
					{Name: "tool", URL: "https://github.com/my-oss/tool.git", Path: "my-oss/tool"},
				},
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(Filter{Resource: "ghub"})
	if err != nil {
		t.Fatalf("ResolveAndFilter failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "tool" {
		t.Errorf("expected 'tool', got: %s", repos[0].Name)
	}
}

func TestApplyFilter_ByRepoName(t *testing.T) {
	repos := []provider.Repo{
		{Name: "app", GroupName: "frontend"},
		{Name: "api", GroupName: "backend"},
		{Name: "lib", GroupName: "frontend"},
	}

	filtered := ApplyFilter(repos, Filter{Name: "api"})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(filtered))
	}
	if filtered[0].Name != "api" {
		t.Errorf("expected 'api', got: %s", filtered[0].Name)
	}
}

func TestFullPath(t *testing.T) {
	r := provider.Repo{Path: "/home/user/projects/frontend/app"}
	result := FullPath("/home/user/projects", r)
	expected := filepath.Clean("/home/user/projects/frontend/app")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestDeriveSSHURL_FromHostAndPath(t *testing.T) {
	result := deriveSSHURL("org/repo", "gitlab.com")
	expected := "git@gitlab.com:org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestDeriveSSHURL_WithPort(t *testing.T) {
	result := deriveSSHURL("org/repo", "gitlab.com:8022")
	expected := "git@gitlab.com:8022:org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExtractRepoPath_HTTPS(t *testing.T) {
	result := extractRepoPath("https://gitlab.com/me/dotfiles.git")
	if result != "me/dotfiles" {
		t.Errorf("expected me/dotfiles, got %s", result)
	}
}

func TestExtractRepoPath_SSH(t *testing.T) {
	result := extractRepoPath("git@gitlab.com:me/dotfiles.git")
	if result != "me/dotfiles" {
		t.Errorf("expected me/dotfiles, got %s", result)
	}
}

func TestExtractRepoPath_PlainPath(t *testing.T) {
	result := extractRepoPath("me/dotfiles.git")
	if result != "me/dotfiles" {
		t.Errorf("expected me/dotfiles, got %s", result)
	}
}

func TestExtractRepoPath_HTTPSWithPort(t *testing.T) {
	result := extractRepoPath("https://gitlab.com:8443/me/dotfiles.git")
	if result != "me/dotfiles" {
		t.Errorf("expected me/dotfiles, got %s", result)
	}
}

// --- Auth merge tests ---

func TestResolve_GroupAuthOverridesResource(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gl": {Provider: "gitlab", URL: "https://gitlab.com", Token: "resource-token", SSHKey: "~/.ssh/id_resource"},
		},
		Groups: []config.Group{
			{
				Name: "frontend", Resource: "gl", Path: "org/frontend", LocalPath: "./frontend",
				Token:  "group-token",
				SSHKey: "~/.ssh/id_group",
				Repos: []config.GroupRepo{
					{Name: "app", URL: "https://gitlab.com/org/frontend/app.git", Path: "org/frontend/app"},
				},
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	r := repos[0]
	if r.Token != "group-token" {
		t.Errorf("expected group token 'group-token', got %s", r.Token)
	}
	if r.SSHKey != "~/.ssh/id_group" {
		t.Errorf("expected group SSH key, got %s", r.SSHKey)
	}
}

func TestResolve_ResourceFallback(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gl": {Provider: "gitlab", URL: "https://gitlab.com", Token: "resource-token", SSHKey: "~/.ssh/id_resource"},
		},
		Groups: []config.Group{
			{
				Name: "frontend", Resource: "gl", Path: "org/frontend", LocalPath: "./frontend",
				// No Token or SSHKey override
				Repos: []config.GroupRepo{
					{Name: "app", URL: "https://gitlab.com/org/frontend/app.git", Path: "org/frontend/app"},
				},
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	r := repos[0]
	if r.Token != "resource-token" {
		t.Errorf("expected resource token 'resource-token', got %s", r.Token)
	}
	if r.SSHKey != "~/.ssh/id_resource" {
		t.Errorf("expected resource SSH key, got %s", r.SSHKey)
	}
}

func TestResolve_StandaloneRepoAuthOverridesResource(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gh": {Provider: "github", URL: "github.com", Token: "resource-token", SSHKey: "~/.ssh/id_resource"},
		},
		Repos: []config.Repo{
			{
				Name:      "dotfiles",
				Resource:  "gh",
				URL:       "https://github.com/me/dotfiles.git",
				LocalPath: "./dotfiles",
				Token:     "repo-token",
				SSHKey:    "~/.ssh/id_repo",
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	r := repos[0]
	if r.Token != "repo-token" {
		t.Errorf("expected repo token 'repo-token', got %s", r.Token)
	}
	if r.SSHKey != "~/.ssh/id_repo" {
		t.Errorf("expected repo SSH key, got %s", r.SSHKey)
	}
}

func TestResolve_StandaloneRepoFallsBackToResource(t *testing.T) {
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gh": {Provider: "github", URL: "github.com", Token: "resource-token", SSHKey: "~/.ssh/id_resource"},
		},
		Repos: []config.Repo{
			{
				Name:      "dotfiles",
				Resource:  "gh",
				URL:       "https://github.com/me/dotfiles.git",
				LocalPath: "./dotfiles",
				// No Token or SSHKey override
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	r := repos[0]
	if r.Token != "resource-token" {
		t.Errorf("expected resource token, got %s", r.Token)
	}
	if r.SSHKey != "~/.ssh/id_resource" {
		t.Errorf("expected resource SSH key, got %s", r.SSHKey)
	}
}

func TestResolve_GroupPartialOverride(t *testing.T) {
	// Group overrides only SSH key, token falls back to resource
	cfg := &config.Config{
		Base: "/home/user/projects",
		Resources: map[string]config.Resource{
			"gl": {Provider: "gitlab", URL: "https://gitlab.com", Token: "resource-token", SSHKey: "~/.ssh/id_resource"},
		},
		Groups: []config.Group{
			{
				Name: "frontend", Resource: "gl", Path: "org/frontend", LocalPath: "./frontend",
				SSHKey: "~/.ssh/id_group", // only SSH key override
				Repos: []config.GroupRepo{
					{Name: "app", URL: "https://gitlab.com/org/frontend/app.git", Path: "org/frontend/app"},
				},
			},
		},
	}

	resolver := NewResolver(cfg)
	repos, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	r := repos[0]
	if r.Token != "resource-token" {
		t.Errorf("expected resource token fallback, got %s", r.Token)
	}
	if r.SSHKey != "~/.ssh/id_group" {
		t.Errorf("expected group SSH key override, got %s", r.SSHKey)
	}
}

// --- ApplySearchFilter tests ---

func TestApplySearchFilter_CaseInsensitive(t *testing.T) {
	repos := []provider.Repo{
		{Name: "Web-App", GroupName: "frontend"},
		{Name: "API-Server", GroupName: "backend"},
		{Name: "web-utils", GroupName: "shared"},
	}

	results := ApplySearchFilter(repos, "web", Filter{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestApplySearchFilter_NoMatch(t *testing.T) {
	repos := []provider.Repo{
		{Name: "frontend", GroupName: "fe"},
		{Name: "backend", GroupName: "be"},
	}

	results := ApplySearchFilter(repos, "xyz", Filter{})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestApplySearchFilter_EmptyKeyword(t *testing.T) {
	repos := []provider.Repo{
		{Name: "app"},
		{Name: "api"},
	}

	results := ApplySearchFilter(repos, "", Filter{})
	if len(results) != 2 {
		t.Fatalf("empty keyword should return all repos, got %d", len(results))
	}
}

func TestApplySearchFilter_WithGroupFilter(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", GroupName: "frontend"},
		{Name: "web-api", GroupName: "backend"},
		{Name: "web-utils", GroupName: "frontend"},
	}

	results := ApplySearchFilter(repos, "web", Filter{Group: "frontend"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results in frontend, got %d", len(results))
	}
	for _, r := range results {
		if r.GroupName != "frontend" {
			t.Errorf("expected group 'frontend', got %s", r.GroupName)
		}
	}
}

func TestApplySearchFilter_WithResourceFilter(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", Resource: "gitlab"},
		{Name: "web-api", Resource: "github"},
	}

	results := ApplySearchFilter(repos, "web", Filter{Resource: "github"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result from github, got %d", len(results))
	}
	if results[0].Name != "web-api" {
		t.Errorf("expected 'web-api', got %s", results[0].Name)
	}
}
