## Context

grepom 是一个 Go CLI 工具，使用 cobra 框架，通过 YAML 配置文件管理 GitLab/GitHub 仓库。当前 `list` 命令仅支持列出 repos（从 groups 和独立 repos 中解析），不支持查看 resources 和 groups 的配置概览。用户需要手动查看 YAML 文件来了解 resource 和 group 的配置。

现有 `list` 命令模式：
- 接受可选位置参数 `[name]` 按 repo 名称过滤
- `--group` 和 `--resource` 标志过滤 repos
- 使用 `text/tabwriter` 输出对齐表格
- 数据流：`loadConfig()` → `repo.NewResolver(cfg)` → `resolver.ResolveAndFilter(filter)` → 表格输出

## Goals / Non-Goals

**Goals:**
- 用户能通过 CLI 列出所有已配置的 resources，查看 provider、url 等关键信息
- 用户能通过 CLI 列出所有已配置的 groups，查看名称、关联 resource、路径等信息
- 保持现有 `grepom list`（列出 repos）的行为完全向后兼容

**Non-Goals:**
- 不修改 resource 或 group 的数据模型
- 不新增外部依赖
- 不实现 resource/group 的删除或修改操作
- 不实现通过 API 从远端拉取 resource/group 信息（仅读取本地配置）

## Decisions

### 决策 1：使用 `--type` 标志扩展 list 命令

**选择**: 在 `list` 命令上新增 `--type` 标志，支持 `repos`（默认）、`resources`、`groups` 三种类型。

**替代方案**:
- **A. 为 list 添加子命令**（`list resources`、`list groups`）：更符合 cobra 惯例，但会导致位置参数 `[name]` 与子命令冲突——`grepom list web-app` 会被 cobra 解析为子命令匹配而非 repo 名称过滤，破坏向后兼容。
- **B. 新增顶层命令**（`grepom resources`、`grepom groups`）：语义清晰且不影响 list，但增加了顶层命令数量，且与 `add resource`/`add group` 的子命令模式不一致。

**理由**: `--type` 标志方案完全保持向后兼容（`grepom list`、`grepom list [name]`、`grepom list --group/--resource` 行为不变），同时自然扩展了 list 命令的语义范围。

### 决策 2：resources 表格输出格式

输出列：NAME、PROVIDER、URL、SSH_KEY

- NAME: resource 名称（map key）
- PROVIDER: gitlab 或 github
- URL: host 地址（含端口）
- SSH_KEY: 配置的 SSH key 路径，未配置显示 `-`

Token 不直接显示（安全考虑），仅在有 token 配置时标记。

### 决策 3：groups 表格输出格式

输出列：NAME、RESOURCE、PATH、LOCAL_PATH、RECURSIVE、REPOS

- NAME: group 名称
- RESOURCE: 关联的 resource 名称
- PATH: 远端路径
- LOCAL_PATH: 本地映射路径
- RECURSIVE: 是否递归（yes/no）
- REPOS: group 下 repo 数量

### 决策 4：--type 与现有标志的交互

当 `--type` 为 `resources` 或 `groups` 时，`--group`、`--resource` 标志和位置参数 `[name]` 不生效，输出忽略这些过滤条件。这避免了语义混淆（如 `--type resources --group xxx` 没有意义）。

## Risks / Trade-offs

- **[风险] --type 标志值冲突** → 使用 `repos`/`resources`/`groups` 三个明确值，未来如需扩展可增加枚举值。cobra 的 `ValidArgs` 可提供补全提示。
- **[权衡] Token 安全性** → resource 列表不显示 token 原始值，避免终端历史或输出重定向泄露敏感信息。如需调试 token，使用 `--verbose` 模式。
