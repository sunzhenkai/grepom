package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServiceCommand represents a service start command as a shell string or argv slice.
type ServiceCommand struct {
	Shell string   `yaml:"-"`
	Args  []string `yaml:"-"`
}

// String returns a human-readable command representation.
func (c ServiceCommand) String() string {
	if c.Shell != "" {
		return c.Shell
	}
	return strings.Join(c.Args, " ")
}

// IsEmpty reports whether the command is unset.
func (c ServiceCommand) IsEmpty() bool {
	return c.Shell == "" && len(c.Args) == 0
}

// UnmarshalYAML accepts either a scalar shell command or a sequence of argv tokens.
func (c *ServiceCommand) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		c.Shell = value.Value
		c.Args = nil
	case yaml.SequenceNode:
		var args []string
		if err := value.Decode(&args); err != nil {
			return fmt.Errorf("service command: %w", err)
		}
		if len(args) == 0 {
			return fmt.Errorf("service command array must not be empty")
		}
		c.Args = args
		c.Shell = ""
	default:
		return fmt.Errorf("service command must be a string or string array")
	}
	return nil
}

// ServiceDef declares how to start a named local development service.
type ServiceDef struct {
	Cwd     string         `yaml:"cwd,omitempty"`
	Command ServiceCommand `yaml:"command"`
}

// ResolveServiceCwd resolves a service working directory relative to the config file directory.
func ResolveServiceCwd(configDir, cwd string) string {
	if cwd == "" {
		return configDir
	}
	if filepath.IsAbs(cwd) {
		return filepath.Clean(cwd)
	}
	return filepath.Clean(filepath.Join(configDir, cwd))
}

// FindService returns a configured service definition by name.
func (c *Config) FindService(name string) (*ServiceDef, error) {
	if c.Services == nil {
		return nil, fmt.Errorf("service %q not found", name)
	}
	def, ok := c.Services[name]
	if !ok {
		return nil, fmt.Errorf("service %q not found", name)
	}
	return &def, nil
}
