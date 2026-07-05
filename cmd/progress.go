package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

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
//
// Concurrency: Handle and Done are safe to call from multiple goroutines
// (e.g. ProgressStart fired by worker goroutines, ProgressComplete fired by
// the result-collector goroutine). All shared state reads/writes and the
// entire stdout write (ANSI cursor motion + text) are serialized by mu, so
// progress lines never interleave and the [N/M] counter is monotonic.
type ProgressRenderer struct {
	mu        sync.Mutex
	isTTY     bool
	out       io.Writer // destination; defaults to os.Stdout, injectable for tests
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
		out:    os.Stdout,
		total:  total,
		action: action,
	}
}

// Handle processes a progress event and updates the display.
// Start events add the repo to the in-flight list; Complete events remove it.
//
// This method is goroutine-safe. The mutex serializes all shared-state
// mutations (inflight slice, completed counter) and the full stdout rewrite,
// so concurrent Start/Complete events cannot interleave output or race.
func (p *ProgressRenderer) Handle(event git.ProgressEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch event.Type {
	case git.ProgressStart:
		p.inflight = append(p.inflight, inflightTask{name: event.RepoName})
		// Note: ProgressStart MUST NOT mutate p.completed; the counter is
		// advanced exclusively by ProgressComplete so that [N/M] stays
		// monotonic non-decreasing regardless of Start/Complete interleaving.
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
				fmt.Fprintf(p.out, "✗ %s: %v\n", event.RepoName, event.Err)
			} else {
				fmt.Fprintf(p.out, "✓ %s\n", event.RepoName)
			}
		}
	}
}

// renderTTY 渲染多行 TTY 进度区域。
// 第一行：[N/M] action...
// 后续行：  action repo-name...
//
// 渲染模型（基线一致）：每次重绘结束后光标停在进度区域第一行起点，
// 因此下一次重绘只需 \r 回列首即可开始覆盖。
// 当本次行数 (lines) 小于上次行数 (prev) 时，对多余的 prev-lines 行用
// 空白（pad 至 maxWidth）显式覆盖，避免残留过期的仓库名。
func (p *ProgressRenderer) renderTTY() {
	prev := p.rendered
	lines := 1 + len(p.inflight) // 摘要行 + in-flight 行

	// 上一帧的光标停在第一行起点；回到列首即可覆盖重绘。
	fmt.Fprint(p.out, "\r")

	// 渲染第一行：进度摘要
	header := fmt.Sprintf("[%d/%d] %s...", p.completed, p.total, p.action)
	fmt.Fprintf(p.out, "%s%s", header, spaces(maxWidth-paddedLen(header)))

	// 渲染 in-flight 行
	for _, t := range p.inflight {
		fmt.Fprint(p.out, "\n")
		line := fmt.Sprintf("  %s %s...", p.action, t.name)
		fmt.Fprintf(p.out, "%s%s", line, spaces(maxWidth-paddedLen(line)))
	}

	// 行数缩减时，用空行覆盖多余的旧行（覆盖式画布必须显式擦除）。
	for i := lines; i < prev; i++ {
		fmt.Fprint(p.out, "\n")
		fmt.Fprint(p.out, spaces(maxWidth))
	}

	// 将光标移回进度区域第一行起点，保证下次重绘基准一致。
	// 本次实际写入的行数（含 pad）取 max(lines, prev)。
	rowsWritten := lines
	if prev > lines {
		rowsWritten = prev
	}
	if rowsWritten > 1 {
		fmt.Fprintf(p.out, "\033[%dA", rowsWritten-1)
	}

	p.rendered = lines
}

// Done clears the progress area (TTY mode) and ensures output continues on a new line.
// Goroutine-safe; the entire clear sequence is a single critical section.
func (p *ProgressRenderer) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !(p.isTTY && p.rendered > 0) {
		return
	}
	// renderTTY parks the cursor at the first line of the progress area,
	// so we clear line-by-line from here and end one line below the area.
	for i := 0; i < p.rendered; i++ {
		fmt.Fprintf(p.out, "%s\r", spaces(maxWidth))
		if i < p.rendered-1 {
			fmt.Fprint(p.out, "\n")
		}
	}
	fmt.Fprintln(p.out)
	p.rendered = 0
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
