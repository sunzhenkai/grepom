package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
