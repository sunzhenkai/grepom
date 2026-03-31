## Context

grepom 是一个 CLI 工具，通过 YAML 配置文件管理多个 git 仓库。当前架构：

- **配置层**（`config/config.go`）：`Source` 结构体包含 provider/url/token/groups/orgs，通过数组索引引用；`Load()` 在加载时对整个文件内容调用 `os.ExpandEnv`，但 `writeConfig()` 直接将展开后的值写回，导致 token 占位符丢失。
- **命令层**（`cmd/sync.go`）：`sync` 命令既做远程 API 发现（获取仓库列表、发现子 group），又执行 clone/pull 操作，职责耦合。
- **Provider 层**（`provider/`）：GitLab/GitHub provider 通过 `ListRepos` 和 `ListSubGroups` 从远程 API 获取信息。

核心问题：
1. sync 职责过重——用户只想刷新配置元数据时被迫触发耗时的 clone/pull
2. source 仅靠数组索引引用——配置变更后索引可能错位，可读性差
3. token 占位符在回写时丢失——`writeConfig` 将展开后的明文写入磁盘

## Goals / Non-Goals

**Goals:**
- 将 sync 拆分为纯元数据同步（只读 API + 更新配置），clone/pull 交给已有命令
- Source 支持可选的 `name` 字段，可通过 `--source <name>` 引用
- Token 字段的 `${ENV_VAR}` 占位符在配置文件回写时保持原样

**Non-Goals:**
- 不改变配置文件的 YAML 格式主体（新增字段为可选，向后兼容）
- 不修改 provider 的 API 调用逻辑
- 不引入新的外部依赖

## Decisions

### 1. sync 命令移除 clone/pull 逻辑

**决策**：从 `cmd/sync.go` 中删除 clone 和 pull 相关代码，仅保留远程 API 发现 + 配置文件更新。

**替代方案**：
- (A) 新增一个 `discover` 子命令，保留 sync 的当前行为 → 增加认知负担，sync 名称更直觉上对应"同步元数据"
- (B) sync 增加 `--dry-run` flag 跳过 clone → 掩盖了核心问题，用户仍需了解 flag

**理由**：sync 的核心语义是"同步配置状态"，clone/pull 是独立的操作步骤。分离后用户可以 `sync` → `clone` → `pull` 逐步控制。

### 2. Source 的 name 字段设计

**决策**：在 `Source` 结构体中新增 `name string yaml:"name,omitempty"` 字段。`--source` 参数先尝试按 name 匹配，回退到按索引匹配。

**替代方案**：
- (A) 强制要求 name → 破坏现有配置文件的向后兼容
- (B) 使用 URL 作为隐式标识 → URL 可能变化，且不够简洁

**理由**：可选字段保证向后兼容。查找优先级 name > index 确保新配置可用名称引用，旧配置仍可索引访问。

### 3. Token 占位符保持策略

**决策**：配置文件使用 raw YAML 读写。加载时记录原始 token 字符串，回写时用原始值替换展开后的值。

具体实现：
- 新增 `RawSource` 结构体，用 `yaml:"token"` 保留原始 token 值
- `Load()` 时分别保存原始 YAML 和展开后的运行时值
- `writeConfig()` 写入时使用原始占位符值

**替代方案**：
- (A) 对 token 字段特殊处理正则替换 → 脆弱，难以维护
- (B) 不使用 `os.ExpandEnv`，改为仅对 token 字段做占位符解析 → 更安全，不会意外展开其他字段

**选择方案 B**：移除全局 `os.ExpandEnv`，改为仅对 `Source.Token` 字段做占位符解析。更精确、更安全。

### 4. 新增 Repos 写入配置

**决策**：sync 发现的仓库信息也写入配置文件的 `repos` 列表（如果尚未存在）。

**理由**：当前 sync 只追加 group/org，不保存发现的仓库列表。将发现的仓库信息持久化后，用户可以在不访问远程 API 的情况下执行 `clone`、`list` 等命令。

## Risks / Trade-offs

- **[行为变更] sync 不再自动 clone** → 用户需手动运行 `clone` 和 `pull`。在文档和输出中明确提示。
- **[数据丢失风险] 配置回写可能覆盖手动编辑** → 使用文件锁和只增不删策略已覆盖，保持现有机制。
- **[兼容性] name 字段可选** → 不填 name 时回退到索引，旧配置无需修改。
- **[安全性] 移除全局 ExpandEnv** → 部分用户可能依赖其他字段的环境变量展开。由于当前只有 token 使用占位符，影响有限。如有需要，可在后续迭代中为其他字段也添加占位符支持。
