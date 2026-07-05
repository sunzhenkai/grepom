## Context

grepom 通过 provider 抽象从远端（Codeup/GitLab/GitHub）发现代码库。Codeup（云效）对"删除"采用回收站机制：被删除的代码库/代码组并不立即消失，而是被重命名为 `<原名>-deletion_scheduled-<id>` 并保留一段宽限期。这些库已不可克隆（鉴权被收回），但 Codeup 的 `ListRepositories`/`ListGroupRepositories` 接口仍会返回它们。

当前 `provider/codeup.go` 的 `ListRepos` 把这些库原样映射为 `provider.Repo`，导致：
1. `grepom sync` 把删除中的库写入 `.grepom.yml`
2. `grepom clone` 尝试克隆并最终抛出 `all authentication methods failed`（来自 `git/git.go:Clone`），错误信息误导用户去排查鉴权配置

相关现状：
- `codeupRepo` 结构体已有 `Archived bool` 字段但未被使用
- `provider.Repo.DisabledReason` 已支持 `""`/`"disabled"`/`"excluded"`，由 `repo/resolver.go` 统一在 `Resolve()`/`ResolveAndFilter()` 中过滤
- `ListReposParams` 当前只有 `ServerURL/Token/Groups/Orgs/OrganizationID`，无控制开关

## Goals / Non-Goals

**Goals:**

- 发现阶段（Codeup `ListRepos`）默认剔除 `deletion_scheduled` 代码库，使其不进入配置
- 运行时兜底：对已写入配置但处于删除中的库，在 resolver 层标记并跳过，输出清晰原因而非鉴权错误
- 提供 `--include-deleted` 显式开关，允许用户查看/操作这些库
- verbose 模式下提示被跳过的删除中库数量
- 覆盖单元测试与中英文 README

**Non-Goals:**

- 不处理 `archived`（归档）代码库——`codeupRepo.Archived` 字段已存在但属另一关注点，留待后续独立 change
- 不变更 GitLab/GitHub provider 的行为（这两个 provider 的归档/删除语义不同，且未观察到等价问题）
- 不解析 Codeup API 可能存在的 `status` 字段——仅依赖可靠的路径命名标记
- 不自动清理配置中已存在的删除中库条目（仅运行时跳过；如需清理用户可手动 `dedup`/编辑）

## Decisions

### 1. 检测方式：基于 name/path 子串 `deletion_scheduled`

**选择**：新增辅助函数 `isDeletionScheduled(name, pathWithNamespace string) bool`，当 `name` 或 `pathWithNamespace` 包含子串 `deletion_scheduled` 时判定为删除中。

**替代方案**：解析 Codeup API 返回的 `status` 字段（如 `deletion_scheduled`）。

**理由**：路径命名标记在实测错误路径中稳定可见（如 `dsp-services-deletion_scheduled-452/creative-matching-deletion_scheduled-499`），且不依赖 API 字段是否存在/字段名是否稳定。子串匹配足够安全——正常业务库不会包含该词。同时保留对 group 段命名的兼容（整个组被删时其下所有库路径都带标记）。

### 2. 过滤位置：provider 发现层为主，resolver 为兜底

**选择**：双层防护。

| 层 | 位置 | 职责 |
|----|------|------|
| 发现层 | `provider/codeup.go` `ListRepos` | 默认不返回删除中的库（防进入配置） |
| 兜底层 | `repo/resolver.go` `resolveInternal` | 对已入库配置的删除中库，置 `DisabledReason="deletion_scheduled"` |

**替代方案**：仅在 resolver 层处理（更集中，但 sync 仍会把废弃库写进配置）。

**理由**：发现层拦截能从源头阻止污染配置；兜底层保证存量配置与手工添加的删除中库也能优雅跳过。两者共用同一检测函数，行为一致。

### 3. 显式开关：`ListReposParams.IncludeDeleted` + CLI `--include-deleted`

**选择**：
- `provider.ListReposParams` 新增 `IncludeDeleted bool`
- Codeup `ListRepos` 在 `IncludeDeleted=false`（默认）时跳过删除中库
- `grepom sync`、`grepom list` 新增 `--include-deleted` 布尔标志，透传该字段
- resolver 的兜底检测同样尊重 `Filter.IncludeDisabled`（复用现有 `--all` 机制，`--all` 已包含所有 disabled/excluded 条目）

**理由**：默认安全 + 可逃逸。复用 `--all`（`IncludeDisabled`）展示机制，无需为展示删除中库新增第二条路径。

### 4. 复用 `DisabledReason` 机制

**选择**：`DisabledReason` 新增取值 `"deletion_scheduled"`。`Resolve()` 默认剔除（与 `disabled`/`excluded` 同等对待）；`ResolveAndFilter` 在 `IncludeDisabled=true` 时保留并标注。

**替代方案**：新增独立字段 `DeletionScheduled bool`。

**理由**：`DisabledReason` 已是统一的排除语义出口，所有命令（clone/pull/status/list）都已通过它过滤，零改动即覆盖全链路。

### 5. 跳过提示

**选择**：在 sync 命令的 verbose 输出中，汇总本次发现阶段跳过的删除中库数量（如 `skipped 3 deletion_scheduled repos`）。list/clone 不额外打印，避免噪音。

## Risks / Trade-offs

- **[误判风险]** 某个正常库恰好命名含 `deletion_scheduled` → 极低概率；且用户可用 `--include-deleted` 临时恢复
- **[Codeup 命名变更]** 若 Codeup 未来改变回收站命名约定 → 检测函数为单一可测试点，易于调整；且兜底层使克隆失败时不再报误导性鉴权错误
- **[存量配置]** 已写入配置的删除中库不会被自动移除，仅运行时跳过 → 用户可手动 `grepom dedup` 或编辑；本 change 不引入破坏性配置迁移
- **[仅覆盖 Codeup]** GitLab/GitHub 未处理 → 这两个 provider 未观察到等价问题，符合最小变更原则
- **[与 archived 的关系]** `codeupRepo.Archived` 仍未使用 → 留待后续 change，避免范围蔓延

## Open Questions

（无 — 检测标记来自实测错误路径，决策已收敛；`--include-deleted` 语义与现有 `--all` 正交，不冲突）
