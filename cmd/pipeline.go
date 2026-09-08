package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"context"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/cicd"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/repo"
)

var (
	pipelineLimit int
	pipelineID    int
)

// WatchTarget 封装了 pipeline watch 所需的全部信息。
// 由 resolvePipelineInput 或 resolveCurrentRepoPipeline 构造，
// 供 runWatchLoop 使用。
type WatchTarget struct {
	Provider       cicd.PipelineProvider
	ServerURL      string
	RepoPath       string // 远程路径，如 "org/team/repo"
	Token          string
	RepoName       string // 用于显示的仓库名称
	OrganizationID string // 仅 Codeup provider 使用（云效 Flow 查询需要）
	WatchTag       string // 可选：tag -w 场景下的新 tag 名（用于提示信息）
	WatchSHA       string // 可选：tag -w 场景下绑定的 commit SHA；非空时先等待该 SHA 的 pipeline 出现
}

// SHA 等待阶段的参数（var 便于测试注入更小的时间窗口）。
var (
	watchSHAWaitTimeout = 60 * time.Second
	watchSHAPollEvery   = 2 * time.Second
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "View CI/CD pipelines",
	Long:  `View and monitor CI/CD pipelines for repositories. Supports GitLab and GitHub.`,
}

var pipelineListCmd = &cobra.Command{
	Use:   "list <repo-name>",
	Short: "List recent pipelines",
	Long:  `List recent CI/CD pipeline runs for a specific repository.`,
	Example: `  grepom pipeline list web-app           # Last 5 pipelines
  grepom pipeline list web-app -n 10     # Last 10 pipelines`,
	Args: cobra.ExactArgs(1),
	RunE: runPipelineList,
}

var pipelineWatchCmd = &cobra.Command{
	Use:   "watch <repo-name>",
	Short: "Watch the latest pipeline",
	Long: `Watch the latest CI/CD pipeline for a repository.
Polls every 5 seconds until the pipeline reaches a terminal state (success, failed, canceled).
Press Ctrl+C to stop early.`,
	Example: `  grepom pipeline watch web-app           # Watch latest pipeline
  grepom pipeline watch web-app --id 1234 # Watch specific pipeline`,
	Args: cobra.ExactArgs(1),
	RunE: runPipelineWatch,
}

func init() {
	pipelineListCmd.Flags().IntVarP(&pipelineLimit, "limit", "n", 5, "number of pipelines to show (max 20)")
	pipelineWatchCmd.Flags().IntVar(&pipelineID, "id", 0, "specific pipeline ID to watch (default: latest)")

	pipelineCmd.AddCommand(pipelineListCmd)
	pipelineCmd.AddCommand(pipelineWatchCmd)
	rootCmd.AddCommand(pipelineCmd)
}

// resolvePipelineInput 是 list 和 watch 共用的 repo 解析逻辑。
// 返回 PipelineProvider、ServerURL、远程路径、Token 和 OrganizationID（仅 Codeup 使用）。
func resolvePipelineInput(cfg *config.Config, repoName string) (cicd.PipelineProvider, string, string, string, string, error) {
	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(repo.Filter{Name: repoName})
	if err != nil {
		return nil, "", "", "", "", err
	}

	if len(repos) == 0 {
		return nil, "", "", "", "", fmt.Errorf("repo not found: %s", repoName)
	}

	r := repos[0]

	if r.Resource == "" {
		return nil, "", "", "", "", fmt.Errorf("repo %q has no resource binding, cannot query pipelines", repoName)
	}

	res, ok := cfg.Resources[r.Resource]
	if !ok {
		return nil, "", "", "", "", fmt.Errorf("resource %q not found", r.Resource)
	}

	remotePath := repo.ExtractRemotePath(r.CloneURL)
	if remotePath == "" {
		remotePath = repo.ExtractRemotePath(r.SSHURL)
	}
	if remotePath == "" {
		return nil, "", "", "", "", fmt.Errorf("cannot determine remote path for repo %q", repoName)
	}

	provider, err := cicd.Get(res.Provider)
	if err != nil {
		return nil, "", "", "", "", err
	}

	resolvedToken, err := res.ResolvedToken()
	if err != nil {
		return nil, "", "", "", "", fmt.Errorf("resource %q: %w", r.Resource, err)
	}

	return provider, res.APIURL(), remotePath, resolvedToken, res.OrganizationID, nil
}

func runPipelineList(cmd *cobra.Command, args []string) error {
	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	repoName := args[0]
	provider, serverURL, remotePath, token, organizationID, err := resolvePipelineInput(cfg, repoName)
	if err != nil {
		return err
	}

	limit := pipelineLimit
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	pipelines, err := provider.ListPipelines(cmd.Context(), cicd.ListPipelinesParams{
		ServerURL:      serverURL,
		Token:          token,
		RepoPath:       remotePath,
		Limit:          limit,
		OrganizationID: organizationID,
	})
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		fmt.Printf("No pipelines found for %s.\n", repoName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tBRANCH\tSHA\tSTATUS\tDURATION")

	for _, p := range pipelines {
		fmt.Fprintf(w, "#%d\t%s\t%s\t%s\t%s\n",
			p.ID, p.Branch, p.SHA, cicd.FormatStatus(p.Status), cicd.FormatDuration(p.Duration))
	}
	w.Flush()

	return nil
}

func runPipelineWatch(cmd *cobra.Command, args []string) error {
	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	repoName := args[0]
	provider, serverURL, remotePath, token, organizationID, err := resolvePipelineInput(cfg, repoName)
	if err != nil {
		return err
	}

	target := WatchTarget{
		Provider:       provider,
		ServerURL:      serverURL,
		RepoPath:       remotePath,
		Token:          token,
		RepoName:       repoName,
		OrganizationID: organizationID,
	}

	return runWatchLoop(target, pipelineID, cmd)
}

// runWatchLoop 是 pipeline watch 和顶级 watch 命令共享的 watch 循环。
// target 包含 pipeline 查询所需的全部信息。
// targetID 为 0 时表示监控最新 pipeline，否则监控指定 ID。
// target.WatchSHA 非空（tag -w 场景）时，先按 SHA 轮询等待目标 pipeline
// 出现，绝不退而监控其他 pipeline（避免 provider 异步创建期间的竞态）。
func runWatchLoop(target WatchTarget, targetID int, cmd *cobra.Command) error {
	// 设置 signal handling：Ctrl+C 优雅退出（覆盖等待阶段与 watch 循环）
	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 确定要 watch 的 pipeline ID
	if targetID == 0 {
		// tag -w：SHA 绑定，等待新 tag 的 pipeline 出现
		if target.WatchSHA != "" {
			id, err := waitForPipelineBySHA(ctx, target)
			if err != nil {
				return err
			}
			if id == 0 {
				// 等待阶段被 Ctrl+C 中断，已输出提示，优雅退出
				return nil
			}
			targetID = id
		} else {
			// 获取最新 pipeline
			pipelines, err := target.Provider.ListPipelines(ctx, cicd.ListPipelinesParams{
				ServerURL:      target.ServerURL,
				Token:          target.Token,
				RepoPath:       target.RepoPath,
				Limit:          1,
				OrganizationID: target.OrganizationID,
			})
			if err != nil {
				return fmt.Errorf("failed to find latest pipeline: %w", err)
			}
			if len(pipelines) == 0 {
				fmt.Printf("No pipelines found for %s.\n", target.RepoName)
				return nil
			}
			targetID = pipelines[0].ID
		}
	}

	fmt.Printf("Watching pipeline #%d for %s... (Ctrl+C to stop)\n", targetID, target.RepoName)

	// 立即查询一次
	pipeline, err := target.Provider.GetPipeline(ctx, cicd.GetPipelineParams{
		ServerURL:      target.ServerURL,
		Token:          target.Token,
		RepoPath:       target.RepoPath,
		PipelineID:     targetID,
		OrganizationID: target.OrganizationID,
	})
	if err != nil {
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	return runWatchPollLoop(target, targetID, ctx, pipeline)
}

// waitForPipelineBySHA 按 SHA 轮询等待目标 pipeline 出现。
// 命中返回其 ID；等待阶段被中断（Ctrl+C）返回 (0, nil)；超时返回错误。
func waitForPipelineBySHA(ctx context.Context, target WatchTarget) (int, error) {
	sha := strings.ToLower(target.WatchSHA)
	shaShort := sha
	if len(shaShort) > 7 {
		shaShort = shaShort[:7]
	}
	if target.WatchTag != "" {
		fmt.Printf("Waiting for pipeline of %s (%s)...\n", target.WatchTag, shaShort)
	}

	deadline := time.Now().Add(watchSHAWaitTimeout)
	for {
		pipelines, err := target.Provider.ListPipelines(ctx, cicd.ListPipelinesParams{
			ServerURL:      target.ServerURL,
			Token:          target.Token,
			RepoPath:       target.RepoPath,
			Limit:          5,
			SHA:            target.WatchSHA,
			OrganizationID: target.OrganizationID,
		})
		if err == nil {
			// 双保险：provider 过滤之外再做本地前缀比对（兼容不支持
			// SHA 过滤的 provider，如 Codeup）。
			for _, p := range pipelines {
				if p.SHA != "" && strings.HasPrefix(sha, strings.ToLower(p.SHA)) {
					return p.ID, nil
				}
			}
		}
		// 查询失败不立即放弃：在窗口内重试（瞬时网络/限流）。

		if time.Now().After(deadline) {
			tag := target.WatchTag
			if tag == "" {
				tag = target.RepoName
			}
			return 0, fmt.Errorf(
				"timeout after %s: no pipeline found for tag %s (commit %s)\n\nPossible causes:\n  • the tag was created locally but not pushed yet (use -p or push it manually)\n  • no workflow/pipeline listens to tag push events for this repository",
				watchSHAWaitTimeout, tag, shaShort)
		}

		select {
		case <-ctx.Done():
			fmt.Println()
			fmt.Printf("Watch stopped while waiting for pipeline of %s (%s).\n", target.WatchTag, shaShort)
			return 0, nil
		case <-time.After(watchSHAPollEvery):
		}
	}
}

// runWatchPollLoop 渲染状态行并按 5s 间隔轮询直到终态/取消。
func runWatchPollLoop(target WatchTarget, targetID int, ctx context.Context, pipeline *cicd.Pipeline) error {
	// 打印 pipeline URL（开始时）
	printPipelineURL(pipeline)

	// watch 循环
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var err error
	for {
		// 渲染状态行
		fmt.Printf("\r  %s  #%d  %s  %s  (%s)",
			cicd.FormatStatus(pipeline.Status),
			pipeline.ID,
			pipeline.Branch,
			pipeline.SHA,
			formatWatchDuration(pipeline),
		)

		// 检查终态
		if pipeline.Status.IsTerminal() {
			fmt.Println()
			fmt.Printf("Pipeline finished: %s (%s)\n",
				cicd.FormatStatus(pipeline.Status),
				cicd.FormatDuration(pipeline.Duration))
			// 打印 pipeline URL（终态退出时）
			printPipelineURL(pipeline)
			return nil
		}

		// 等待下一次轮询或 ctx 取消
		select {
		case <-ctx.Done():
			fmt.Println()
			fmt.Printf("Watch stopped. Current status: %s\n", cicd.FormatStatus(pipeline.Status))
			// 打印 pipeline URL（Ctrl+C 退出时）
			printPipelineURL(pipeline)
			return nil
		case <-ticker.C:
		}

		// 轮询
		pipeline, err = target.Provider.GetPipeline(ctx, cicd.GetPipelineParams{
			ServerURL:      target.ServerURL,
			Token:          target.Token,
			RepoPath:       target.RepoPath,
			PipelineID:     targetID,
			OrganizationID: target.OrganizationID,
		})
		if err != nil {
			fmt.Println()
			return fmt.Errorf("polling error: %w", err)
		}
	}
}

// printPipelineURL 打印 pipeline 的 Web URL。
// URL 为空时静默跳过。
func printPipelineURL(p *cicd.Pipeline) {
	if p.URL != "" {
		fmt.Printf("  👉 %s\n", p.URL)
	}
}

// formatWatchDuration 返回 watch 状态行用的持续时间文本。
// 对于正在运行的 pipeline，使用 wall clock 计算 elapsed 时间。
func formatWatchDuration(p *cicd.Pipeline) string {
	if p.Duration > 0 {
		return cicd.FormatDuration(p.Duration)
	}
	if !p.StartedAt.IsZero() && !p.Status.IsTerminal() {
		elapsed := time.Since(p.StartedAt)
		if elapsed > 0 {
			return cicd.FormatDuration(elapsed)
		}
	}
	return "-"
}
