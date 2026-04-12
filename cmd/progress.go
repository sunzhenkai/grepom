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

// inflightTask 记录一个正在处理的任务。
type inflightTask struct {
	name string
}

// ProgressRenderer handles real-time progress display for batch operations.
// In TTY mode, it renders a multi-line progress area showing all in-flight tasks.
// In non-TTY mode, it prints each completed result on its own line.
type ProgressRenderer struct {
	isTTY     bool
	total     int
	completed int
	action    string // "cloning" or "pulling"
	inflight  []inflightTask
	rendered  int // 上次渲染的行数，用于清除
}

// NewProgressRenderer creates a progress renderer for the given action and total count.
func NewProgressRenderer(action string, total int) *ProgressRenderer {
	return &ProgressRenderer{
		isTTY:  isStdoutTerminal(),
		total:  total,
		action: action,
	}
}

// Handle processes a progress event and updates the display.
// Start events add the repo to the in-flight list; Complete events remove it.
func (p *ProgressRenderer) Handle(event git.ProgressEvent) {
	switch event.Type {
	case git.ProgressStart:
		p.inflight = append(p.inflight, inflightTask{name: event.RepoName})
		if p.isTTY {
			p.renderTTY()
		}
	case git.ProgressComplete:
		// 从 in-flight 列表中移除
		for i, t := range p.inflight {
			if t.name == event.RepoName {
				p.inflight = append(p.inflight[:i], p.inflight[i+1:]...)
				break
			}
		}
		p.completed = event.Completed
		if p.isTTY {
			p.renderTTY()
		} else {
			// 非 TTY：逐行输出完成结果
			if event.Err != nil {
				fmt.Fprintf(os.Stdout, "✗ %s: %v\n", event.RepoName, event.Err)
			} else {
				fmt.Fprintf(os.Stdout, "✓ %s\n", event.RepoName)
			}
		}
	}
}

// renderTTY 渲染多行 TTY 进度区域。
// 第一行：[N/M] action...
// 后续行：  action repo-name...
func (p *ProgressRenderer) renderTTY() {
	// 计算需要渲染的总行数
	lines := 1 + len(p.inflight) // 摘要行 + in-flight 行

	// 如果之前渲染过行，先将光标移回到第一行位置
	if p.rendered > 0 {
		// 光标上移 (rendered - 1) 行到第一行，然后 \r 回到行首
		if p.rendered > 1 {
			fmt.Fprintf(os.Stdout, "\033[%dA", p.rendered-1)
		}
		fmt.Fprint(os.Stdout, "\r")
	}

	// 渲染第一行：进度摘要
	header := fmt.Sprintf("[%d/%d] %s...", p.completed, p.total, p.action)
	fmt.Fprintf(os.Stdout, "%s%s", header, spaces(maxWidth-paddedLen(header)))

	// 渲染 in-flight 行
	for _, t := range p.inflight {
		fmt.Fprint(os.Stdout, "\n")
		line := fmt.Sprintf("  %s %s...", p.action, t.name)
		fmt.Fprintf(os.Stdout, "%s%s", line, spaces(maxWidth-paddedLen(line)))
	}

	p.rendered = lines
}

// Done clears the progress area (TTY mode) and ensures output continues on a new line.
func (p *ProgressRenderer) Done() {
	if p.isTTY && p.rendered > 0 {
		// 将光标移回进度区域第一行
		if p.rendered > 1 {
			fmt.Fprintf(os.Stdout, "\033[%dA", p.rendered-1)
		}
		fmt.Fprint(os.Stdout, "\r")
		// 用空行覆盖所有已渲染的行
		for i := 0; i < p.rendered; i++ {
			fmt.Fprintf(os.Stdout, "%s\r", spaces(maxWidth))
			if i < p.rendered-1 {
				fmt.Fprint(os.Stdout, "\n")
			}
		}
		fmt.Fprintln(os.Stdout)
		p.rendered = 0
	}
}

// maxWidth 是进度行的最大宽度，用于计算清除空格。
const maxWidth = 120

// paddedLen 返回字符串的字节长度（用于计算清除空格数）。
func paddedLen(s string) int {
	return len(s)
}

// spaces 返回 n 个空格的字符串。
func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
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
