package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
)

var (
	tagTest   bool
	tagPush   bool
	tagDryRun bool
	tagWatch  bool
	tagMsg    string
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create version tags",
	Long: `Automatically calculate and create the next version tag in the current git repository.

Supports two version formats:
  - v version (default): vMAJOR.MINOR.PATCH (e.g. v0.1.2)
  - t version (-t flag): tMAJOR.MINOR.PATCH.ITER (e.g. t0.1.2.3)

Use -w/--watch to automatically watch the CI/CD pipeline after creating the tag.
This is a convenience shortcut for "tag && watch" — the pipeline is auto-detected
from the current git repository using the same 3-level fallback as "grepom watch".

This command does not depend on grepom config files and can be used in any git repository.`,
	Example: `  grepom tag                      # v0.1.5 -> v0.1.6
  grepom tag -t                   # t0.1.6.0 (first t version)
  grepom tag -p                   # create v and push to all remotes
  grepom tag -t -p                # create t and push
  grepom tag -w                   # create v, then watch pipeline
  grepom tag -w -p                # create v, push, then watch pipeline
  grepom tag --dry-run            # preview only
  grepom tag -m "release v0.2.0"  # annotated tag`,
	RunE: runTag,
}

func init() {
	tagCmd.Flags().BoolVarP(&tagTest, "test", "t", false, "create t-prefix version (test release)")
	tagCmd.Flags().BoolVarP(&tagPush, "push", "p", false, "push tag to all remotes after creation")
	tagCmd.Flags().BoolVarP(&tagWatch, "watch", "w", false, "watch pipeline after tag creation")
	tagCmd.Flags().BoolVar(&tagDryRun, "dry-run", false, "preview what would be created, without actually creating")
	tagCmd.Flags().StringVarP(&tagMsg, "message", "m", "", "message for annotated tag (default: lightweight tag)")
	rootCmd.AddCommand(tagCmd)
}

const maxConflictRetries = 10000

// runTag is the main entry point for the tag command.
func runTag(cmd *cobra.Command, args []string) error {
	// Pre-check: current directory must be a git repo
	if !gitpkg.IsCloned(".") {
		return fmt.Errorf("current directory is not a git repository")
	}

	// Fetch latest tags from all remotes
	fmt.Print("Fetching tags...")
	if err := gitpkg.FetchTags("."); err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}
	fmt.Println(" done")

	// Get v* tags sorted by time (newest first)
	vTags, err := gitpkg.ListTagsByTime(".", "v*")
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	// Parse the latest v tag or use default (0, 0, 0)
	var major, minor, patch int
	var latestVTag string
	if len(vTags) > 0 {
		latestVTag = vTags[0]
		_, digits, err := gitpkg.ParseVersion(latestVTag)
		if err != nil {
			return fmt.Errorf("failed to parse latest tag %q: %w", latestVTag, err)
		}
		normalized := gitpkg.NormalizeDigits(digits, 3)
		major, minor, patch = normalized[0], normalized[1], normalized[2]
	}

	// Compute next v version
	major, minor, patch = gitpkg.NextVPatch(major, minor, patch)

	var tagErr error
	if tagTest {
		// t version: first 3 digits follow v calculation, 4th digit is independent
		tagErr = runTTag(major, minor, patch, latestVTag)
	} else {
		// v version
		tagErr = runVTag(major, minor, patch, latestVTag)
	}

	// If tag creation failed, return the error without watching
	if tagErr != nil {
		return tagErr
	}

	// If --dry-run was used, tag was not actually created — skip watch
	if tagDryRun {
		return nil
	}

	// If -w/--watch is set, watch the latest pipeline after tag creation
	if tagWatch {
		target, resolveErr := resolveCurrentRepoPipeline()
		if resolveErr != nil {
			fmt.Printf("\nTag created successfully, but failed to auto-detect pipeline:\n%v\n", resolveErr)
			return resolveErr
		}
		return runWatchLoop(target, 0, cmd)
	}

	return nil
}

// runVTag computes, creates, and optionally pushes a v tag.
func runVTag(major, minor, patch int, latestVTag string) error {
	newTag := gitpkg.FormatVTag(major, minor, patch)

	// Conflict detection: auto-increment until we find an unused tag
	for i := 0; i < maxConflictRetries; i++ {
		exists, err := gitpkg.TagExists(".", newTag)
		if err != nil {
			return fmt.Errorf("failed to check tag existence: %w", err)
		}
		if !exists {
			break
		}
		major, minor, patch = gitpkg.NextVPatch(major, minor, patch)
		newTag = gitpkg.FormatVTag(major, minor, patch)
	}

	// Display info
	if latestVTag != "" {
		fmt.Printf("Latest v tag: %s\n", latestVTag)
	} else {
		fmt.Println("No existing v tags found")
	}
	fmt.Printf("New tag:      %s\n", newTag)

	// Dry-run mode
	if tagDryRun {
		fmt.Printf("[dry-run] Would create tag %s locally\n", newTag)
		return nil
	}

	// Create tag
	if err := createTag(newTag); err != nil {
		return err
	}

	fmt.Printf("✓ created locally\n")

	// Handle push
	return handlePush(newTag)
}

// runTTag computes, creates, and optionally pushes a t tag.
func runTTag(major, minor, patch int, latestVTag string) error {
	// The first 3 digits follow v calculation
	// Find matching t{M}.{m}.{p}.* tags to determine 4th digit
	iter := 0

	pattern := fmt.Sprintf("t%d.%d.%d.*", major, minor, patch)
	tTags, err := gitpkg.ListTagsByTime(".", pattern)
	if err != nil {
		return fmt.Errorf("failed to list t tags: %w", err)
	}

	if len(tTags) > 0 {
		// Parse the newest matching t tag to get the 4th digit
		_, digits, err := gitpkg.ParseVersion(tTags[0])
		if err == nil && len(digits) >= 4 {
			iter = digits[3] + 1
		}
	}

	newTag := gitpkg.FormatTTag(major, minor, patch, iter)

	// Conflict detection: auto-increment 4th digit
	for i := 0; i < maxConflictRetries; i++ {
		exists, err := gitpkg.TagExists(".", newTag)
		if err != nil {
			return fmt.Errorf("failed to check tag existence: %w", err)
		}
		if !exists {
			break
		}
		iter++
		newTag = gitpkg.FormatTTag(major, minor, patch, iter)
	}

	// Display info
	if latestVTag != "" {
		fmt.Printf("Latest v tag: %s\n", latestVTag)
	} else {
		fmt.Println("No existing v tags found")
	}
	fmt.Printf("New tag:      %s\n", newTag)

	// Dry-run mode
	if tagDryRun {
		fmt.Printf("[dry-run] Would create tag %s locally\n", newTag)
		return nil
	}

	// Create tag
	if err := createTag(newTag); err != nil {
		return err
	}

	fmt.Printf("✓ created locally\n")

	// Handle push
	return handlePush(newTag)
}

// createTag creates either a lightweight or annotated tag.
func createTag(name string) error {
	if tagMsg != "" {
		return gitpkg.CreateAnnotatedTag(".", name, tagMsg)
	}
	return gitpkg.CreateTag(".", name)
}

// handlePush manages the push-to-remotes behavior:
// -p flag: push directly
// TTY + no -p: ask user
// No TTY + no -p: show hint
func handlePush(tag string) error {
	if tagDryRun {
		return nil
	}

	if tagPush {
		fmt.Println("Pushing to all remotes...")
		return pushToAllRemotes(tag)
	}

	if isTerminal() {
		confirm := false
		err := survey.AskOne(&survey.Confirm{
			Message: "Push to all remotes?",
			Default: false,
		}, &confirm)
		if err == terminal.InterruptErr {
			fmt.Println("\nCancelled.")
			return nil
		}
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		if confirm {
			return pushToAllRemotes(tag)
		}
	} else {
		fmt.Println("(no TTY, skipping push prompt. Use -p to push.)")
	}

	return nil
}

// pushToAllRemotes pushes the tag to every configured remote.
func pushToAllRemotes(tag string) error {
	remotes, err := gitpkg.ListRemotes(".")
	if err != nil {
		return fmt.Errorf("failed to list remotes: %w", err)
	}

	if len(remotes) == 0 {
		fmt.Println("No remotes configured")
		return nil
	}

	var failures []string
	for _, remote := range remotes {
		err := gitpkg.PushTag(".", strings.TrimSpace(remote), tag)
		if err != nil {
			fmt.Printf("  ✗ %-10s failed: %v\n", remote, err)
			failures = append(failures, remote)
		} else {
			fmt.Printf("  ✓ %-10s pushed\n", remote)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("failed to push to remotes: %s", strings.Join(failures, ", "))
	}
	return nil
}

// parseIterFromTag extracts the 4th digit from a t tag string.
// Returns 0 if parsing fails.
func parseIterFromTag(tag string) int {
	_, digits, err := gitpkg.ParseVersion(tag)
	if err != nil {
		return 0
	}
	normalized := gitpkg.NormalizeDigits(digits, 4)
	return normalized[3]
}

// formatInt is a simple wrapper for strconv.Itoa.
func formatInt(n int) string {
	return strconv.Itoa(n)
}
