package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// WithRegistryLock acquires an exclusive lock on the registry file and runs fn.
func WithRegistryLock(registryPath string, fn func(*Registry) error) error {
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	lockPath := registryPath + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("open registry lock: %w", err)
	}
	defer func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		os.Remove(lockPath)
	}()

	done := make(chan error, 1)
	go func() {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
			done <- fmt.Errorf("acquire registry lock: %w", err)
			return
		}
		reg, err := loadRegistry(registryPath)
		if err != nil {
			done <- err
			return
		}
		if err := fn(reg); err != nil {
			done <- err
			return
		}
		done <- saveRegistry(registryPath, reg)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timed out waiting for service registry lock")
	}
}

func loadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Services: make(map[string]Record)}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	if reg.Services == nil {
		reg.Services = make(map[string]Record)
	}
	return &reg, nil
}

func saveRegistry(path string, reg *Registry) error {
	if reg.Services == nil {
		reg.Services = make(map[string]Record)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write registry temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename registry: %w", err)
	}
	return nil
}
