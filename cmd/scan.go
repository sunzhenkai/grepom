package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
	"github.com/wii/grepom/scanner"
)

var (
	scanGroup       string
	scanResource    string
	scanHistory     bool
	scanFormat      string
	scanGitleaksCfg string
	scanOutput      string
)

var scanCmd = &cobra.Command{
	Use:   "scan [name]",
	Short: "扫描仓库中的敏感信息",
	Long: `使用 gitleaks 规则引擎扫描已克隆仓库中的敏感信息（SSH 私钥、API Token、密码、AK/SK 等）。

默认扫描工作区文件。使用 --history 扫描 git 历史（包括已删除提交中的泄露）。
支持 --group 和 --resource 过滤扫描范围。`,
	Example: `  grepom scan                           # 扫描所有已克隆仓库的工作区
  grepom scan --group frontend          # 仅扫描 frontend 组
  grepom scan --resource work-gl        # 仅扫描 work-gl 资源下的仓库
  grepom scan web-app                   # 仅扫描 web-app 仓库
  grepom scan --history                 # 扫描工作区 + git 历史
  grepom scan --format json             # JSON 格式输出
  grepom scan --output results.txt      # 输出到文件
  grepom scan --gitleaks-config rules.toml  # 使用自定义规则`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVarP(&scanGroup, "group", "g", "", "按组名过滤")
	scanCmd.Flags().StringVarP(&scanResource, "resource", "R", "", "按资源名过滤")
	scanCmd.Flags().BoolVar(&scanHistory, "history", false, "扫描 git 历史（包括已删除提交中的泄露）")
	scanCmd.Flags().StringVarP(&scanFormat, "format", "f", "table", "输出格式: table, json")
	scanCmd.Flags().StringVar(&scanGitleaksCfg, "gitleaks-config", "", "自定义 gitleaks.toml 配置文件路径")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "将扫描结果输出到指定文件")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		// 配置文件不存在时回退到扫描当前目录
		if config.IsConfigNotFound(err) {
			return runScanCurrentDir()
		}
		return err
	}

	// 解析 repo 列表
	filter := repo.Filter{
		Group:    scanGroup,
		Resource: scanResource,
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

	// 过滤已克隆的仓库，收集扫描目标
	type scanTarget struct {
		name string
		path string
	}
	var targets []scanTarget
	var notCloned []string
	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		if git.IsCloned(fullPath) {
			targets = append(targets, scanTarget{name: r.Name, path: fullPath})
		} else {
			notCloned = append(notCloned, r.Name)
		}
	}

	// 提示未克隆的仓库
	for _, name := range notCloned {
		fmt.Fprintf(os.Stderr, "  skipping %s: not cloned\n", name)
	}

	if len(targets) == 0 {
		fmt.Println("No cloned repositories to scan.")
		return nil
	}

	// 创建 scanner
	s := scanner.NewScanner(scanner.Options{
		GitleaksConfigPath: scanGitleaksCfg,
	})

	// 并行扫描
	var (
		mu        sync.Mutex
		findings  []scanner.Finding
		completed int
		total     = len(targets)
	)

	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		go func(t scanTarget) {
			defer wg.Done()

			ctx := context.Background()
			var result []scanner.Finding
			var err error

			if scanHistory {
				result, err = s.ScanGitHistory(ctx, t.path)
			} else {
				result, err = s.ScanDir(ctx, t.path)
			}

			mu.Lock()
			completed++
			fmt.Fprintf(os.Stderr, "  Scanning... %d/%d repos\n", completed, total)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  error: %s: %v\n", t.name, err)
			} else {
				// 设置 repo 名称
				for i := range result {
					result[i].Repo = t.name
				}
				findings = append(findings, result...)
			}
			mu.Unlock()
		}(t)
	}
	wg.Wait()

	return outputFindings(findings)
}

// runScanCurrentDir 在无配置文件时扫描当前工作目录。
func runScanCurrentDir() error {
	fmt.Fprintln(os.Stderr, "  Scanning current directory (no config file found)...")

	s := scanner.NewScanner(scanner.Options{
		GitleaksConfigPath: scanGitleaksCfg,
	})

	ctx := context.Background()
	findings, err := s.ScanDir(ctx, ".")
	if err != nil {
		return fmt.Errorf("扫描当前目录失败: %w", err)
	}

	// 设置 repo 名称为当前目录
	dir, _ := os.Getwd()
	for i := range findings {
		findings[i].Repo = dir
	}

	return outputFindings(findings)
}

// outputFindings 根据格式标志将扫描结果输出到指定目标。
func outputFindings(findings []scanner.Finding) error {
	var w io.Writer = os.Stdout
	if scanOutput != "" {
		f, err := os.Create(scanOutput)
		if err != nil {
			return fmt.Errorf("无法创建输出文件 %q: %w", scanOutput, err)
		}
		defer f.Close()
		w = f
	}

	switch scanFormat {
	case "json":
		return outputJSON(w, findings)
	default:
		return outputTable(w, findings)
	}
}

func outputTable(w io.Writer, findings []scanner.Finding) error {
	if len(findings) == 0 {
		fmt.Fprintln(w, "No secrets found.")
		return nil
	}

	// 按 repo 分组
	grouped := make(map[string][]scanner.Finding)
	repoSet := make(map[string]bool)
	severityCount := make(map[scanner.Severity]int)

	for _, f := range findings {
		grouped[f.Repo] = append(grouped[f.Repo], f)
		repoSet[f.Repo] = true
		severityCount[f.Severity]++
	}

	// 排序 repo 名称以保证输出稳定
	repoNames := make([]string, 0, len(repoSet))
	for name := range repoSet {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	for _, repoName := range repoNames {
		fmt.Fprintf(w, "\n%s\n", repoName)
		for _, f := range grouped[repoName] {
			truncatedFile := scanner.TruncatePath(f.File, 40)
			masked := scanner.MaskSecret(f.Secret)
			fmt.Fprintf(w, "  %-40s  line %-4d  %-20s  %-8s  %s\n",
				truncatedFile, f.Line, f.RuleID, f.Severity, masked)
		}
	}

	// 汇总统计
	fmt.Fprintf(w, "\nFound %d findings in %d repos.\n", len(findings), len(repoSet))

	var parts []string
	for _, sev := range []scanner.Severity{
		scanner.SeverityCritical,
		scanner.SeverityHigh,
		scanner.SeverityMedium,
		scanner.SeverityLow,
	} {
		if c, ok := severityCount[sev]; ok && c > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c, sev))
		}
	}
	if len(parts) > 0 {
		fmt.Fprintf(w, "  %s\n", joinWithComma(parts))
	}

	return nil
}

func outputJSON(w io.Writer, findings []scanner.Finding) error {
	data, err := scanner.FindingsToJSON(findings)
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

// joinWithComma 将字符串切片用 ", " 连接。
func joinWithComma(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
