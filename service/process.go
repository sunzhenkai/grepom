package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/wii/grepom/config"
)

func evaluateStatus(rec Record) string {
	if !isProcessAlive(rec.PID) {
		if rec.LastStatus == StatusRunning || rec.LastStatus == "" {
			return StatusExited
		}
		return rec.LastStatus
	}
	return StatusRunning
}

func startCommand(cwd string, command config.ServiceCommand, logPath string) (*exec.Cmd, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	separator := fmt.Sprintf("\n===== grepom svc start %s =====\n", time.Now().Format(time.RFC3339))
	if _, err := logFile.WriteString(separator); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("write log separator: %w", err)
	}

	var cmd *exec.Cmd
	switch {
	case command.Shell != "":
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", command.Shell)
		} else {
			cmd = exec.Command("sh", "-c", command.Shell)
		}
	case len(command.Args) > 0:
		cmd = exec.Command(command.Args[0], command.Args[1:]...)
	default:
		logFile.Close()
		return nil, fmt.Errorf("empty command")
	}

	cmd.Dir = cwd
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("start process: %w", err)
	}

	// Detach from parent; log file stays open in child via inherited fd until child exec replaces.
	// On Unix the child inherits the fd; we keep our handle until Start returns.
	logFile.Close()

	return cmd, nil
}

func commandDisplay(command config.ServiceCommand) (string, []string) {
	if command.Shell != "" {
		return command.Shell, nil
	}
	return strings.Join(command.Args, " "), command.Args
}
