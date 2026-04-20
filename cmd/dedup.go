package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var (
	dedupGroup     string
	dedupReference string
	dedupApply     bool
)

var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Deduplicate repos across groups by name",
	Long: `Detect repos with the same name across groups and exclude them from the target group.

For each repo in the target group that shares a name with a repo in the reference group(s),
the repo is removed from the target group's repos list and added to its exclude_repos.

By default runs in dry-run mode (shows what would change without modifying config).
Use --apply to actually write changes.`,
	Example: `  grepom dedup --group core-team                        # Dedup core-team against all other groups
  grepom dedup --group core-team --reference infra-team  # Dedup core-team against infra-team only
  grepom dedup --group core-team --reference infra,legacy # Dedup against multiple groups
  grepom dedup --group core-team --apply                 # Apply changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		configPath, err := resolvedConfigPath()
		if err != nil {
			configPath = configFile
		}

		// 找到 target group
		targetIdx, targetGroup, err := cfg.FindGroup(dedupGroup)
		if err != nil {
			return err
		}

		// 解析 reference groups
		var refGroups []config.Group
		if dedupReference != "" {
			for _, name := range strings.Split(dedupReference, ",") {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				_, g, err := cfg.FindGroup(name)
				if err != nil {
					return fmt.Errorf("reference group: %w", err)
				}
				refGroups = append(refGroups, *g)
			}
		} else {
			// 不指定时对比所有其他 group
			for i, g := range cfg.Groups {
				if i != targetIdx {
					refGroups = append(refGroups, g)
				}
			}
		}

		if len(refGroups) == 0 {
			fmt.Println("no reference groups to compare against")
			return nil
		}

		// 收集 reference 的所有 repo names
		refNames := make(map[string]bool)
		for _, g := range refGroups {
			for _, r := range g.Repos {
				refNames[r.Name] = true
			}
		}

		if len(refNames) == 0 {
			fmt.Println("no repos found in reference groups")
			return nil
		}

		// 检测 target 中与 reference 同名的 repos
		type dupEntry struct {
			name string
			ref  string // 来自哪个 reference group
		}
		var dups []dupEntry
		var kept int

		for _, r := range targetGroup.Repos {
			if refNames[r.Name] {
				// 找到这个 name 来自哪个 reference group（用于输出）
				refSource := ""
				for _, g := range refGroups {
					for _, gr := range g.Repos {
						if gr.Name == r.Name {
							refSource = g.Name
							break
						}
					}
					if refSource != "" {
						break
					}
				}
				dups = append(dups, dupEntry{name: r.Name, ref: refSource})
			} else {
				kept++
			}
		}

		if len(dups) == 0 {
			fmt.Println("no duplicates found")
			return nil
		}

		// 输出去重计划
		fmt.Printf("dedup: checking %s against %s\n\n", dedupGroup, refGroupNames(refGroups))

		var dupNames []string
		for _, d := range dups {
			fmt.Printf("  %s → exclude (exists in %s)\n", d.name, d.ref)
			dupNames = append(dupNames, d.name)
		}

		fmt.Printf("\n%d repos excluded, %d kept\n", len(dups), kept)

		if !dedupApply {
			fmt.Println("No changes written. Add --apply to execute.")
			return nil
		}

		// 执行写入
		excluded, err := config.DedupGroupRepos(configPath, dedupGroup, dupNames)
		if err != nil {
			return err
		}

		if len(excluded) > 0 {
			fmt.Printf("\n%d repos excluded in %s. Run 'grepom prune' to clean up cloned repos.\n", len(excluded), dedupGroup)
		}

		return nil
	},
}

// refGroupNames 返回 reference group 名称的逗号分隔字符串
func refGroupNames(groups []config.Group) string {
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}
	return strings.Join(names, ", ")
}

func init() {
	dedupCmd.Flags().StringVarP(&dedupGroup, "group", "g", "", "target group to dedup (required)")
	dedupCmd.Flags().StringVarP(&dedupReference, "reference", "r", "", "reference group(s), comma-separated (default: all other groups)")
	dedupCmd.Flags().BoolVar(&dedupApply, "apply", false, "apply changes (default is dry-run)")
	dedupCmd.MarkFlagRequired("group")
	rootCmd.AddCommand(dedupCmd)
}
