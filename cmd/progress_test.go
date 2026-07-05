package cmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/wii/grepom/git"
)

// newTestRenderer constructs a ProgressRenderer with an injected output buffer
// so tests can run deterministically without touching the real stdout / TTY.
func newTestRenderer(isTTY bool, total int) (*ProgressRenderer, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return &ProgressRenderer{
		isTTY:  isTTY,
		out:    buf,
		total:  total,
		action: "pulling",
	}, buf
}

// TestProgressRenderer_ConcurrentHandleNoRace exercises Handle/Done under heavy
// concurrency from many goroutines. Run with `go test -race` to verify there
// are no data races on inflight/completed/rendered and no panics.
func TestProgressRenderer_ConcurrentHandleNoRace(t *testing.T) {
	// Non-TTY path: also stresses the per-line stdout write critical section.
	p, _ := newTestRenderer(false, 100)
	p.out = io.Discard // we don't assert content here, only safety

	const workers = 50
	const tasksPerWorker = 20
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < tasksPerWorker; j++ {
				name := fmt.Sprintf("repo-%d-%d", id, j)
				p.Handle(git.ProgressEvent{Type: git.ProgressStart, RepoName: name})
				p.Handle(git.ProgressEvent{
					Type:      git.ProgressComplete,
					RepoName:  name,
					Completed: id*tasksPerWorker + j + 1,
					Total:     workers * tasksPerWorker,
				})
			}
		}(w)
	}
	wg.Wait()

	// Done is called concurrently-once by the deferred cleanup in real code;
	// here we just call it after the workers finish to ensure TTY clear path
	// is also exercised without panicking.
	p.isTTY = true
	p.rendered = 3
	p.out = io.Discard
	p.Done()
}

// TestProgressRenderer_TTYShrinkClearsResidual verifies that when the in-flight
// count shrinks, renderTTY writes blank padding lines to overwrite the now-stale
// rows and parks the cursor back at the first line of the progress area.
func TestProgressRenderer_TTYShrinkClearsResidual(t *testing.T) {
	p, buf := newTestRenderer(true, 20)
	// Simulate a prior render that drew 6 lines (1 header + 5 in-flight).
	p.inflight = []inflightTask{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}}
	p.rendered = 6
	p.completed = 13

	// Reduce in-flight to 2 by removing c, d, e via the public API.
	for _, name := range []string{"c", "d", "e"} {
		buf.Reset()
		p.completed++ // mimic collector advancing the counter
		p.Handle(git.ProgressEvent{
			Type:      git.ProgressComplete,
			RepoName:  name,
			Completed: p.completed,
			Total:     20,
		})
	}

	// Inspect the LAST complete render (the final shrink to lines=3).
	out := buf.String()
	// lines = 1 + len(inflight) = 3, prev was 4 before this last render.
	// Pad lines written = prev - lines = 4 - 3 = 1.
	// Newlines: (lines-1) content separators + pad = 2 + 1 = 3.
	if got := strings.Count(out, "\n"); got != 3 {
		t.Errorf("shrink render newline count: got %d, want 3 (2 content + 1 pad)", got)
	}
	// rowsWritten = max(lines, prev) = max(3,4) = 4 → park cursor up 3 rows.
	if !strings.Contains(out, "\033[3A") {
		t.Errorf("expected cursor-up 3 to park at first line; output=%q", out)
	}
	if p.rendered != 3 {
		t.Errorf("rendered = %d, want 3 after shrink", p.rendered)
	}
	if len(p.inflight) != 2 {
		t.Errorf("inflight len = %d, want 2", len(p.inflight))
	}
}

// TestProgressRenderer_TTYGrowAndShrink covers the repeated grow/shrink cycle
// to ensure the baseline stays consistent across many transitions (no drift,
// rendered always tracks the true content height).
func TestProgressRenderer_TTYGrowAndShrink(t *testing.T) {
	p, _ := newTestRenderer(true, 50)
	for i := 0; i < 50; i++ {
		p.Handle(git.ProgressEvent{Type: git.ProgressStart, RepoName: fmt.Sprintf("r%d", i)})
	}
	if p.rendered != 51 {
		t.Fatalf("after 50 starts rendered = %d, want 51", p.rendered)
	}
	// Now drain them all; rendered should converge to 1 (header only).
	for i := 0; i < 50; i++ {
		p.Handle(git.ProgressEvent{
			Type:      git.ProgressComplete,
			RepoName:  fmt.Sprintf("r%d", i),
			Completed: i + 1,
			Total:     50,
		})
	}
	if p.rendered != 1 {
		t.Errorf("after drain rendered = %d, want 1", p.rendered)
	}
	if len(p.inflight) != 0 {
		t.Errorf("inflight len = %d, want 0", len(p.inflight))
	}
}

// TestProgressRenderer_NonTTYConcurrentLinesIntact fires many concurrent
// ProgressComplete events in non-TTY mode and asserts every emitted line is a
// complete, well-formed single line (no interleaving/duplication).
func TestProgressRenderer_NonTTYConcurrentLinesIntact(t *testing.T) {
	p, buf := newTestRenderer(false, 100)

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("repo-%03d", idx)
			p.Handle(git.ProgressEvent{
				Type:      git.ProgressComplete,
				RepoName:  name,
				Completed: idx + 1,
				Total:     n,
			})
		}(i)
	}
	wg.Wait()

	got := buf.String()
	if got == "" {
		t.Fatal("expected output, got empty buffer")
	}
	trimmed := strings.TrimRight(got, "\n")
	lines := strings.Split(trimmed, "\n")
	if len(lines) != n {
		t.Fatalf("line count: got %d, want %d (buf=%q)", len(lines), n, got)
	}
	seen := make(map[string]bool, n)
	for _, l := range lines {
		if !strings.HasPrefix(l, "✓ repo-") {
			t.Errorf("malformed/interleaved line: %q", l)
		}
		if seen[l] {
			t.Errorf("duplicate line (possible interleaving): %q", l)
		}
		seen[l] = true
	}
	if len(seen) != n {
		t.Errorf("unique lines: got %d, want %d", len(seen), n)
	}
}

// TestProgressRenderer_TTYCompletedMonotonic mirrors the production event model
// (Start fired by many worker goroutines, Complete fired by a single collector
// goroutine with a strictly increasing counter) and asserts the rendered
// [N/M] header counter never goes backwards.
func TestProgressRenderer_TTYCompletedMonotonic(t *testing.T) {
	p, buf := newTestRenderer(true, 200)

	const workers = 50
	const tasksPerWorker = 4
	const total = workers * tasksPerWorker // 200

	// Build the set of repo names the collector will complete, in order.
	names := make([]string, 0, total)
	for w := 0; w < workers; w++ {
		for j := 0; j < tasksPerWorker; j++ {
			names = append(names, fmt.Sprintf("repo-%d-%d", w, j))
		}
	}

	var startWG, doneWG sync.WaitGroup

	// Workers fire ProgressStart concurrently (mirrors parallel.go workers).
	for w := 0; w < workers; w++ {
		startWG.Add(1)
		go func(id int) {
			defer startWG.Done()
			for j := 0; j < tasksPerWorker; j++ {
				p.Handle(git.ProgressEvent{
					Type:     git.ProgressStart,
					RepoName: fmt.Sprintf("repo-%d-%d", id, j),
				})
			}
		}(w)
	}

	// Collector fires ProgressComplete with a strictly increasing counter
	// (single goroutine, exactly like parallel.go's result loop).
	doneWG.Add(1)
	go func() {
		defer doneWG.Done()
		// Wait until at least some starts have landed so completes have an
		// inflight entry to remove; then interleave with the remaining starts.
		for i := 0; i < total; i++ {
			p.Handle(git.ProgressEvent{
				Type:      git.ProgressComplete,
				RepoName:  names[i],
				Completed: i + 1,
				Total:     total,
			})
		}
	}()

	startWG.Wait()
	doneWG.Wait()

	// Extract every [N/200] counter rendered and assert monotonic non-decreasing.
	re := regexp.MustCompile(`\[(\d+)/200\]`)
	matches := re.FindAllStringSubmatch(buf.String(), -1)
	if len(matches) == 0 {
		t.Fatal("no [N/200] headers captured")
	}
	prev := -1
	for _, m := range matches {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatalf("bad counter %q: %v", m[1], err)
		}
		if n < prev {
			t.Fatalf("completed counter went backwards: %d → %d", prev, n)
		}
		prev = n
	}
	if prev != total {
		t.Errorf("final counter = %d, want %d", prev, total)
	}
}
