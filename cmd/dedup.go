package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var (
	dedupGroup     string
	dedupVGroup    string
	dedupReference string
	dedupApply     bool
)

// intraDup 记录组内 URL 重复条目
type intraDup struct {
	groupName string
	name      string
	url       string
}

// crossDup 记录跨组 URL 重复
type crossDup struct {
	normalizedURL string
	groups        []string
}

var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Deduplicate repos within and across groups",
	Long: `Detect and handle duplicate repositories:

  Step 1: Intra-group dedup - Remove repos with the same URL within a group (keep first)
  Step 2: Cross-group URL warnings - Warn when the same URL appears in multiple groups
  Step 3: Cross-group name dedup - Exclude repos by name from target group (only when --group + --reference)

By default runs in dry-run mode (shows what would change without modifying config).
Use --apply to actually write changes.`,
	Example: `  grepom dedup                                              # Check all groups for intra-group dupes and cross-group warnings
  grepom dedup --group core-team                           # Check only core-team
  grepom dedup --group core-team --reference infra-team    # Also exclude by name against infra-team
  grepom dedup --group core-team --reference infra,legacy  # Exclude by name against multiple groups
  grepom dedup --apply                                     # Apply all changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := loadConfig()
		if err != nil {
			return err
		}

		configPath, err := resolvedConfigPath()
		if err != nil {
			configPath = defaultConfigPath()
		}

		// 确定要处理的 groups
		groupSelection, err := cfg.ResolveGroupSelection(dedupGroup, dedupVGroup)
		if err != nil {
			return err
		}
		groupsToCheck, err := cfg.FilterGroups(dedupGroup, dedupVGroup)
		if err != nil {
			return err
		}
		if len(groupSelection) == 0 {
			groupsToCheck = cfg.Groups
		}

		// ═══════════════════════════════════════════
		// Step 1: 组内 URL 去重
		// ═══════════════════════════════════════════
		intraDups := detectIntraGroupDups(groupsToCheck)
		printIntraGroupDups(intraDups, groupsToCheck)

		// ═══════════════════════════════════════════
		// Step 2: 跨组 URL 警告
		// ═══════════════════════════════════════════
		crossDups := detectCrossGroupDups(cfg.Groups, groupSelection)
		printCrossGroupDups(crossDups)

		// ═══════════════════════════════════════════
		// Step 3: 跨组 name 去重（仅当 --group + --reference 同时指定时）
		// ═══════════════════════════════════════════
		var step3Dups []string
		if dedupGroup != "" && dedupReference != "" {
			step3Dups, err = runCrossNameDedup(cfg, configPath)
			if err != nil {
				return err
			}
		}

		// 检查是否有需要 --apply 的变更
		hasIntraChanges := len(intraDups) > 0
		hasStep3Changes := len(step3Dups) > 0

		if !hasIntraChanges && !hasStep3Changes && len(crossDups) == 0 {
			fmt.Println("no duplicates found")
			return nil
		}

		if !dedupApply && (hasIntraChanges || hasStep3Changes) {
			fmt.Println("\nNo changes written. Add --apply to execute.")
			return nil
		}

		// 执行写入
		if dedupApply {
			// Step 1 写入：组内去重
			for _, dup := range intraDups {
				removed, err := config.DedupIntraGroupRepos(configPath, dup.groupName)
				if err != nil {
					return fmt.Errorf("intra-group dedup for %s: %w", dup.groupName, err)
				}
				if len(removed) > 0 {
					fmt.Printf("\n%d duplicate repos removed from %s\n", len(removed), dup.groupName)
				}
			}

			// Step 3 写入：跨组 name 去重
			if len(step3Dups) > 0 {
				excluded, err := config.DedupGroupRepos(configPath, dedupGroup, step3Dups)
				if err != nil {
					return err
				}
				if len(excluded) > 0 {
					fmt.Printf("\n%d repos excluded in %s. Run 'grepom prune' to clean up cloned repos.\n", len(excluded), dedupGroup)
				}
			}
		}

		return nil
	},
}

// detectIntraGroupDups 检测指定 groups 中的组内 URL 重复
func detectIntraGroupDups(groups []config.Group) []intraDup {
	var dups []intraDup
	for _, g := range groups {
		seen := make(map[string]bool)
		for _, r := range g.Repos {
			norm := config.NormalizeRepoURL(r.URL)
			if seen[norm] {
				dups = append(dups, intraDup{
					groupName: g.Name,
					name:      r.Name,
					url:       r.URL,
				})
			} else {
				seen[norm] = true
			}
		}
	}
	return dups
}

// printIntraGroupDups 输出组内去重结果
func printIntraGroupDups(dups []intraDup, groups []config.Group) {
	fmt.Println("Intra-group dedup:")

	// 按 group 分组
	groupDups := make(map[string][]intraDup)
	for _, d := range dups {
		groupDups[d.groupName] = append(groupDups[d.groupName], d)
	}

	for _, g := range groups {
		if gd, ok := groupDups[g.Name]; ok {
			for _, d := range gd {
				fmt.Printf("  %s: %s (duplicate) → would remove\n", d.groupName, d.name)
			}
			fmt.Printf("  %s: %d duplicate(s) found\n", g.Name, len(gd))
		} else {
			fmt.Printf("  %s: no duplicates\n", g.Name)
		}
	}
}

// detectCrossGroupDups 检测跨组 URL 重复
func detectCrossGroupDups(allGroups []config.Group, filterGroups []string) []crossDup {
	// 构建 url → group names 映射
	urlToGroups := make(map[string][]string)
	urlToOriginal := make(map[string]string) // normalized → original URL (取第一个)
	for _, g := range allGroups {
		for _, r := range g.Repos {
			norm := config.NormalizeRepoURL(r.URL)
			if norm == "" {
				continue
			}
			if _, exists := urlToOriginal[norm]; !exists {
				urlToOriginal[norm] = r.URL
			}
			// 避免同一 group 内重复记录
			already := false
			for _, gn := range urlToGroups[norm] {
				if gn == g.Name {
					already = true
					break
				}
			}
			if !already {
				urlToGroups[norm] = append(urlToGroups[norm], g.Name)
			}
		}
	}

	var dups []crossDup
	for norm, groups := range urlToGroups {
		if len(groups) <= 1 {
			continue
		}
		// 如果指定了 filterGroups，只报告涉及这些 groups 的跨组重复
		if len(filterGroups) > 0 {
			found := false
			for _, g := range groups {
				for _, fg := range filterGroups {
					if g == fg {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}
		dups = append(dups, crossDup{
			normalizedURL: urlToOriginal[norm],
			groups:        groups,
		})
	}

	// 按 URL 排序以保证输出稳定
	sort.Slice(dups, func(i, j int) bool {
		return dups[i].normalizedURL < dups[j].normalizedURL
	})

	return dups
}

// printCrossGroupDups 输出跨组 URL 警告
func printCrossGroupDups(dups []crossDup) {
	if len(dups) == 0 {
		return
	}

	fmt.Println("\nCross-group URL warnings:")
	for _, d := range dups {
		fmt.Printf("  ⚠️  %s\n", d.normalizedURL)
		fmt.Printf("       appears in: %s\n", strings.Join(d.groups, ", "))
	}
}

// runCrossNameDedup 执行 Step 3 跨组按 name 去重（兼容原有逻辑）
func runCrossNameDedup(cfg *config.Config, configPath string) ([]string, error) {
	_, targetGroup, err := cfg.FindGroup(dedupGroup)
	if err != nil {
		return nil, err
	}

	// 解析 reference groups
	var refGroups []config.Group
	for _, name := range strings.Split(dedupReference, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		_, g, err := cfg.FindGroup(name)
		if err != nil {
			return nil, fmt.Errorf("reference group: %w", err)
		}
		refGroups = append(refGroups, *g)
	}

	if len(refGroups) == 0 {
		return nil, nil
	}

	// 收集 reference 的所有 repo names
	refNames := make(map[string]bool)
	for _, g := range refGroups {
		for _, r := range g.Repos {
			refNames[r.Name] = true
		}
	}

	if len(refNames) == 0 {
		return nil, nil
	}

	// 检测 target 中与 reference 同名的 repos
	type dupEntry struct {
		name string
		ref  string
	}
	var dups []dupEntry
	var kept int

	for _, r := range targetGroup.Repos {
		if refNames[r.Name] {
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
		return nil, nil
	}

	// 输出
	fmt.Println("\nCross-group name dedup:")
	fmt.Printf("  checking %s against %s\n\n", dedupGroup, refGroupNames(refGroups))

	var dupNames []string
	for _, d := range dups {
		fmt.Printf("  %s → exclude (exists in %s)\n", d.name, d.ref)
		dupNames = append(dupNames, d.name)
	}

	fmt.Printf("\n  %d repos excluded, %d kept\n", len(dups), kept)

	return dupNames, nil
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
	dedupCmd.Flags().StringVarP(&dedupGroup, "group", "g", "", "target group to check (optional, defaults to all groups)")
	dedupCmd.Flags().StringVarP(&dedupVGroup, "vgroup", "V", "", "virtual group to check (optional)")
	dedupCmd.Flags().StringVarP(&dedupReference, "reference", "r", "", "reference group(s), comma-separated (triggers cross-group name dedup with --group)")
	dedupCmd.Flags().BoolVar(&dedupApply, "apply", false, "apply changes (default is dry-run)")
	rootCmd.AddCommand(dedupCmd)
}
