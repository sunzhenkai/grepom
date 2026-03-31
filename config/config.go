package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Base    string       `yaml:"base"`
	Sources []Source     `yaml:"sources"`
	Repos   []RepoEntry  `yaml:"repos"`
}

type Source struct {
	Provider string         `yaml:"provider"`
	URL      string         `yaml:"url"`
	Token    string         `yaml:"token"`
	Groups   []GroupSource  `yaml:"groups"`
	Orgs     []OrgSource    `yaml:"orgs"`
}

type GroupSource struct {
	Path      string `yaml:"path"`
	Recursive bool   `yaml:"recursive"`
}

type OrgSource struct {
	Name string `yaml:"name"`
}

type RepoEntry struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Path string `yaml:"path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Expand tilde in base path
	cfg.Base = expandTilde(cfg.Base)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func FindConfig(explicitPath string) (string, error) {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err != nil {
			return "", fmt.Errorf("config file not found: %s", explicitPath)
		}
		return explicitPath, nil
	}

	// Look for .grepom.yml in current directory
	if _, err := os.Stat(".grepom.yml"); err == nil {
		return ".grepom.yml", nil
	}

	return "", fmt.Errorf("no config file found. Use -c to specify a config file or create .grepom.yml in current directory")
}

func (c *Config) validate() error {
	if c.Base == "" {
		return fmt.Errorf("config: 'base' field is required")
	}
	for i, s := range c.Sources {
		if s.Provider == "" {
			return fmt.Errorf("config: sources[%d]: 'provider' field is required", i)
		}
		if s.Provider != "gitlab" && s.Provider != "github" {
			return fmt.Errorf("config: sources[%d]: unsupported provider %q (use 'gitlab' or 'github')", i, s.Provider)
		}
		if s.URL == "" {
			return fmt.Errorf("config: sources[%d]: 'url' field is required", i)
		}
		if !strings.HasPrefix(s.URL, "http://") && !strings.HasPrefix(s.URL, "https://") {
			c.Sources[i].URL = "https://" + s.URL
		}
	}
	return nil
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// AddSource appends a new source entry to the config file.
func AddSource(configPath string, source Source) error {
	cfg, err := ensureConfigFile(configPath)
	if err != nil {
		return err
	}

	// Expand env vars before writing
	cfg.Sources = append(cfg.Sources, source)

	return writeConfig(configPath, cfg)
}

// AddRepo appends a new explicit repo entry to the config file.
func AddRepo(configPath string, repo RepoEntry) error {
	cfg, err := ensureConfigFile(configPath)
	if err != nil {
		return err
	}

	cfg.Repos = append(cfg.Repos, repo)

	return writeConfig(configPath, cfg)
}

// InitConfig creates a minimal config file at the given path.
// Returns an error if the file already exists.
func InitConfig(path, base string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	if base == "" {
		base = "~/projects"
	}

	cfg := &Config{Base: base}
	return writeConfig(path, cfg)
}

func ensureConfigFile(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		// Create new config file
		cfg := &Config{
			Base: "~/projects",
		}
		return cfg, nil
	}
	return Load(path)
}

// WithFileLock acquires an exclusive file lock on the given path,
// executes fn, then releases the lock. Returns an error if the lock
// cannot be acquired within the timeout.
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

// SyncGroups reads the config file, appends new groups to the specified source
// (only if not already present), and writes back. Protected by file lock.
func SyncGroups(configPath string, sourceIndex int, newGroups []GroupSource) error {
	return WithFileLock(configPath, 30*time.Second, func() error {
		cfg, err := Load(configPath)
		if err != nil {
			return err
		}

		if sourceIndex < 0 || sourceIndex >= len(cfg.Sources) {
			return fmt.Errorf("source index %d out of range (0-%d)", sourceIndex, len(cfg.Sources)-1)
		}

		existing := cfg.Sources[sourceIndex].Groups
		added := 0
		for _, ng := range newGroups {
			found := false
			for _, eg := range existing {
				if eg.Path == ng.Path {
					found = true
					break
				}
			}
			if !found {
				cfg.Sources[sourceIndex].Groups = append(cfg.Sources[sourceIndex].Groups, ng)
				added++
			}
		}

		if added == 0 {
			return nil
		}

		return writeConfig(configPath, cfg)
	})
}

func writeConfig(path string, cfg *Config) error {
	// Re-tilde-ify base for storage
	base := cfg.Base
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(base, home+"/") {
		cfg.Base = "~/" + strings.TrimPrefix(base, home+"/")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
