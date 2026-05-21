package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// --- Git tag operations ---

// FetchTags fetches tags from all remotes.
func FetchTags(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "--tags", "--all")
	cmd.Stderr = &strings.Builder{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch --tags --all: %w", err)
	}
	return nil
}

// ListTagsByTime returns tags matching pattern, sorted by creation time descending (newest first).
func ListTagsByTime(path, pattern string) ([]string, error) {
	cmd := exec.Command("git", "-C", path, "tag", "--sort=-creatordate", "--list", pattern)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git tag --list: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	return strings.Split(raw, "\n"), nil
}

// TagExists checks whether a tag with the given name exists.
func TagExists(path, name string) (bool, error) {
	cmd := exec.Command("git", "-C", path, "tag", "-l", name)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git tag -l: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// CreateTag creates a lightweight tag.
func CreateTag(path, name string) error {
	cmd := exec.Command("git", "-C", path, "tag", name)
	cmd.Stderr = &strings.Builder{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git tag %s: %w", name, err)
	}
	return nil
}

// CreateAnnotatedTag creates an annotated tag with the given message.
func CreateAnnotatedTag(path, name, message string) error {
	cmd := exec.Command("git", "-C", path, "tag", "-a", name, "-m", message)
	cmd.Stderr = &strings.Builder{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git tag -a %s: %w", name, err)
	}
	return nil
}

// PushTag pushes a single tag to the specified remote.
func PushTag(path, remote, tag string) error {
	cmd := exec.Command("git", "-C", path, "push", remote, tag)
	cmd.Stdout = &strings.Builder{}
	cmd.Stderr = &strings.Builder{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push %s %s: %w", remote, tag, err)
	}
	return nil
}

// ListRemotes returns the names of all remotes.
func ListRemotes(path string) ([]string, error) {
	cmd := exec.Command("git", "-C", path, "remote")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git remote: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	return strings.Split(raw, "\n"), nil
}

// --- Version parsing and calculation ---

// ParseVersion parses a tag string, extracting the prefix (e.g. "v" or "t")
// and the numeric segments after the prefix.
// Non-numeric segments are ignored.
func ParseVersion(tag string) (prefix string, digits []int, err error) {
	if tag == "" {
		return "", nil, fmt.Errorf("empty tag")
	}

	// Extract prefix: first non-digit character(s) until first digit or dot
	prefix = ""
	i := 0
	for i < len(tag) && tag[i] >= 'a' && tag[i] <= 'z' {
		prefix += string(tag[i])
		i++
	}

	if prefix == "" {
		return "", nil, fmt.Errorf("tag %q has no letter prefix", tag)
	}

	rest := tag[i:]
	if rest == "" {
		return prefix, nil, nil
	}

	// Split by dot and parse each segment as int
	parts := strings.Split(rest, ".")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, e := strconv.Atoi(p)
		if e != nil {
			// Ignore non-numeric segments
			continue
		}
		digits = append(digits, n)
	}

	return prefix, digits, nil
}

// NormalizeDigits pads or truncates the digits slice to exactly size elements.
// If digits has more than size elements, only the first size are kept.
// If digits has fewer than size elements, zeros are appended.
func NormalizeDigits(digits []int, size int) []int {
	result := make([]int, size)
	for i := 0; i < size; i++ {
		if i < len(digits) {
			result[i] = digits[i]
		} else {
			result[i] = 0
		}
	}
	return result
}

// NextVPatch increments the PATCH version.
// If PATCH exceeds 99, it carries over to MINOR (MINOR has no upper limit).
func NextVPatch(major, minor, patch int) (int, int, int) {
	patch++
	if patch > 99 {
		minor++
		patch = 0
	}
	return major, minor, patch
}

// FormatVTag formats a v-prefixed version tag: v{MAJOR}.{MINOR}.{PATCH}
func FormatVTag(major, minor, patch int) string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

// FormatTTag formats a t-prefixed version tag: t{MAJOR}.{MINOR}.{PATCH}.{ITER}
func FormatTTag(major, minor, patch, iter int) string {
	return fmt.Sprintf("t%d.%d.%d.%d", major, minor, patch, iter)
}
