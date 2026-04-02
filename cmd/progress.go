package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/git"
)

// isStdoutTerminal returns true if stdout is connected to a terminal.
// Uses a multi-strategy detection for robustness across shells and terminal emulators:
//  1. go-isatty (ioctl-based, most reliable)
//  2. TERM environment variable (non-empty, non-dumb)
//  3. /proc/self/fd/1 inode check (Linux only, major=5 for tty devices)
func isStdoutTerminal() bool {
	// 策略 1: go-isatty
	if isatty.IsTerminal(os.Stdout.Fd()) {
		config.Verbose("TTY check: isatty=true")
		return true
	}
	config.Verbose("TTY check: isatty=false")

	// 策略 2: TERM 环境变量
	term := os.Getenv("TERM")
	if term != "" && term != "dumb" && term != "unknown" {
		config.Verbose("TTY check: TERM=%s → TTY", term)
		return true
	}
	config.Verbose("TTY check: TERM=%q", term)

	// 策略 3: /proc/self/fd/1 inode 检查（仅 Linux）
	if runtime.GOOS == "linux" {
		if isTTYViaProc() {
			config.Verbose("TTY check: /proc=true → TTY")
			return true
		}
		config.Verbose("TTY check: /proc=false")
	}

	return false
}

// isTTYViaProc checks if stdout fd points to a tty device via /proc filesystem.
// On Linux, tty devices have major number 5 (char device 5,0 = /dev/tty).
func isTTYViaProc() bool {
	// 通过 os.Stat 检查 /proc/self/fd/1 指向的文件
	fi, err := os.Stat("/proc/self/fd/1")
	if err != nil {
		return false
	}
	// 检查是否为字符设备 (os.ModeCharDevice)
	// tty 设备的 major number 为 5，但 Go 的 os.FileMode 不直接暴露 major/minor
	// 使用 ModeCharDevice 作为近似判断：字符设备 + 非 stdin/stdout 本身
	return fi.Mode()&os.ModeCharDevice != 0
}

// ProgressRenderer handles real-time progress display for batch operations.
type ProgressRenderer struct {
	isTTY     bool
	total     int
	completed int
	action    string // "cloning" or "pulling"
}

// NewProgressRenderer creates a progress renderer for the given action and total count.
func NewProgressRenderer(action string, total int) *ProgressRenderer {
	return &ProgressRenderer{
		isTTY:  isStdoutTerminal(),
		total:  total,
		action: action,
	}
}

// Update prints the current progress line.
// In TTY mode, it uses \r to overwrite; in non-TTY mode, it prints a new line.
func (p *ProgressRenderer) Update(completed int) {
	p.completed = completed
	line := fmt.Sprintf("[%d/%d] %s...", completed, p.total, p.action)
	if p.isTTY {
		fmt.Fprintf(os.Stdout, "\r%s", line)
	} else {
		fmt.Fprintln(os.Stdout, line)
	}
}

// Done clears the progress line (TTY mode) and prints final status.
func (p *ProgressRenderer) Done() {
	if p.isTTY && p.total > 0 {
		// Clear the progress line
		fmt.Fprintf(os.Stdout, "\r%s\r", spaces(len(fmt.Sprintf("[%d/%d] %s...", p.completed, p.total, p.action))))
		fmt.Fprintln(os.Stdout)
	}
}

// PrintCloneSummary outputs the clone operation summary.
func PrintCloneSummary(results []git.CloneResult, writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	succeeded := 0
	failed := 0
	var failedResults []git.CloneResult

	for _, r := range results {
		if r.Err != nil {
			failed++
			failedResults = append(failedResults, r)
		} else {
			succeeded++
		}
	}

	total := len(results)

	if total == 0 {
		fmt.Fprintln(writer, "all repositories already cloned")
		return
	}

	if failed == 0 {
		fmt.Fprintf(writer, "clone complete: %d/%d succeeded\n", succeeded, total)
	} else {
		fmt.Fprintf(writer, "clone complete: %d/%d succeeded, %d failed\n", succeeded, total, failed)
		for _, r := range failedResults {
			mark := "✗"
			fmt.Fprintf(writer, "  %s %s: %v\n", mark, r.Repo.Path, r.Err)
		}
	}
}

// PrintPullSummary outputs the pull operation summary.
func PrintPullSummary(results []git.PullResult, skipped int, writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	succeeded := 0
	failed := 0
	var failedResults []git.PullResult

	for _, r := range results {
		if r.Err != nil {
			failed++
			failedResults = append(failedResults, r)
		} else {
			succeeded++
		}
	}

	total := len(results)

	if total == 0 && skipped > 0 {
		fmt.Fprintf(writer, "nothing to pull: %d skipped (not on default branch or dirty)\n", skipped)
		return
	}

	if total == 0 {
		fmt.Fprintln(writer, "no repositories to pull")
		return
	}

	if failed == 0 {
		if skipped > 0 {
			fmt.Fprintf(writer, "pull complete: %d/%d succeeded, %d skipped\n", succeeded, total, skipped)
		} else {
			fmt.Fprintf(writer, "pull complete: %d/%d succeeded\n", succeeded, total)
		}
	} else {
		if skipped > 0 {
			fmt.Fprintf(writer, "pull complete: %d/%d succeeded, %d failed, %d skipped\n", succeeded, total, failed, skipped)
		} else {
			fmt.Fprintf(writer, "pull complete: %d/%d succeeded, %d failed\n", succeeded, total, failed)
		}
		for _, r := range failedResults {
			mark := "✗"
			fmt.Fprintf(writer, "  %s %s: %v\n", mark, r.Repo.Path, r.Err)
		}
	}
}

func spaces(n int) string {
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}
