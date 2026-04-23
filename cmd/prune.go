package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	pruneGroup    string
	pruneResource string
	pruneForce    bool
	pruneApply    bool
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove cloned repos that are excluded by exclude_repos",
	Long: `Scan for repositories that are excluded via exclude_repos but still cloned on disk,
and optionally remove them.

By default runs in dry-run mode (shows what would be deleted without deleting).
Use --apply to actually delete. Use --force to skip safety checks for dirty/ahead repos.`,
	Example: `  grepom prune                          # Dry-run: show what would be deleted
  grepom prune --apply                  # Delete excluded repos (skip dirty/ahead)
  grepom prune --apply --force          # Delete all excluded repos, even dirty/ahead
  grepom prune --group frontend         # Only prune repos in the frontend group
  grepom prune --resource work-gl       # Only prune repos from the work-gl resource`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Group:           pruneGroup,
			Resource:        pruneResource,
			IncludeDisabled: true,
		}

		resolver := repo.NewResolver(cfg)
		allRepos, err := resolver.ResolveAndFilter(filter)
		if err != nil {
			return err
		}

		// 只保留 excluded 的 repos
		var excludedRepos []gitpkg.CloneTask
		for _, r := range allRepos {
			if r.DisabledReason == "excluded" {
				fullPath := repo.FullPath(cfg.Base, r)
				excludedRepos = append(excludedRepos, gitpkg.CloneTask{
					Repo:     r,
					FullPath: fullPath,
				})
			}
		}

		if len(excludedRepos) == 0 {
			fmt.Println("no excluded repos to prune")
			return nil
		}

		// 分类每个 excluded repo
		type pruneEntry struct {
			task     gitpkg.CloneTask
			category string // "safe_delete", "unsafe_delete", "not_cloned"
			reason   string
		}

		var entries []pruneEntry
		for _, t := range excludedRepos {
			if !gitpkg.IsCloned(t.FullPath) {
				entries = append(entries, pruneEntry{task: t, category: "not_cloned"})
				continue
			}

			if pruneForce {
				entries = append(entries, pruneEntry{task: t, category: "safe_delete"})
				continue
			}

			st := gitpkg.GetStatus(t.FullPath)
			if !st.Clean {
				entries = append(entries, pruneEntry{
					task:     t,
					category: "unsafe_delete",
					reason:   fmt.Sprintf("dirty, %d files changed", st.Dirty),
				})
			} else if st.Ahead > 0 {
				entries = append(entries, pruneEntry{
					task:     t,
					category: "unsafe_delete",
					reason:   fmt.Sprintf("ahead %d commits", st.Ahead),
				})
			} else {
				entries = append(entries, pruneEntry{task: t, category: "safe_delete"})
			}
		}

		// 输出计划
		var toDelete []pruneEntry
		var skipped []pruneEntry
		var notCloned int

		for _, e := range entries {
			switch e.category {
			case "safe_delete":
				toDelete = append(toDelete, e)
				if pruneApply {
					fmt.Printf("  deleted: %s\n", e.task.Repo.Name)
				} else {
					fmt.Printf("  will delete: %s (clean)\n", e.task.Repo.Name)
				}
			case "unsafe_delete":
				skipped = append(skipped, e)
				fmt.Printf("  skipped: %s (%s)\n", e.task.Repo.Name, e.reason)
			case "not_cloned":
				notCloned++
				if verbose {
					fmt.Printf("  not cloned: %s\n", e.task.Repo.Name)
				}
			}
		}

		// 执行删除
		if pruneApply && len(toDelete) > 0 {
			for _, e := range toDelete {
				if err := os.RemoveAll(e.task.FullPath); err != nil {
					fmt.Fprintf(os.Stderr, "error removing %s: %v\n", e.task.Repo.Name, err)
				}
			}
		}

		// 输出摘要
		fmt.Println()
		deleted := len(toDelete)
		if pruneApply {
			fmt.Printf("prune: %d deleted, %d skipped, %d not cloned\n", deleted, len(skipped), notCloned)
		} else {
			fmt.Printf("prune: %d would delete, %d skipped, %d not cloned\n", deleted, len(skipped), notCloned)
			if deleted > 0 {
				fmt.Println("No files deleted. Add --apply to execute.")
			}
		}

		return nil
	},
}

func init() {
	pruneCmd.Flags().StringVarP(&pruneGroup, "group", "g", "", "only prune repos in a specific group")
	pruneCmd.Flags().StringVarP(&pruneResource, "resource", "R", "", "only prune repos from a specific resource")
	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "skip safety checks (delete dirty/ahead repos)")
	pruneCmd.Flags().BoolVar(&pruneApply, "apply", false, "actually delete files (default is dry-run)")
	rootCmd.AddCommand(pruneCmd)
}
