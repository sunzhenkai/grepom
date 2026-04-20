## Context

grepom 管理多个 group 下的 repos，通过 `exclude_repos` 字段排除不需要的 repo。当前存在的问题：

1. 被 exclude 的 repo 如果之前已经克隆到磁盘，不会被自动清理，残留目录造成空间浪费和混淆。
2. 多个 group 之间可能存在同名 repo（不同 team 的仓库，name 相同但 remote path 不同），用户需要按优先级去重，但当前只能手动编辑 `exclude_repos`。

现有相关代码：
- `repo/resolver.go`：`IsExcluded()` 匹配 exclude 规则，`resolveInternal()` 为 excluded repo 设置 `DisabledReason="excluded"`
- `git/git.go`：`IsCloned()`、`GetStatus()` 可用于检查磁盘状态
- `config/config.go`：`SyncGroupRepos()`、`WithFileLock()` 等方法用于安全写 config

## Goals / Non-Goals

**Goals:**
- 提供 `grepom prune` 命令清理已克隆但被 exclude 的 repos
- 提供 `grepom dedup` 命令自动检测跨 group 同名 repo 并排除
- prune 默认 dry-run，安全优先（dirty/ahead 跳过）
- dedup 只修改用户指定的 target group，reference group 不受影响

**Non-Goals:**
- 不自动触发 prune 或 dedup（均为用户主动执行）
- 不实现跨 group 的 repo 合并或迁移
- 不处理 standalone repos（`cfg.Repos`）的去重，仅处理 group repos

## Decisions

### 1. prune 默认 dry-run，`--apply` 执行删除

**选择**：默认只输出计划不删除，`--apply` flag 才真正执行。

**理由**：删除磁盘目录是破坏性操作。默认 dry-run 让用户先预览，避免误删。这与 `--force` 不同：`--force` 跳过 dirty/ahead 安全检查，`--apply` 控制是否执行。

**替代方案**：默认执行 + `--dry-run` 预览 → 风险太高，用户可能误操作。

### 2. prune 安全检查策略

**选择**：
- `dirty`（有未提交改动）→ 跳过
- `ahead`（本地领先远程）→ 跳过
- `clean` → 可删除
- `--force` → 跳过所有安全检查

**理由**：dirty 和 ahead 意味着本地有未保存的工作，删除会造成不可逆的数据丢失。

### 3. dedup 的 `--group` + `--reference` 模型

**选择**：`--group` 指定 target（必需），`--reference` 指定参考 group(s)（可选，逗号分隔，不指定时对比所有其他 group）。只修改 target group。

**理由**：
- 明确的 target/reference 语义让用户精确控制哪些 group 受影响
- `--reference` 可选，不指定时对比所有其他 group 是最常见用法
- 只改 target 符合最小修改原则

**替代方案**：全局自动 dedup（所有 group 互相比较）→ 用户无法控制哪些 group 被修改，风险不可控。

### 4. dedup 匹配粒度：repo name 精确匹配

**选择**：按 `repo.Name` 精确匹配，不考虑 path。

**理由**：用户关心的是 "同名 repo"，如 `api-lib` 在多个 team 都有。path 不同但 name 相同时，用户通常只保留优先级高的那个。

### 5. dedup 写入策略：移除 repos 条目 + 追加 exclude_repos

**选择**：从 target group 的 `repos` 列表移除该 repo，同时追加到 `exclude_repos`。

**理由**：
- 从 `repos` 移除：config 更干净，`list` 不会显示已排除的 repo
- 追加 `exclude_repos`：防止下次 `sync` 重新发现该 repo
- 两者配合确保一次 dedup 持久生效

### 6. dedup 的 exclude_repos 去重

**选择**：追加前检查是否已存在（精确匹配或 glob 覆盖），不重复添加。

**实现**：用 `repo.IsExcluded()` 检查现有 `exclude_repos` 是否已覆盖该 repo name。

## Risks / Trade-offs

- **[prune 误删风险]** → 默认 dry-run + 安全检查 + `--apply` 二次确认，三重保护
- **[dedup 误排除风险]** → `--dry-run` 预览，用户可先检查再执行
- **[dedup 后 sync 重新发现问题]** → `exclude_repos` 确保 sync 时 `IsExcluded()` 拦截，不会重新加入
- **[dedup 多 reference 合并]** → 多个 reference group 的 repo names 取并集，可能排除比预期更多的 repos → dry-run 让用户确认
- **[并发安全]** → 写入 config 使用现有的 `WithFileLock` 机制
