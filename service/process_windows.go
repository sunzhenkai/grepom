//go:build windows

package service

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	var code uint32
	err = windows.GetExitCodeProcess(handle, &code)
	return err == nil && code == 259 // STILL_ACTIVE
}

func signalProcess(pid, pgid int, sig syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}

func processPGID(pid int) int {
	return 0
}
