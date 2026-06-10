package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const defaultLogLines = 50

// ReadTailLines reads up to n trailing lines while keeping only a bounded line buffer in memory.
func ReadTailLines(path string, n int) ([]string, error) {
	if n <= 0 {
		n = defaultLogLines
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("log file not found: %s", path)
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// ReadLinesFromOffset reads complete lines from path starting at offset.
// Returns the lines, the offset for the next read, and any error.
// A partial line without a trailing newline is left unread for the next call.
func ReadLinesFromOffset(path string, offset int64) ([]string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, offset, nil
		}
		return nil, offset, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, offset, err
	}

	reader := bufio.NewReader(f)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			lines = append(lines, strings.TrimSuffix(line, "\n"))
		}
		if err == io.EOF {
			newOffset, _ := f.Seek(0, io.SeekCurrent)
			if len(line) > 0 && !strings.HasSuffix(line, "\n") {
				newOffset -= int64(len(line))
			}
			return lines, newOffset, nil
		}
		if err != nil {
			return lines, offset, err
		}
	}
}

// LogFileSize returns the size of a log file, or 0 if it does not exist.
func LogFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// FollowLog prints new lines appended to a log file until ctx is cancelled.
func FollowLog(ctx context.Context, path string, w io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if len(line) > 0 {
					if _, werr := io.WriteString(w, line); werr != nil {
						return werr
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
			}
		}
	}
}

// OpenLog opens a log file with editor or platform opener, or prints the path.
func OpenLog(path string, stdout io.Writer) error {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		if editor := os.Getenv(key); editor != "" {
			parts := splitShell(editor)
			if len(parts) == 0 {
				continue
			}
			cmd := exec.Command(parts[0], append(parts[1:], path)...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			return cmd.Run()
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(stdout, "%s\n", path)
		return nil
	}
	return nil
}

func splitShell(s string) []string {
	fields := strings.Fields(s)
	return fields
}
