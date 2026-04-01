package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// envVarPattern matches ${VAR_NAME} placeholder syntax in token fields
var envVarPattern = regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)}$`)

// Config represents the top-level configuration file structure.
type Config struct {
	Base           string              `yaml:"base"`
	Resources      map[string]Resource `yaml:"resources"`
	Groups         []Group             `yaml:"groups"`
	Repos          []Repo              `yaml:"repos"`
	rawTokens      map[string]string   // resource name → original token string (with ${...} placeholders)
	rawGroupTokens map[int]string      // group index → original group token string
	rawRepoTokens  map[int]string      // repo index → original repo token string
}

// Resource defines authentication and connection information for a remote provider.
type Resource struct {
	Provider string `yaml:"provider"`
	URL      string `yaml:"url"`
	Token    string `yaml:"token"`             // resolved at runtime; raw placeholder stored in Config.rawTokens
	SSHKey   string `yaml:"ssh_key,omitempty"` // optional SSH key path for clone authentication
}

// Group defines a remote group (GitLab group or GitHub org) whose repos are managed together.
type Group struct {
	Name      string      `yaml:"name"`
	Resource  string      `yaml:"resource"`
	Path      string      `yaml:"path"`
	LocalPath string      `yaml:"local_path,omitempty"`
	Recursive bool        `yaml:"recursive,omitempty"`
	SSHKey    string      `yaml:"ssh_key,omitempty"` // optional, overrides resource SSH key for clone
	Token     string      `yaml:"token,omitempty"`   // optional, overrides resource token for clone (supports ${ENV_VAR})
	Repos     []GroupRepo `yaml:"repos,omitempty"`
}

// GroupRepo is a repo entry within a group. Local path is auto-derived from group settings.
type GroupRepo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Path string `yaml:"path"` // full remote path (e.g. my-org/frontend/ui/design-system)
}

// Repo is a standalone repo entry not belonging to any group.
type Repo struct {
	Name      string `yaml:"name"`
	Resource  string `yaml:"resource"`
	URL       string `yaml:"url"`
	LocalPath string `yaml:"local_path,omitempty"` // defaults to ./<name> if empty
	Token     string `yaml:"token,omitempty"`      // optional, overrides resource token for clone (supports ${ENV_VAR})
	SSHKey    string `yaml:"ssh_key,omitempty"`    // optional, overrides resource SSH key for clone
}

// Load reads and parses a configuration file, resolves token placeholders,
// expands tildes in base path, and validates the configuration.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Initialize nil maps/slices
	if cfg.Resources == nil {
		cfg.Resources = make(map[string]Resource)
	}
	if cfg.Groups == nil {
		cfg.Groups = []Group{}
	}
	if cfg.Repos == nil {
		cfg.Repos = []Repo{}
	}

	// Save raw token values and resolve environment variable placeholders
	cfg.rawTokens = make(map[string]string)
	for name, res := range cfg.Resources {
		raw := res.Token
		cfg.rawTokens[name] = raw

		resolved, err := resolveToken(raw)
		if err != nil {
			return nil, fmt.Errorf("config: resource %q: %w", name, err)
		}
		res.Token = resolved
		cfg.Resources[name] = res
	}

	// Resolve Group tokens (supports ${ENV_VAR} syntax)
	cfg.rawGroupTokens = make(map[int]string)
	for i, g := range cfg.Groups {
		if g.Token != "" {
			cfg.rawGroupTokens[i] = g.Token
			resolved, err := resolveToken(g.Token)
			if err != nil {
				return nil, fmt.Errorf("config: group %q: %w", g.Name, err)
			}
			g.Token = resolved
			cfg.Groups[i] = g
		}
	}

	// Resolve Repo tokens (supports ${ENV_VAR} syntax)
	cfg.rawRepoTokens = make(map[int]string)
	for i, r := range cfg.Repos {
		if r.Token != "" {
			cfg.rawRepoTokens[i] = r.Token
			resolved, err := resolveToken(r.Token)
			if err != nil {
				return nil, fmt.Errorf("config: repo %q: %w", r.Name, err)
			}
			r.Token = resolved
			cfg.Repos[i] = r
		}
	}

	// Expand tilde in base path
	cfg.Base = expandTilde(cfg.Base)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// resolveToken checks if the token is an environment variable placeholder (${VAR})
// and resolves it to the actual value. Returns the original value if not a placeholder.
func resolveToken(token string) (string, error) {
	if token == "" {
		return "", nil
	}

	matches := envVarPattern.FindStringSubmatch(token)
	if matches == nil {
		return token, nil
	}

	envName := matches[1]
	value, ok := os.LookupEnv(envName)
	if !ok {
		return "", fmt.Errorf("token: environment variable %s is not set", envName)
	}
	return value, nil
}

// FindConfig locates the configuration file.
func FindConfig(explicitPath string) (string, error) {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err != nil {
			return "", fmt.Errorf("config file not found: %s", explicitPath)
		}
		return explicitPath, nil
	}

	if _, err := os.Stat(".grepom.yml"); err == nil {
		return ".grepom.yml", nil
	}

	return "", fmt.Errorf("no config file found. Use -c to specify a config file or create .grepom.yml in current directory")
}

// FindResource finds a resource by name.
func (c *Config) FindResource(name string) (*Resource, error) {
	res, ok := c.Resources[name]
	if !ok {
		return nil, fmt.Errorf("resource %q not found", name)
	}
	return &res, nil
}

// FindGroup finds a group by name. It first tries exact name match.
func (c *Config) FindGroup(name string) (int, *Group, error) {
	for i, g := range c.Groups {
		if g.Name == name {
			return i, &c.Groups[i], nil
		}
	}
	return -1, nil, fmt.Errorf("group %q not found", name)
}

// validate checks the configuration for required fields and consistency.
func (c *Config) validate() error {
	if c.Base == "" {
		return fmt.Errorf("config: 'base' field is required")
	}

	// Validate resources
	for name, res := range c.Resources {
		if res.Provider == "" {
			return fmt.Errorf("config: resource %q: 'provider' field is required", name)
		}
		if res.Provider != "gitlab" && res.Provider != "github" {
			return fmt.Errorf("config: resource %q: unsupported provider %q (use 'gitlab' or 'github')", name, res.Provider)
		}
		if res.URL == "" {
			return fmt.Errorf("config: resource %q: 'url' field is required", name)
		}
		if !strings.HasPrefix(res.URL, "http://") && !strings.HasPrefix(res.URL, "https://") {
			updated := res
			updated.URL = "https://" + res.URL
			c.Resources[name] = updated
		}
	}

	// Validate group names are unique
	groupNames := make(map[string]bool)
	for i, g := range c.Groups {
		if g.Name == "" {
			return fmt.Errorf("config: groups[%d]: 'name' field is required", i)
		}
		if groupNames[g.Name] {
			return fmt.Errorf("config: duplicate group name %q", g.Name)
		}
		groupNames[g.Name] = true

		if g.Resource == "" {
			return fmt.Errorf("config: group %q: 'resource' field is required", g.Name)
		}
		if _, ok := c.Resources[g.Resource]; !ok {
			return fmt.Errorf("config: group %q: resource %q not found", g.Name, g.Resource)
		}
		if g.Path == "" {
			return fmt.Errorf("config: group %q: 'path' field is required", g.Name)
		}

		// Default local_path to ./<name>
		if g.LocalPath == "" {
			g.LocalPath = "./" + g.Name
			c.Groups[i] = g
		}

		// Validate group repos
		for j, r := range g.Repos {
			if !strings.HasPrefix(r.Path, g.Path) {
				return fmt.Errorf("config: group %q: repo[%d] path %q does not start with group path %q", g.Name, j, r.Path, g.Path)
			}
		}
	}

	// Validate independent repos
	for i, r := range c.Repos {
		if r.Resource == "" {
			return fmt.Errorf("config: repos[%d]: 'resource' field is required", i)
		}
		if _, ok := c.Resources[r.Resource]; !ok {
			return fmt.Errorf("config: repos[%d]: resource %q not found", i, r.Resource)
		}
		// Default local_path to ./<name>
		if r.LocalPath == "" {
			r.LocalPath = "./" + r.Name
			c.Repos[i] = r
		}
	}

	return nil
}

// expandTilde expands ~/ to the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// ResolveGroupRepoPath derives the local path for a repo within a group.
// Formula: base + group.local_path + trimPrefix(repo.path, group.path)
func ResolveGroupRepoPath(base, groupLocalPath, groupPath, repoPath string) string {
	relative := strings.TrimPrefix(repoPath, groupPath)
	relative = strings.TrimPrefix(relative, "/")

	// Clean the local_path prefix
	lp := groupLocalPath
	lp = strings.TrimPrefix(lp, "./")

	parts := []string{base, lp}
	if relative != "" {
		parts = append(parts, relative)
	}
	return filepath.Join(parts...)
}

// ResolveRepoPath derives the local path for a standalone repo.
func ResolveRepoPath(base, localPath string) string {
	lp := strings.TrimPrefix(localPath, "./")
	return filepath.Join(base, lp)
}

// DetectPathConflicts checks for duplicate local paths across all repos.
func (c *Config) DetectPathConflicts() error {
	seen := make(map[string]string) // normalized path → description

	for _, g := range c.Groups {
		for _, r := range g.Repos {
			fullPath := ResolveGroupRepoPath(c.Base, g.LocalPath, g.Path, r.Path)
			norm := filepath.Clean(fullPath)
			desc := fmt.Sprintf("group %q repo %q", g.Name, r.Name)
			if prev, ok := seen[norm]; ok {
				return fmt.Errorf("path conflict: %s and %s both resolve to %s", prev, desc, norm)
			}
			seen[norm] = desc
		}
	}

	for _, r := range c.Repos {
		fullPath := ResolveRepoPath(c.Base, r.LocalPath)
		norm := filepath.Clean(fullPath)
		desc := fmt.Sprintf("repo %q", r.Name)
		if prev, ok := seen[norm]; ok {
			return fmt.Errorf("path conflict: %s and %s both resolve to %s", prev, desc, norm)
		}
		seen[norm] = desc
	}

	return nil
}

// --- File operations ---

// InitConfig creates a minimal config file at the given path.
func InitConfig(path, base string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	if base == "" {
		base = "~/projects"
	}

	cfg := &Config{
		Base:      base,
		Resources: map[string]Resource{},
		Groups:    []Group{},
		Repos:     []Repo{},
	}
	return writeConfig(path, cfg)
}

// AddResource appends a new resource entry to the config file.
func AddResource(configPath, name string, resource Resource) error {
	cfg, err := ensureConfigFile(configPath)
	if err != nil {
		return err
	}

	if cfg.Resources == nil {
		cfg.Resources = make(map[string]Resource)
	}

	// Save raw token and resolve it for runtime use
	rawToken := resource.Token
	resolved, err := resolveToken(rawToken)
	if err != nil {
		return err
	}
	resource.Token = resolved

	cfg.Resources[name] = resource

	// Store raw token for write-back
	if cfg.rawTokens == nil {
		cfg.rawTokens = make(map[string]string)
	}
	cfg.rawTokens[name] = rawToken

	return writeConfig(configPath, cfg)
}

// AddGroup appends a new group to the config file.
func AddGroup(configPath string, group Group) error {
	cfg, err := ensureConfigFile(configPath)
	if err != nil {
		return err
	}

	cfg.Groups = append(cfg.Groups, group)

	return writeConfig(configPath, cfg)
}

// AddGroupRepo appends a repo to a specific group's repos list (dedup by URL).
func AddGroupRepo(configPath, groupName string, repo GroupRepo) error {
	return WithFileLock(configPath, 30*time.Second, func() error {
		cfg, err := Load(configPath)
		if err != nil {
			return err
		}

		idx, group, err := cfg.FindGroup(groupName)
		if err != nil {
			return err
		}

		// Check for duplicate by URL
		for _, r := range group.Repos {
			if r.URL == repo.URL {
				return nil // already exists
			}
		}

		cfg.Groups[idx].Repos = append(cfg.Groups[idx].Repos, repo)
		return writeConfig(configPath, cfg)
	})
}

// AddRepo appends a new standalone repo to the config file.
func AddRepo(configPath string, repo Repo) error {
	cfg, err := ensureConfigFile(configPath)
	if err != nil {
		return err
	}

	cfg.Repos = append(cfg.Repos, repo)

	return writeConfig(configPath, cfg)
}

// SyncGroupRepos discovers repos for a group and writes them to config (dedup by URL).
// Returns the number of newly added repos.
func SyncGroupRepos(configPath, groupName string, newRepos []GroupRepo) (int, error) {
	var added int
	err := WithFileLock(configPath, 30*time.Second, func() error {
		cfg, err := Load(configPath)
		if err != nil {
			return err
		}

		idx, group, err := cfg.FindGroup(groupName)
		if err != nil {
			return err
		}

		for _, nr := range newRepos {
			found := false
			for _, er := range group.Repos {
				if er.URL == nr.URL {
					found = true
					break
				}
			}
			if !found {
				cfg.Groups[idx].Repos = append(cfg.Groups[idx].Repos, nr)
				added++
			}
		}

		if added == 0 {
			return nil
		}

		return writeConfig(configPath, cfg)
	})
	return added, err
}

func ensureConfigFile(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		cfg := &Config{
			Base:      "~/projects",
			Resources: map[string]Resource{},
			Groups:    []Group{},
			Repos:     []Repo{},
		}
		return cfg, nil
	}
	return Load(path)
}

// WithFileLock acquires an exclusive file lock on the given path,
// executes fn, then releases the lock.
func WithFileLock(path string, timeout time.Duration, fn func() error) error {
	lockPath := path + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		os.Remove(lockPath)
	}()

	done := make(chan error, 1)
	go func() {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
			done <- fmt.Errorf("acquire lock: %w", err)
			return
		}
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for config lock (another sync may be running)")
	}
}

// writeConfig writes the configuration to a YAML file.
// Preserves token placeholders and tilde notation in base path.
func writeConfig(path string, cfg *Config) error {
	// Re-tilde-ify base for storage
	base := cfg.Base
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(base, home+"/") {
		cfg.Base = "~/" + strings.TrimPrefix(base, home+"/")
	}

	// Restore raw tokens for storage (preserve ${VAR} placeholders)
	resolvedTokens := make(map[string]string)
	for name := range cfg.Resources {
		if raw, ok := cfg.rawTokens[name]; ok {
			res := cfg.Resources[name]
			resolvedTokens[name] = res.Token
			res.Token = raw
			cfg.Resources[name] = res
		}
	}

	// Restore raw group tokens for storage
	resolvedGroupTokens := make(map[int]string)
	for i, g := range cfg.Groups {
		if raw, ok := cfg.rawGroupTokens[i]; ok {
			resolvedGroupTokens[i] = g.Token
			g.Token = raw
			cfg.Groups[i] = g
		}
	}

	// Restore raw repo tokens for storage
	resolvedRepoTokens := make(map[int]string)
	for i, r := range cfg.Repos {
		if raw, ok := cfg.rawRepoTokens[i]; ok {
			resolvedRepoTokens[i] = r.Token
			r.Token = raw
			cfg.Repos[i] = r
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		// Restore resolved tokens even on error
		for name, t := range resolvedTokens {
			res := cfg.Resources[name]
			res.Token = t
			cfg.Resources[name] = res
		}
		for i, t := range resolvedGroupTokens {
			g := cfg.Groups[i]
			g.Token = t
			cfg.Groups[i] = g
		}
		for i, t := range resolvedRepoTokens {
			r := cfg.Repos[i]
			r.Token = t
			cfg.Repos[i] = r
		}
		return fmt.Errorf("marshal config: %w", err)
	}

	// Restore resolved tokens after marshaling
	for name, t := range resolvedTokens {
		res := cfg.Resources[name]
		res.Token = t
		cfg.Resources[name] = res
	}
	for i, t := range resolvedGroupTokens {
		g := cfg.Groups[i]
		g.Token = t
		cfg.Groups[i] = g
	}
	for i, t := range resolvedRepoTokens {
		r := cfg.Repos[i]
		r.Token = t
		cfg.Repos[i] = r
	}

	// Restore base path
	cfg.Base = base

	return os.WriteFile(path, data, 0644)
}
