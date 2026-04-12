package git

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/wii/grepom/provider"
)

// ProgressEventType 表示进度事件的类型。
type ProgressEventType int

const (
	ProgressStart    ProgressEventType = iota // 任务开始处理
	ProgressComplete                          // 任务处理完成
)

// ProgressEvent 表示一个进度事件，包含事件类型、仓库名、已完成数、总数和错误信息。
type ProgressEvent struct {
	Type      ProgressEventType
	RepoName  string // 正在处理的仓库名
	Completed int    // 已完成的任务数
	Total     int    // 总任务数
	Err       error  // 仅 Complete 事件，nil 表示成功
}

// ProgressFunc is called when a parallel task starts or completes.
// The event contains the type (Start/Complete), repo name, and progress counts.
type ProgressFunc func(ProgressEvent)

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
// onProgress is called after each repository clone completes (may be nil).
func CloneAll(concurrency int, tasks []CloneTask, onProgress ProgressFunc) []CloneResult {
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
				// 通知任务开始
				if onProgress != nil {
					onProgress(ProgressEvent{
						Type:     ProgressStart,
						RepoName: task.Repo.Name,
					})
				}
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

	// Collect results preserving order, invoke progress callback
	resultMap := make(map[string]CloneResult, len(tasks))
	completed := 0
	for r := range results {
		resultMap[r.FullPath] = r
		completed++
		if onProgress != nil {
			onProgress(ProgressEvent{
				Type:      ProgressComplete,
				RepoName:  r.Repo.Name,
				Completed: completed,
				Total:     len(tasks),
				Err:       r.Err,
			})
		}
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
// onProgress is called after each repository pull completes (may be nil).
func PullAll(concurrency int, tasks []PullTask, onProgress ProgressFunc) []PullResult {
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
				// 通知任务开始
				if onProgress != nil {
					onProgress(ProgressEvent{
						Type:     ProgressStart,
						RepoName: task.Repo.Name,
					})
				}
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

	// Collect results preserving order, invoke progress callback
	resultMap := make(map[string]PullResult, len(tasks))
	completed := 0
	for r := range results {
		resultMap[r.FullPath] = r
		completed++
		if onProgress != nil {
			onProgress(ProgressEvent{
				Type:      ProgressComplete,
				RepoName:  r.Repo.Name,
				Completed: completed,
				Total:     len(tasks),
				Err:       r.Err,
			})
		}
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
