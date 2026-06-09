package service

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// UserConfigDirFunc resolves the user config directory. Injectable for tests.
var UserConfigDirFunc = os.UserConfigDir

// ScopeFromPath returns a stable scope identifier for a config path or working directory.
func ScopeFromPath(configPath string) (string, error) {
	key := configPath
	if key == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		key, err = filepath.Abs(cwd)
		if err != nil {
			return "", err
		}
	} else {
		abs, err := filepath.Abs(configPath)
		if err != nil {
			return "", err
		}
		key = abs
	}
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:16]), nil
}

// StateDir returns the machine-local state directory for a scope.
func StateDir(scopeID string) (string, error) {
	base, err := UserConfigDirFunc()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "grepom", "services", scopeID), nil
}

// RegistryPath returns the registry file path inside a state directory.
func RegistryPath(stateDir string) string {
	return filepath.Join(stateDir, "registry.json")
}

// LogsDir returns the logs directory inside a state directory.
func LogsDir(stateDir string) string {
	return filepath.Join(stateDir, "logs")
}

// LogPathForService returns the log file path for a service name.
func LogPathForService(stateDir, name string) string {
	return filepath.Join(LogsDir(stateDir), name+".log")
}
