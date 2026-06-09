package service

import "time"

// Status values for a managed service process.
const (
	StatusRunning = "running"
	StatusExited  = "exited"
	StatusStale   = "stale"
)

// Record stores runtime metadata for a managed service.
type Record struct {
	Name        string    `json:"name"`
	PID         int       `json:"pid"`
	PGID        int       `json:"pgid,omitempty"`
	Cwd         string    `json:"cwd"`
	Command     string    `json:"command"`
	CommandArgs []string  `json:"command_args,omitempty"`
	LogPath     string    `json:"log_path"`
	StartedAt   time.Time `json:"started_at"`
	LastStatus  string    `json:"last_status,omitempty"`
	ExitStatus  string    `json:"exit_status,omitempty"`
	ConfigPath  string    `json:"config_path,omitempty"`
}

// Entry is a service record with a live status evaluation.
type Entry struct {
	Record Record
	Status string
}

// Registry stores all managed services for a scope.
type Registry struct {
	Services map[string]Record `json:"services"`
}
