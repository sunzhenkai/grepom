package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/wii/grepom/git"
)

// isStdoutTerminal returns true if stdout is connected to a terminal.
func isStdoutTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
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
