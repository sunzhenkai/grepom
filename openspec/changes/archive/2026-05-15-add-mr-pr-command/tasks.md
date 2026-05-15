## 1. Git 辅助函数（git/ 包扩展）

- [x] 1.1 在 `git/git.go` 中实现 `GetCurrentBranch(path string) (string, error)` — 通过 `git rev-parse --abbrev-ref HEAD` 获取当前分支名
- [x] 1.2 在 `git/git.go` 中实现 `GetRemoteURL(path string, remote string) (string, error)` — 通过 `git remote get-url` 获取 remote URL
- [x] 1.3 在 `git/git.go` 中实现 `HasUnpushedCommits(path string, branch string) (bool, int, error)` — 通过 `git log origin/{branch}..HEAD --oneline` 检测未推送 commit
- [x] 1.4 在 `git/git.go` 中实现 `GetHeadCommitMessage(path string) (string, error)` — 通过 `git log -1 --format=%B` 获取 HEAD commit message
- [x] 1.5 为上述四个函数编写单元测试 `git/git_test.go`

## 2. MergeRequest Provider 接口层（mergerequest/ 包）

- [x] 2.1 创建 `mergerequest/mergerequest.go`，定义 `MergeRequest` 结构体、`CreateMergeRequestParams`、`WebURLParams`、`MergeRequestProvider` 接口（含 `CreateMergeRequest` 和 `BuildWebURL` 方法）、注册表（`Register`/`Get` 函数）
- [x] 2.2 创建 `mergerequest/github.go`，实现 `GitHubMRProvider`：`CreateMergeRequest` 调用 `POST /repos/{owner}/{repo}/pulls`；`BuildWebURL` 构建 `https://github.com/{owner}/{repo}/compare/{base}...{head}?expand=1`；处理 GHE API URL 差异；通过 `init()` 注册
- [x] 2.3 创建 `mergerequest/gitlab.go`，实现 `GitLabMRProvider`：`CreateMergeRequest` 调用 `POST /projects/{url-encoded-path}/merge_requests`，draft 时 title 前缀加 "Draft: "；`BuildWebURL` 构建 `{server}/{path}/-/merge_requests/new?merge_request[...]=...`；通过 `init()` 注册
- [x] 2.4 编写 `mergerequest/github_test.go` — 测试 PR 创建、Web URL 构建、GHE URL 处理、错误响应
- [x] 2.5 编写 `mergerequest/gitlab_test.go` — 测试 MR 创建（含 draft）、Web URL 构建、URL 编码、错误响应

## 3. Provider 识别与 Token 获取

- [x] 3.1 在 `cmd/mr.go` 中实现 `detectProvider(remoteURL string, cfg *config.Config) (providerName string, serverURL string, token string, error)` — 从 remote URL 解析 host，优先匹配 config resource，其次匹配知名域名，最后报错
- [x] 3.2 实现 `resolveToken(providerName string, host string, cfg *config.Config) (string, error)` — 优先从 config 获取 token，fallback 到环境变量 `GREPOM_GITHUB_TOKEN`/`GREPOM_GITLAB_TOKEN`，都没有则报错

## 4. MR/PR 命令实现（cmd/ 包）

- [x] 4.1 创建 `cmd/mr.go`，定义 cobra 命令：flags 包括 `--from`、`--to`、`--title`、`--body`、`--body-file`、`--draft`、`--web`；注册 `mr` 命令
- [x] 4.2 实现 `runMR` 命令处理函数主流程：检测当前目录是否 git 仓库 → 获取当前分支/默认分支 → 解析 from/to → 检测 provider → 获取 token
- [x] 4.3 实现未推送 commit 检测逻辑：调用 `HasUnpushedCommits`，有则检测 TTY，有 TTY 用 `survey.Confirm` 提示是否 push，无 TTY 直接报错；用户确认后执行 `git push`
- [x] 4.4 实现 title/body 提取逻辑：调用 `GetHeadCommitMessage`，解析第一行为 title、后续为 body；`--title`/`--body`/`--body-file` 可覆盖
- [x] 4.5 实现 `--web` 模式：调用 `BuildWebURL` 构建浏览器 URL，使用 `exec.Command("open"` / `xdg-open` / `cmd /c start` 打开浏览器，输出 URL 到 stdout
- [x] 4.6 实现 API 创建模式（非 --web）：调用 `CreateMergeRequest`，成功后输出 MR/PR 编号和 Web URL
- [x] 4.7 实现 Codeup 不支持处理：provider 为 codeup 时输出友好提示和浏览器 URL，不调用 API
- [x] 4.8 实现 `from == to` 保护：当源分支和目标分支相同时报错
- [x] 4.9 在 `cmd/root.go` 中注册 `mr` 命令；在 `mr.go` 的 `init()` 中额外注册 `pr` 别名命令（复用同一 `runMR` 函数）

## 5. 文档更新

- [x] 5.1 更新 `README.md`，在命令列表中新增 `mr`/`pr` 命令说明，包含使用示例
- [x] 5.2 更新 `README_en.md`，同步英文文档
