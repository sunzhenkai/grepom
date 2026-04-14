package provider

import (
	"context"
	"fmt"
)

// Repo represents a discovered remote repository.
type Repo struct {
	Name      string
	CloneURL  string
	SSHURL    string
	Path      string // local relative path (derived from group settings)
	Provider  string // "gitlab", "github", or "explicit"
	Resource  string // resource name this repo came from
	GroupName string // group name this repo belongs to (empty for standalone repos)
	Token     string // resolved token for clone (group/repo override or resource fallback)
	SSHKey    string // SSH key path for clone (group/repo override or resource fallback)

	// Source tracking for 6-level auth priority chain
	HasGroupToken  bool // true if token was set at group/repo level (override)
	HasGroupSSHKey bool // true if ssh_key was set at group/repo level (override)

	// Exclusion tracking
	DisabledReason string // "", "disabled" (resource/group/repo enabled:false), "excluded" (exclude_repos)
}

// GroupQuery specifies a remote group to discover repos from.
type GroupQuery struct {
	Path      string
	Recursive bool
}

// ListReposParams contains the parameters for listing repos from a provider.
type ListReposParams struct {
	ServerURL      string
	Token          string
	Groups         []GroupQuery
	Orgs           []string
	OrganizationID string // optional, used by Codeup provider
}

// RemoteGroup represents a remote group/org discovered from a provider API.
type RemoteGroup struct {
	Name     string // group/org 名称
	Path     string // 完整路径（GitLab: full_path, GitHub: login）
	Provider string // "gitlab" 或 "github"
}

// ListGroupsParams contains the parameters for listing groups from a provider.
type ListGroupsParams struct {
	ServerURL      string
	Token          string
	OrganizationID string // optional, used by Codeup provider
}

// Provider is the interface for remote repository providers.
type Provider interface {
	ListRepos(ctx context.Context, params ListReposParams) ([]Repo, error)
	ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error)
}

var registry = map[string]func() Provider{}

func Register(name string, factory func() Provider) {
	registry[name] = factory
}

func Get(name string) (Provider, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return factory(), nil
}

func AvailableProviders() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
