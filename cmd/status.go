package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	statusGroup    string
	statusResource string
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show git status for repositories",
	Long:  "Show git status summary and per-repo status for cloned repositories.",
	Example: `  grepom status           # Status of all repos
  grepom status web-app    # Status of a specific repo
  grepom status --group frontend  # Status of repos in a group`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Group:    statusGroup,
			Resource: statusResource,
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

		// Collect status data for all repos
		type repoStatus struct {
			name     string
			label    string // display status label (only highest priority)
			fullPath string
		}

		var entries []repoStatus
		counts := struct {
			total, clean, dirty, ahead, behind, notCloned int
		}{}

		for _, r := range repos {
			fullPath := repo.FullPath(cfg.Base, r)
			st := gitpkg.GetStatus(fullPath)

			if !st.Cloned {
				counts.notCloned++
				entries = append(entries, repoStatus{
					name:     r.Name,
					label:    "not cloned",
					fullPath: fullPath,
				})
				continue
			}

			if st.NotARepo {
				// Treat not-a-repo as a special case; count as not cloned
				counts.notCloned++
				entries = append(entries, repoStatus{
					name:     r.Name,
					label:    "not a git repository",
					fullPath: fullPath,
				})
				continue
			}

			// Determine status label by priority: dirty > ahead > behind > clean
			var label string
			if !st.Clean {
				label = fmt.Sprintf("dirty (%d)", st.Dirty)
				counts.dirty++
			} else if st.Ahead > 0 {
				label = fmt.Sprintf("ahead %d", st.Ahead)
				counts.ahead++
			} else if st.Behind > 0 {
				label = fmt.Sprintf("behind %d", st.Behind)
				counts.behind++
			} else {
				label = "clean"
				counts.clean++
			}

			entries = append(entries, repoStatus{
				name:     r.Name,
				label:    label,
				fullPath: fullPath,
			})
		}

		counts.total = len(repos)

		// Print summary line
		var parts []string
		parts = append(parts, fmt.Sprintf("%d repos:", counts.total))
		parts = append(parts, fmt.Sprintf("%d clean", counts.clean))
		if counts.dirty > 0 {
			parts = append(parts, fmt.Sprintf("%d dirty", counts.dirty))
		}
		if counts.ahead > 0 {
			parts = append(parts, fmt.Sprintf("%d ahead", counts.ahead))
		}
		if counts.behind > 0 {
			parts = append(parts, fmt.Sprintf("%d behind", counts.behind))
		}
		var summary strings.Builder
		summary.WriteString(parts[0])
		for _, p := range parts[1:] {
			summary.WriteString(", ")
			summary.WriteString(p)
		}
		if counts.notCloned > 0 {
			fmt.Fprintf(&summary, " · %d not cloned", counts.notCloned)
		}
		fmt.Println(summary.String())

		// Print repo list
		fmt.Println()
		// Calculate column widths
		nameWidth := 0
		labelWidth := 0
		for _, e := range entries {
			if len(e.name) > nameWidth {
				nameWidth = len(e.name)
			}
			if len(e.label) > labelWidth {
				labelWidth = len(e.label)
			}
		}

		for _, e := range entries {
			fmt.Printf("  %-*s   %-*s   %s\n", nameWidth, e.name, labelWidth, e.label, e.fullPath)
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusGroup, "group", "", "filter by group name")
	statusCmd.Flags().StringVar(&statusResource, "resource", "", "filter by resource name")
	rootCmd.AddCommand(statusCmd)
}
