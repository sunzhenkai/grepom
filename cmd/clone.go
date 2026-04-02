package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	cloneGroup       string
	cloneResource    string
	cloneConcurrency int
)

var cloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone repositories to local filesystem",
	Long:  "Clone repositories from configured groups and standalone repos. Repositories are cloned preserving directory hierarchy.",
	Example: `  grepom clone                              # Clone all repos (parallel, 4 workers)
  grepom clone --concurrency 1                # Sequential clone (backward compatible)
  grepom clone --concurrency 8                # Parallel clone with 8 workers
  grepom clone web-app                        # Clone a specific repo
  grepom clone --group frontend               # Clone all repos in a group
  grepom clone --resource work-gl             # Clone all repos from a resource`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cloneConcurrency < 1 {
			return fmt.Errorf("--concurrency must be a positive integer, got %d", cloneConcurrency)
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{Group: cloneGroup, Resource: cloneResource}
		if len(args) > 0 {
			filter.Name = args[0]
		}

		resolver := repo.NewResolver(cfg)
		repos, err := resolver.ResolveAndFilter(filter)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories to clone.")
			return nil
		}

		// Filter out already-cloned repos
		var toClone []gitpkg.CloneTask
		for _, r := range repos {
			fullPath := repo.FullPath(cfg.Base, r)
			if gitpkg.IsCloned(fullPath) {
				if verbose {
					fmt.Printf("skip %s (already cloned)\n", r.Path)
				}
				continue
			}
			toClone = append(toClone, gitpkg.CloneTask{
				Repo:     r,
				FullPath: fullPath,
			})
		}

		if len(toClone) == 0 {
			fmt.Println("all repositories already cloned")
			return nil
		}

		if cloneConcurrency > 1 && len(toClone) > 1 {
			// Parallel clone with progress
			return runParallelClone(toClone)
		}

		// Sequential clone (backward compatible)
		return runSequentialClone(toClone)
	},
}

func runParallelClone(tasks []gitpkg.CloneTask) error {
	progress := NewProgressRenderer("cloning", len(tasks))
	defer progress.Done()

	results := gitpkg.CloneAll(cloneConcurrency, tasks, func(completed, total int) {
		progress.Update(completed)
	})

	// Print individual results in non-TTY mode
	if !progress.isTTY {
		for _, r := range results {
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "error cloning %s: %v\n", r.Repo.Path, r.Err)
			}
		}
	}

	PrintCloneSummary(results, nil)
	return nil
}

func runSequentialClone(tasks []gitpkg.CloneTask) error {
	for _, task := range tasks {
		fmt.Printf("cloning %s...\n", task.Repo.Path)
		opts := gitpkg.CloneOptions{
			Token:          task.Repo.Token,
			Provider:       task.Repo.Provider,
			SSHKey:         task.Repo.SSHKey,
			HasGroupToken:  task.Repo.HasGroupToken,
			HasGroupSSHKey: task.Repo.HasGroupSSHKey,
		}
		if err := gitpkg.Clone(task.FullPath, task.Repo.SSHURL, task.Repo.CloneURL, opts); err != nil {
			fmt.Fprintf(os.Stderr, "error cloning %s: %v\n", task.Repo.Path, err)
			continue
		}
		fmt.Printf("  %s done\n", task.Repo.Name)
	}
	return nil
}

func init() {
	cloneCmd.Flags().StringVarP(&cloneGroup, "group", "g", "", "clone all repos under a group")
	cloneCmd.Flags().StringVarP(&cloneResource, "resource", "R", "", "clone all repos from a resource")
	cloneCmd.Flags().IntVarP(&cloneConcurrency, "concurrency", "j", 4, "number of parallel clone workers")
	rootCmd.AddCommand(cloneCmd)
}
