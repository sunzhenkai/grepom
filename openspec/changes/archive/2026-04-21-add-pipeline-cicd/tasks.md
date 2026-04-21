## 1. 导出现有工具函数

- [x] 1.1 在 `repo/resolver.go` 中将 `extractRepoPath` 导出为 `ExtractRemotePath`，更新内部调用点

## 2. cicd 包核心接口

- [x] 2.1 创建 `cicd/cicd.go`，定义 `PipelineStatus` 类型及常量（running, pending, success, failed, canceled）、`Pipeline` 结构体、`PipelineProvider` 接口（ListPipelines, GetPipeline）、参数结构体（ListPipelinesParams, GetPipelineParams）、注册表（Register/Get/AvailableProviders）
- [x] 2.2 定义 `FormatStatus(status PipelineStatus) string` 和 `FormatDuration(d time.Duration) string` 辅助函数，用于终端输出格式化

## 3. GitLab Pipeline 实现

- [x] 3.1 创建 `cicd/gitlab.go`，实现 `GitLabPipelineProvider`，注册为 "gitlab"
- [x] 3.2 实现 `ListPipelines`：调用 `GET /projects/:id/pipelines?per_page=N&order_by=id&sort=desc`，`:id` 用 `url.PathEscape(repoPath)`，认证用 `PRIVATE-TOKEN` header，解析 JSON 响应映射为 `[]Pipeline`
- [x] 3.3 实现 `GetPipeline`：调用 `GET /projects/:id/pipelines/:pipeline_id`，解析单个 pipeline 详情
- [x] 3.4 实现 GitLab status 字符串到 `PipelineStatus` 的映射：running→running, pending→pending, success→success, failed→failed, canceled→canceled
- [x] 3.5 实现 HTTP client（复用现有 provider 的 `get`/`getWithPagination` 模式），含 30s timeout、错误处理

## 4. GitHub Actions 实现

- [x] 4.1 创建 `cicd/github.go`，实现 `GitHubPipelineProvider`，注册为 "github"
- [x] 4.2 实现 `ListPipelines`：调用 `GET /repos/:owner/:repo/actions/runs?per_page=N`，URL 转换 github.com→api.github.com，认证用 `Bearer` token + `Accept` header，解析 `workflow_runs` 数组映射为 `[]Pipeline`
- [x] 4.3 实现 `GetPipeline`：调用 `GET /repos/:owner/:repo/actions/runs/:run_id`
- [x] 4.4 实现 GitHub status+conclusion 到 `PipelineStatus` 的映射：in_progress→running, queued→pending, completed+success→success, completed+failure→failed, completed+cancelled→canceled
- [x] 4.5 实现 HTTP client，含 rate limit 检测和错误处理

## 5. pipeline list 命令

- [x] 5.1 创建 `cmd/pipeline.go`，定义 `pipelineCmd`（子命令容器）、`pipelineListCmd`（`pipeline list <repo>`），注册到 rootCmd
- [x] 5.2 定义 flags：`-n` / `--limit`（默认 5，上限 20）
- [x] 5.3 实现 `runPipelineList`：加载 config → `resolver.ResolveAndFilter({Name: repoName})` → 校验 repo 存在且有 Resource → 从 CloneURL 反推远程路径 → 反查 Resource 获取 ServerURL/Provider → `cicd.Get(provider).ListPipelines(...)` → 格式化输出表格
- [x] 5.4 实现输出格式化：tabwriter 表格（ID, BRANCH, SHA, STATUS, DURATION），状态图标映射（🔄⏳✅❌🚫），duration 人可读格式

## 6. pipeline watch 命令

- [x] 6.1 在 `cmd/pipeline.go` 中定义 `pipelineWatchCmd`（`pipeline watch <repo>`），flags：`--id`（可选，指定 pipeline ID）
- [x] 6.2 实现 `runPipelineWatch`：加载 config → 同上解析流程 → 如未指定 `--id` 则先调 `ListPipelines(limit=1)` 获取最新 pipeline ID → 进入 watch 循环
- [x] 6.3 实现 watch 循环：`context.WithCancel` + `signal.NotifyContext(SIGINT)` → 循环内调 `GetPipeline` → 用 `\r` 覆盖当前行输出状态 → 检测终态退出 → 5s sleep（可被 ctx cancel 打断）
- [x] 6.4 实现终态退出输出：检测到终态后输出最终状态行（含 duration 和状态图标），换行后退出

## 7. 测试

- [x] 7.1 为 `cicd/cicd.go` 编写测试：注册表 Register/Get、status 格式化、duration 格式化
- [x] 7.2 为 `cicd/gitlab.go` 编写测试：用 httptest mock GitLab API，测试 ListPipelines、GetPipeline、status 映射、分页、错误处理
- [x] 7.3 为 `cicd/github.go` 编写测试：用 httptest mock GitHub API，测试 ListPipelines、GetPipeline、status 映射、错误处理
- [x] 7.4 为 `repo/resolver.go` 的 `ExtractRemotePath` 编写测试：HTTPS URL、SSH URL、带端口、纯路径
