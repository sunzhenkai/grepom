## Context

grepom 是一个 Git 仓库编排管理工具，通过 YAML 配置文件管理多个 GitLab/GitHub 的 resource、group 和 repo。当前所有配置条目始终处于活跃状态，没有禁用/排除机制。用户在管理大量 repo 时需要临时屏蔽某些条目，目前只能通过删除配置来实现。

当前数据模型：
- `Resource`：定义认证资源（provider、url、token、ssh_key）
- `Group`：引用 resource，包含 sync 发现的 repos 列表
- `Repo`：独立 repo，引用 resource
- `Resolver`（`repo/resolver.go`）：将配置解析为 `[]provider.Repo`，支持按 name/group/resource 过滤

## Goals / Non-Goals

**Goals:**
- 为 resource、group、repo 添加 `enabled` 开关，默认 `true`，设为 `false` 时排除该条目
- 为 group 添加 `exclude_repos` 列表，通过 repo 名称匹配排除特定 repo
- 在 resolver 解析阶段统一处理排除逻辑，所有下游命令自动生效
- sync 时保留 `exclude_repos` 配置不被覆盖
- 提供 `--all` 标志让用户在需要时包含被排除的条目

**Non-Goals:**
- 不实现通配符/glob 模式匹配（如 `exclude_repos: ["backend-*"]`），仅支持精确名称匹配
- 不实现条件启用（如基于时间、环境变量的启用逻辑）
- 不修改 interactive 模式的行为
- 不实现 repo 级别的 `exclude_repos`（独立 repo 直接使用 `enabled: false`）

## Decisions

### 1. 过滤层级与位置

**决定**：在 `Resolve()` 方法中添加排除逻辑，而非在各命令中分别处理。

**理由**：
- 当前 `Resolve()` 已经是所有命令的统一入口（clone、pull、status、list、search 都通过它获取 repo 列表）
- 集中过滤避免在每个命令中重复实现排除逻辑
- `Filter` 结构体新增 `IncludeDisabled` 字段控制是否包含被禁用的条目

**替代方案**：在各命令中分别过滤 → 代码重复，容易遗漏。

### 2. exclude_repos 匹配方式

**决定**：`exclude_repos` 使用 repo 的 `name` 字段进行精确匹配，列表格式为 `[]string`。

**理由**：
- GroupRepo 的 `name` 是用户可读的简短名称，最直观
- 精确匹配简单可靠，避免 glob 模式的复杂性和歧义
- YAML 数组格式与现有 `repos` 字段一致

**替代方案**：支持 glob 模式 → 增加复杂度，且大多数场景精确匹配已足够。

### 3. 排除逻辑的优先级

**决定**：`Resource.enabled` → `Group.enabled` → `Group.exclude_repos` → `Repo.enabled`，逐层过滤。

**具体规则**：
- Resource `enabled: false` → 该 resource 下所有 group 和独立 repo 被排除
- Group `enabled: false` → 该 group 下所有 repo 被排除
- repo name 在 group 的 `exclude_repos` 列表中 → 该 repo 被排除
- 独立 repo `enabled: false` → 该 repo 被排除

### 4. sync 保留 exclude_repos

**决定**：sync 命令在追加新 repo 时，保留 group 的 `exclude_repos` 列表不变。被排除的 repo 如果在远程仍存在，不会被重新添加到 repos 列表中（按 URL 去重已存在），但如果被删除又重新出现则会正常追加。

**理由**：`exclude_repos` 是用户的手动配置，sync 不应覆盖用户意图。与现有的"只增不删"策略一致。

### 5. 字段命名与默认值

**决定**：使用 `enabled` 布尔字段（默认 `true`），而非 `disabled` 或 `skip`。

**理由**：
- `enabled: true` 是正常状态，省略时表示启用，符合"默认启用"的直觉
- `disabled: false` 容易造成双重否定困惑
- `skip` 语义不够明确（跳过什么？何时跳过？）

## Risks / Trade-offs

- **[风险] 用户忘记启用的条目** → 在 `list` 命令输出中标注被禁用的条目状态（如 `[disabled]`），使用 `--all` 可查看所有条目
- **[风险] exclude_repos 与 enabled 的混淆** → 文档中明确区分：`enabled` 控制整体开关，`exclude_repos` 控制细粒度排除
- **[权衡] 不支持 glob 模式** → 如果 repo 数量很多且排除规则复杂，精确匹配需要逐个列出，但实现简单且不易出错。未来可以扩展支持 glob。
