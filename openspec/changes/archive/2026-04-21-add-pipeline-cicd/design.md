## Context

grepom 管理多个 GitLab/GitHub 仓库。用户在本地推送代码后，需要快速查看 CI/CD 状态。现有代码：

- `provider/` 包：GitLab、GitHub、Codeup 的仓库发现 API（ListRepos、ListGroups）
- `repo/resolver.go`：将 config 解析为 `provider.Repo` 列表，含已解析的 Token（repo > group > resource 优先级）
- `config/config.go`：Resource 定义含 ServerURL、Token、Provider 等信息
- `provider/github.go`：`githubAPIURL()` 处理 github.com → api.github.com 的 URL 转换
- `repo/resolver.go`：`extractRepoPath()` 从 CloneURL 提取远程仓库路径（未导出）

CI/CD 是全新领域，与仓库发现在概念上独立。需要新的 interface 和实现，但复用 token 解析和 config 结构。

## Goals / Non-Goals

**Goals:**
- 支持 `grepom pipeline list <repo>` 列出最近 N 条 pipeline
- 支持 `grepom pipeline watch <repo>` 实时监控最新 pipeline 状态
- 支持 GitLab Pipeline API 和 GitHub Actions API
- 复用现有 token 解析链路（repo → group → resource）
- watch 自然结束（pipeline 终态）或 Ctrl+C 提前终止

**Non-Goals:**
- 不支持 Codeup（云效）CI/CD API
- 不支持触发/重试/取消 pipeline（只读操作）
- 不支持 job 级别详情或实时日志流
- 不支持批量查看多个仓库的 pipeline

## Decisions

### 1. 独立 cicd 包，不扩展 Provider 接口

**选择**：新建 `cicd/` 包，定义 `PipelineProvider` 接口，与 `provider/` 包平级。

**理由**：CI/CD 和仓库发现在概念上是不同领域。现有 `Provider` 接口只有 `ListRepos` 和 `ListGroups`，强行扩展会违反接口隔离原则。独立的包和接口更清晰，也更容易测试。

### 2. 从 CloneURL 反推远程路径

**选择**：导出 `repo.ExtractRemotePath`，从 `provider.Repo.CloneURL` 提取远程仓库路径。

**理由**：
- 不修改现有 `provider.Repo` 结构体
- `extractRepoPath` 函数逻辑已经存在且测试过
- GitLab API 支持用 URL-encoded path 查 project，GitHub API 直接用 owner/repo

**替代方案**：给 Repo 加 RemotePath 字段 → 影响面更大，且 resolver 当前会在计算 local path 时覆盖 Path 字段。

### 3. Token 通过 Resolver 获取，ServerURL 通过 Resource 反查

**选择**：
- Token：通过 `repo.Resolver.ResolveAndFilter({Name: "web-app"})` 获取已解析的 `provider.Repo.Token`
- ServerURL + Provider 类型：通过 `cfg.Resources[repo.Resource]` 反查

**理由**：Resolver 已经处理了完整的 token 优先级链路（repo > group > resource），直接复用。ServerURL 需要从 Resource 获取（provider.Repo 不含此信息）。

### 4. List 默认 5 条，`-n` 参数调整

**选择**：`grepom pipeline list <repo>` 默认显示最近 5 条，`-n` / `--limit` 可调整，上限 20 条。

**理由**：5 条是大多数场景下合理的信息量，避免过多 API 调用和输出噪音。

### 5. Watch 轮询间隔 5 秒

**选择**：每 5 秒调一次 API 获取 pipeline 状态。

**理由**：5 秒是平衡实时性和 API 限流的合理间隔。GitLab 和 GitHub 的 rate limit 通常允许足够频率的轮询。不使用 SSE/WebSocket（各 provider 实现差异大，复杂度高）。

### 6. 终态定义

**选择**：`success`、`failed`、`canceled` 为终态，watch 检测到终态时自动退出。`skipped` 也视为终态。

### 7. Pipeline 数据模型

```go
type Pipeline struct {
    ID        int
    Status    PipelineStatus  // running, pending, success, failed, canceled
    Branch    string
    SHA       string          // 短 commit hash（前 7 位）
    StartedAt time.Time
    Duration  time.Duration   // 0 表示未开始
    URL       string          // Web URL
}
```

### 8. PipelineProvider 接口设计

```go
type PipelineProvider interface {
    ListPipelines(ctx context.Context, params ListPipelinesParams) ([]Pipeline, error)
    GetPipeline(ctx context.Context, params GetPipelineParams) (*Pipeline, error)
}

type ListPipelinesParams struct {
    ServerURL string
    Token     string
    RepoPath  string   // "org/fe/web-app"
    Limit     int
}

type GetPipelineParams struct {
    ServerURL  string
    Token      string
    RepoPath   string
    PipelineID int
}
```

使用与 `provider` 包相同的注册表模式（`Register` + `Get`）。

### 9. Provider API 映射

**GitLab：**
- 列表：`GET /projects/:id/pipelines?per_page=N&order_by=id&sort=desc`
  - `:id` = `url.PathEscape(repoPath)`
  - 认证：`PRIVATE-TOKEN` header
- 详情：`GET /projects/:id/pipelines/:pipeline_id`
- 状态映射：`running` → running, `pending` → pending, `success` → success, `failed` → failed, `canceled` → canceled

**GitHub：**
- 列表：`GET /repos/:owner/:repo/actions/runs?per_page=N`
  - URL 转换：`github.com` → `api.github.com`（复用 `githubAPIURL` 逻辑）
  - 认证：`Bearer` token + `Accept: application/vnd.github+json`
- 详情：`GET /repos/:owner/:repo/actions/runs/:run_id`
- 状态映射：`in_progress` → running, `queued` → pending, `completed+success` → success, `completed+failure` → failed, `completed+cancelled` → canceled

## Risks / Trade-offs

- **[API 限流风险]** → 5 秒轮询间隔 + 每次 watch 只查一个 pipeline，单用户场景下不太可能触发限流。如果触发，显示错误信息并退出。
- **[GitLab project ID 解析]** → 使用 URL-encoded path 可能有特殊字符问题。标准路径（字母数字/连字符/下划线/点）没有问题，极端情况可能报错——报错即可，不需要特殊处理。
- **[GitHub Actions vs Pipelines 语义差异]** → GitHub 用 Workflow Run 概念，统一映射为 Pipeline 可能丢失部分信息（如 workflow name）。初始版本可接受，后续可扩展。
- **[没有 RemoteID]** → 当前不从 API 获取 repository 的数字 ID，每次调用都通过 path 查询。多一次 API 调用（GitLab 的 project resolve），但简化了数据模型。
