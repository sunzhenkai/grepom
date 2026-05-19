package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

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
	Provider  cicd.PipelineProvider
	ServerURL string
	RepoPath  string // 远程路径，如 "org/team/repo"
	Token     string
	RepoName  string // 用于显示的仓库名称
}

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
// 返回 PipelineProvider、ServerURL、远程路径和 Token。
func resolvePipelineInput(cfg *config.Config, repoName string) (cicd.PipelineProvider, string, string, string, error) {
	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(repo.Filter{Name: repoName})
	if err != nil {
		return nil, "", "", "", err
	}

	if len(repos) == 0 {
		return nil, "", "", "", fmt.Errorf("repo not found: %s", repoName)
	}

	r := repos[0]

	if r.Resource == "" {
		return nil, "", "", "", fmt.Errorf("repo %q has no resource binding, cannot query pipelines", repoName)
	}

	res, ok := cfg.Resources[r.Resource]
	if !ok {
		return nil, "", "", "", fmt.Errorf("resource %q not found", r.Resource)
	}

	remotePath := repo.ExtractRemotePath(r.CloneURL)
	if remotePath == "" {
		remotePath = repo.ExtractRemotePath(r.SSHURL)
	}
	if remotePath == "" {
		return nil, "", "", "", fmt.Errorf("cannot determine remote path for repo %q", repoName)
	}

	provider, err := cicd.Get(res.Provider)
	if err != nil {
		return nil, "", "", "", err
	}

	resolvedToken, err := res.ResolvedToken()
	if err != nil {
		return nil, "", "", "", fmt.Errorf("resource %q: %w", r.Resource, err)
	}

	return provider, res.APIURL(), remotePath, resolvedToken, nil
}

func runPipelineList(cmd *cobra.Command, args []string) error {
	_, cfg, err := loadConfig()
	if err != nil {
		return err
	}

	repoName := args[0]
	provider, serverURL, remotePath, token, err := resolvePipelineInput(cfg, repoName)
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
		ServerURL: serverURL,
		Token:     token,
		RepoPath:  remotePath,
		Limit:     limit,
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
	provider, serverURL, remotePath, token, err := resolvePipelineInput(cfg, repoName)
	if err != nil {
		return err
	}

	target := WatchTarget{
		Provider:  provider,
		ServerURL: serverURL,
		RepoPath:  remotePath,
		Token:     token,
		RepoName:  repoName,
	}

	return runWatchLoop(target, pipelineID, cmd)
}

// runWatchLoop 是 pipeline watch 和顶级 watch 命令共享的 watch 循环。
// target 包含 pipeline 查询所需的全部信息。
// targetID 为 0 时表示监控最新 pipeline，否则监控指定 ID。
func runWatchLoop(target WatchTarget, targetID int, cmd *cobra.Command) error {
	// 确定要 watch 的 pipeline ID
	if targetID == 0 {
		// 获取最新 pipeline
		pipelines, err := target.Provider.ListPipelines(cmd.Context(), cicd.ListPipelinesParams{
			ServerURL: target.ServerURL,
			Token:     target.Token,
			RepoPath:  target.RepoPath,
			Limit:     1,
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

	fmt.Printf("Watching pipeline #%d for %s... (Ctrl+C to stop)\n", targetID, target.RepoName)

	// 设置 signal handling：Ctrl+C 优雅退出
	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 立即查询一次
	pipeline, err := target.Provider.GetPipeline(ctx, cicd.GetPipelineParams{
		ServerURL:  target.ServerURL,
		Token:      target.Token,
		RepoPath:   target.RepoPath,
		PipelineID: targetID,
	})
	if err != nil {
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	// 打印 pipeline URL（开始时）
	printPipelineURL(pipeline)

	// watch 循环
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

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
			ServerURL:  target.ServerURL,
			Token:      target.Token,
			RepoPath:   target.RepoPath,
			PipelineID: targetID,
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
