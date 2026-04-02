package git

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/wii/grepom/provider"
)

// CloneTask represents a single clone operation to be executed.
type CloneTask struct {
	Repo     provider.Repo
	FullPath string
}

// CloneResult represents the result of a single clone operation.
type CloneResult struct {
	Repo     provider.Repo
	FullPath string
	Err      error
	Log      string // clone attempt log text (collected via LogWriter)
}

// CloneAll clones multiple repositories in parallel using a worker pool.
// concurrency controls the number of parallel workers (must be >= 1).
func CloneAll(concurrency int, tasks []CloneTask) []CloneResult {
	if concurrency < 1 {
		concurrency = 1
	}
	if len(tasks) == 0 {
		return nil
	}

	jobs := make(chan CloneTask, len(tasks))
	results := make(chan CloneResult, len(tasks))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				var buf bytes.Buffer
				opts := CloneOptions{
					Token:          task.Repo.Token,
					Provider:       task.Repo.Provider,
					SSHKey:         task.Repo.SSHKey,
					HasGroupToken:  task.Repo.HasGroupToken,
					HasGroupSSHKey: task.Repo.HasGroupSSHKey,
					LogWriter:      &buf,
				}
				err := Clone(task.FullPath, task.Repo.SSHURL, task.Repo.CloneURL, opts)
				results <- CloneResult{
					Repo:     task.Repo,
					FullPath: task.FullPath,
					Err:      err,
					Log:      buf.String(),
				}
			}
		}()
	}

	// Send jobs
	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results preserving order
	resultMap := make(map[string]CloneResult, len(tasks))
	for r := range results {
		resultMap[r.FullPath] = r
	}

	ordered := make([]CloneResult, 0, len(tasks))
	for _, task := range tasks {
		ordered = append(ordered, resultMap[task.FullPath])
	}
	return ordered
}

// PullTask represents a single pull operation to be executed.
type PullTask struct {
	Repo     provider.Repo
	FullPath string
}

// PullResult represents the result of a single pull operation.
type PullResult struct {
	Repo       provider.Repo
	FullPath   string
	Err        error
	Skipped    bool
	SkipReason string
}

// PullAll pulls multiple repositories in parallel using a worker pool.
// concurrency controls the number of parallel workers (must be >= 1).
func PullAll(concurrency int, tasks []PullTask) []PullResult {
	if concurrency < 1 {
		concurrency = 1
	}
	if len(tasks) == 0 {
		return nil
	}

	jobs := make(chan PullTask, len(tasks))
	results := make(chan PullResult, len(tasks))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				err := Pull(task.FullPath)
				results <- PullResult{
					Repo:     task.Repo,
					FullPath: task.FullPath,
					Err:      err,
				}
			}
		}()
	}

	// Send jobs
	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results preserving order
	resultMap := make(map[string]PullResult, len(tasks))
	for r := range results {
		resultMap[r.FullPath] = r
	}

	ordered := make([]PullResult, 0, len(tasks))
	for _, task := range tasks {
		ordered = append(ordered, resultMap[task.FullPath])
	}
	return ordered
}

// CheckPullSafety checks whether a repo is eligible for a safe pull.
// Returns (eligible, skipReason).
// A repo is eligible when: cloned, on default branch, and clean.
func CheckPullSafety(fullPath string) (bool, string) {
	if !IsCloned(fullPath) {
		return false, "not cloned"
	}

	status := GetStatus(fullPath)
	if status.NotARepo {
		return false, "not a git repository"
	}

	// Check clean
	if !status.Clean {
		return false, "dirty working tree"
	}

	// Check default branch
	defaultBranch, err := GetDefaultBranch(fullPath)
	if err != nil {
		return false, "cannot detect default branch"
	}
	if status.Branch != defaultBranch {
		return false, fmt.Sprintf("on %s, not default branch", status.Branch)
	}

	return true, ""
}
