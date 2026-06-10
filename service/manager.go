package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wii/grepom/config"
)

// Manager coordinates service lifecycle operations for a single scope.
type Manager struct {
	ScopeID    string
	StateDir   string
	Registry   string
	ConfigPath string
	ConfigDir  string
	Services   map[string]config.ServiceDef
}

// RunOptions configures service startup.
type RunOptions struct {
	Name    string
	Cwd     string
	Command config.ServiceCommand
	Force   bool
}

// CleanOptions configures registry cleanup.
type CleanOptions struct {
	RemoveLogs bool
	All        bool
}

// NewManager creates a manager for the given config path and optional service definitions.
func NewManager(configPath string, services map[string]config.ServiceDef) (*Manager, error) {
	scopeID, err := ScopeFromPath(configPath)
	if err != nil {
		return nil, err
	}
	stateDir, err := StateDir(scopeID)
	if err != nil {
		return nil, err
	}
	configDir := ""
	if configPath != "" {
		configDir = filepath.Dir(configPath)
	}
	return &Manager{
		ScopeID:    scopeID,
		StateDir:   stateDir,
		Registry:   RegistryPath(stateDir),
		ConfigPath: configPath,
		ConfigDir:  configDir,
		Services:   services,
	}, nil
}

// Run starts a service in the background and records it in the registry.
func (m *Manager) Run(opts RunOptions) (*Record, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if opts.Cwd == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		opts.Cwd = cwd
	}
	absCwd, err := filepath.Abs(opts.Cwd)
	if err != nil {
		return nil, err
	}
	opts.Cwd = absCwd
	if opts.Command.IsEmpty() {
		return nil, fmt.Errorf("command is required")
	}

	logPath := LogPathForService(m.StateDir, opts.Name)
	cmd, err := startCommand(opts.Cwd, opts.Command, logPath)
	if err != nil {
		return nil, err
	}

	pgid := processPGID(cmd.Process.Pid)
	cmdStr, cmdArgs := commandDisplay(opts.Command)
	rec := Record{
		Name:        opts.Name,
		PID:         cmd.Process.Pid,
		PGID:        pgid,
		Cwd:         opts.Cwd,
		Command:     cmdStr,
		CommandArgs: cmdArgs,
		LogPath:     logPath,
		StartedAt:   time.Now().UTC(),
		LastStatus:  StatusRunning,
		ConfigPath:  m.ConfigPath,
	}

	err = WithRegistryLock(m.Registry, func(reg *Registry) error {
		if existing, ok := reg.Services[opts.Name]; ok {
			status := evaluateStatus(existing)
			if status == StatusRunning && !opts.Force {
				_ = signalProcess(cmd.Process.Pid, pgid, syscall.SIGTERM)
				return fmt.Errorf("service %q is already running (pid %d); use --force to replace", opts.Name, existing.PID)
			}
		}
		reg.Services[opts.Name] = rec
		return nil
	})
	if err != nil {
		_ = signalProcess(cmd.Process.Pid, pgid, syscall.SIGTERM)
		return nil, err
	}

	return &rec, nil
}

// List returns all services with live status.
func (m *Manager) List() ([]Entry, error) {
	reg, err := loadRegistry(m.Registry)
	if err != nil {
		return nil, err
	}
	return m.entriesFromRegistry(reg), nil
}

// Status returns one service entry by name.
func (m *Manager) Status(name string) (*Entry, error) {
	reg, err := loadRegistry(m.Registry)
	if err != nil {
		return nil, err
	}
	rec, ok := reg.Services[name]
	if !ok {
		return nil, fmt.Errorf("service %q not found", name)
	}
	status := evaluateStatus(rec)
	rec.LastStatus = status
	return &Entry{Record: rec, Status: status}, nil
}

// Logs prints or follows service logs.
func (m *Manager) Logs(ctx context.Context, name string, lines int, follow bool, open bool, w interface{ Write([]byte) (int, error) }) error {
	entry, err := m.Status(name)
	if err != nil {
		return err
	}
	if open {
		return OpenLog(entry.Record.LogPath, os.Stdout)
	}
	if follow {
		return FollowLog(ctx, entry.Record.LogPath, w)
	}
	if lines <= 0 {
		lines = defaultLogLines
	}
	tail, err := ReadTailLines(entry.Record.LogPath, lines)
	if err != nil {
		return err
	}
	for _, line := range tail {
		fmt.Fprintln(w, line)
	}
	return nil
}

// Restart stops a running service if needed and starts it again with the same command.
func (m *Manager) Restart(name string) (*Record, error) {
	entry, err := m.Status(name)
	if err != nil {
		return nil, err
	}

	if err := m.stopForRestart(entry); err != nil {
		return nil, err
	}

	opts, err := m.runOptionsForRestart(name, entry.Record)
	if err != nil {
		return nil, err
	}
	opts.Force = true
	return m.Run(opts)
}

func (m *Manager) stopForRestart(entry *Entry) error {
	if !isProcessAlive(entry.Record.PID) {
		return nil
	}
	return m.Kill(entry.Record.Name, true)
}

func (m *Manager) runOptionsForRestart(name string, rec Record) (RunOptions, error) {
	if def, ok := m.Services[name]; ok && !def.Command.IsEmpty() {
		cwd := config.ResolveServiceCwd(m.ConfigDir, def.Cwd)
		if cwd == "" {
			cwd = rec.Cwd
		}
		return RunOptions{
			Name:    name,
			Cwd:     cwd,
			Command: def.Command,
		}, nil
	}

	opts := RunOptions{Name: name, Cwd: rec.Cwd}
	if len(rec.CommandArgs) > 0 {
		opts.Command = config.ServiceCommand{Args: rec.CommandArgs}
		return opts, nil
	}
	if rec.Command != "" {
		opts.Command = config.ServiceCommand{Shell: rec.Command}
		return opts, nil
	}
	return RunOptions{}, fmt.Errorf("service %q has no command to restart", name)
}

// Kill sends a signal to a managed service.
func (m *Manager) Kill(name string, force bool) error {
	entry, err := m.Status(name)
	if err != nil {
		return err
	}
	if entry.Status != StatusRunning {
		return fmt.Errorf("service %q is not running", name)
	}

	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	if err := signalProcess(entry.Record.PID, entry.Record.PGID, sig); err != nil {
		return fmt.Errorf("kill service %q: %w", name, err)
	}

	return WithRegistryLock(m.Registry, func(reg *Registry) error {
		rec, ok := reg.Services[name]
		if !ok {
			return fmt.Errorf("service %q not found", name)
		}
		rec.LastStatus = StatusExited
		if force {
			rec.ExitStatus = "killed"
		} else {
			rec.ExitStatus = "terminated"
		}
		reg.Services[name] = rec
		return nil
	})
}

// Clean removes exited or stale service records.
func (m *Manager) Clean(opts CleanOptions) (int, error) {
	removed := 0
	err := WithRegistryLock(m.Registry, func(reg *Registry) error {
		for name, rec := range reg.Services {
			status := evaluateStatus(rec)
			if status == StatusRunning {
				continue
			}
			if !opts.All && status != StatusExited && status != StatusStale {
				continue
			}
			if opts.RemoveLogs && rec.LogPath != "" {
				_ = os.Remove(rec.LogPath)
			}
			delete(reg.Services, name)
			removed++
		}
		return nil
	})
	return removed, err
}

// Dir returns the working directory for a service.
func (m *Manager) Dir(name string) (string, error) {
	if def, ok := m.Services[name]; ok {
		return config.ResolveServiceCwd(m.ConfigDir, def.Cwd), nil
	}
	entry, err := m.Status(name)
	if err != nil {
		return "", err
	}
	return entry.Record.Cwd, nil
}

func (m *Manager) entriesFromRegistry(reg *Registry) []Entry {
	names := make([]string, 0, len(reg.Services))
	for name := range reg.Services {
		names = append(names, name)
	}
	// stable order
	sortStrings(names)

	var entries []Entry
	for _, name := range names {
		rec := reg.Services[name]
		status := evaluateStatus(rec)
		rec.LastStatus = status
		entries = append(entries, Entry{Record: rec, Status: status})
	}
	return entries
}

func sortStrings(ss []string) {
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[j] < ss[i] {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}
