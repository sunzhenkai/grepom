package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/provider"
	"github.com/wii/grepom/repo"
)

var (
	pullGroup       string
	pullResource    string
	pullConcurrency int
	pullForce       bool
)

var pullCmd = &cobra.Command{
	Use:   "pull [name]",
	Short: "Pull updates for repositories",
	Long: `Run git pull on cloned repositories. By default, pull only runs on repos that are
on their default branch and have a clean working tree. Use --force to skip safety checks.`,
	Example: `  grepom pull                              # Pull eligible repos (parallel, 4 workers)
  grepom pull --concurrency 1                # Sequential pull
  grepom pull --force                       # Pull all cloned repos (skip safety checks)
  grepom pull web-app                        # Pull a specific repo
  grepom pull --group frontend               # Pull repos in a group`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pullConcurrency < 1 {
			return fmt.Errorf("--concurrency must be a positive integer, got %d", pullConcurrency)
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Group:    pullGroup,
			Resource: pullResource,
		}
		if len(args) > 0 {
			filter.Name = args[0]
		}

		resolver := repo.NewResolver(cfg)
		repos, err := resolver.ResolveAndFilter(filter)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories found.")
			return nil
		}

		if pullForce {
			// --force: skip safety checks, pull all cloned repos
			return runForcePull(cfg, repos)
		}

		// Smart pull with safety checks
		return runSmartPull(cfg, repos)
	},
}

// runSmartPull performs safety checks and only pulls eligible repos.
func runSmartPull(cfg *config.Config, repos []provider.Repo) error {
	// Phase 1: Safety check - determine which repos are eligible
	var toPull []gitpkg.PullTask
	skipped := 0

	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)

		if !gitpkg.IsCloned(fullPath) {
			if verbose {
				fmt.Printf("skip %s (not cloned)\n", r.Path)
			}
			skipped++
			continue
		}

		ok, reason := gitpkg.CheckPullSafety(fullPath)
		if !ok {
			fmt.Printf("skip %s (%s)\n", r.Path, reason)
			skipped++
			continue
		}

		toPull = append(toPull, gitpkg.PullTask{
			Repo:     r,
			FullPath: fullPath,
		})
	}

	if len(toPull) == 0 {
		if skipped > 0 {
			fmt.Printf("nothing to pull: %d skipped (not on default branch or dirty)\n", skipped)
		} else {
			fmt.Println("no repositories to pull")
		}
		return nil
	}

	// Phase 2: Execute pull
	if pullConcurrency > 1 && len(toPull) > 1 {
		return runParallelPull(toPull, skipped)
	}

	return runSequentialPull(toPull, skipped)
}

// runForcePull pulls all cloned repos without safety checks.
func runForcePull(cfg *config.Config, repos []provider.Repo) error {
	var toPull []gitpkg.PullTask
	skipped := 0

	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		if !gitpkg.IsCloned(fullPath) {
			fmt.Printf("skip %s (not cloned)\n", r.Path)
			skipped++
			continue
		}
		toPull = append(toPull, gitpkg.PullTask{
			Repo:     r,
			FullPath: fullPath,
		})
	}

	if len(toPull) == 0 {
		fmt.Println("no repositories to pull")
		return nil
	}

	if pullConcurrency > 1 && len(toPull) > 1 {
		return runParallelPull(toPull, skipped)
	}

	return runSequentialPull(toPull, skipped)
}

func runParallelPull(tasks []gitpkg.PullTask, skipped int) error {
	progress := NewProgressRenderer("pulling", len(tasks))
	defer progress.Done()

	results := gitpkg.PullAll(pullConcurrency, tasks)

	progress.Update(len(results))

	if !progress.isTTY {
		for _, r := range results {
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "error pulling %s: %v\n", r.Repo.Path, r.Err)
			}
		}
	}

	PrintPullSummary(results, skipped, nil)
	return nil
}

func runSequentialPull(tasks []gitpkg.PullTask, skipped int) error {
	for _, task := range tasks {
		fmt.Printf("pulling %s...\n", task.Repo.Path)
		if err := gitpkg.Pull(task.FullPath); err != nil {
			fmt.Fprintf(os.Stderr, "error pulling %s: %v\n", task.Repo.Path, err)
			continue
		}
	}

	// Build results for summary
	results := make([]gitpkg.PullResult, 0, len(tasks))
	for _, task := range tasks {
		results = append(results, gitpkg.PullResult{
			Repo:     task.Repo,
			FullPath: task.FullPath,
		})
	}
	PrintPullSummary(results, skipped, nil)
	return nil
}

func init() {
	pullCmd.Flags().StringVar(&pullGroup, "group", "", "filter by group name")
	pullCmd.Flags().StringVar(&pullResource, "resource", "", "filter by resource name")
	pullCmd.Flags().IntVarP(&pullConcurrency, "concurrency", "j", 4, "number of parallel pull workers")
	pullCmd.Flags().BoolVar(&pullForce, "force", false, "skip safety checks, pull all cloned repos")
	rootCmd.AddCommand(pullCmd)
}
