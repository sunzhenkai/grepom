## Context

grepom 是一个 Go CLI 工具，通过 `.grepom.yml` 管理多 git 仓库。当前 `list` 命令仅展示本地配置中已有的仓库信息，查看远程 provider 上的仓库需要先 `sync` 再 `list`，流程不直观。`status` 命令的头部摘要使用单行文本输出，repo 数量多时可读性差。

现有架构中，`provider.Provider` 接口已支持通过 API 查询远程仓库列表（`ListRepos` 方法），`sync` 命令已有完整的 provider 调用逻辑可复用。

## Goals / Non-Goals

**Goals:**
- `list` 命令支持 `--remote` 标志，实时查询远程 provider 仓库并表格展示
- `status` 命令头部摘要改为表格形式，提升扫读效率
- 复用现有 `provider.ListRepos` 接口，不引入新依赖

**Non-Goals:**
- 不支持远程仓库的 diff 对比（远程 vs 本地配置的差异）
- 不修改 `list --type resources` 和 `list --type groups` 的行为
- 不修改 `status` 命令的仓库列表部分（仅改头部摘要）
- 不引入新的 provider API 调用方法

## Decisions

### 1. `--remote` 标志仅作用于 `--type repos`（默认）

**选择**：`--remote` 作为 `list` 命令的顶级标志，仅在 `--type repos`（默认）时生效。

**理由**：`--type resources` 和 `--type groups` 展示的是本地配置元数据，与远程无关。若用户同时传入 `--remote --type resources`，报错提示不兼容。

**替代方案**：为 `--type` 新增 `remote-repos` 值。但 `--remote` 作为布尔标志更符合 CLI 惯例，且与现有 `--group`、`--resource` 过滤标志组合更自然。

### 2. 远程列表复用 sync 的 provider 调用逻辑

**选择**：提取 sync 中的 provider 调用模式（根据 resource 获取 provider 实例、构造 `ListReposParams`、调用 API），在 `runListRemoteRepos` 中复用相同模式。

**理由**：`provider.Provider.ListRepos` 接口已封装好分页、鉴权等逻辑，直接复用可避免重复代码。无需新增接口方法。

### 3. 远程列表输出与本地 list repos 保持一致的表格格式

**选择**：远程列表输出列：`NAME`、`PATH`、`GROUP`、`RESOURCE`、`CLONE_URL`。其中 `CLONE_URL` 替代本地模式的 `CLONED` 列，因为远程模式下仓库未克隆，展示 clone URL 更有价值。

**理由**：保持表格风格一致，同时提供远程模式下的有用信息。

### 4. status 表格格式

**选择**：使用 tabwriter 输出两列表格，左列为状态名称（STATUS），右列为数量（COUNT）。表格前不添加额外标题行。

**理由**：简洁直观，与项目其他表格输出风格一致（使用 `text/tabwriter`）。

## Risks / Trade-offs

- **[网络延迟]** `--remote` 需要调用远程 API，首次响应可能较慢 → 显示请求中的提示信息，或在 verbose 模式下输出 API 调用详情
- **[API 限流]** 远程 provider 可能有 rate limit → 复用现有 provider 的 rate limit 检测逻辑，遇到限流时输出友好错误信息
- **[大量仓库]** 某些 org/group 可能有数百个仓库 → 远程列表与本地列表一样全量输出，不做分页（与现有行为一致）
