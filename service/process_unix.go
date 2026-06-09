//go:build !windows

package service

import "syscall"

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func signalProcess(pid, pgid int, sig syscall.Signal) error {
	if pgid > 0 {
		return syscall.Kill(-pgid, sig)
	}
	return syscall.Kill(pid, sig)
}

func processPGID(pid int) int {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return pid
	}
	return pgid
}
