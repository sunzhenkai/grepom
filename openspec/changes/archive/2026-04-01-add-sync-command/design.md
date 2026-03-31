## Context

grepom 通过配置文件中的 `sources` 条目管理 GitLab groups 和 GitHub orgs 的仓库。当前 `clone` 命令只 clone 未存在的仓库，`pull` 命令只更新已存在的仓库。当远程 group/org 新增了仓库或子 group 时，用户需要手动更新配置文件并分别执行 clone/pull。

现有架构：
- `provider.Provider` 接口：`ListRepos(ctx, source) → []Repo`，返回所有仓库
- `repo.Resolver`：聚合所有 source 的仓库，支持过滤
- `config.Config`：YAML 配置结构，包含 `Sources []Source`，每个 Source 有 `Groups []GroupSource` 和 `Orgs []OrgSource`
- GitLab provider 已实现 BFS 遍历子 group（当 `recursive: true`），但子 group 信息不会回写到配置
- 配置文件写入通过 `config.writeConfig` 完成，使用 `yaml.Marshal` 全量序列化

## Goals / Non-Goals

**Goals:**
- 提供一个 `sync` 命令，一步完成"发现新仓库 + clone + pull 已有仓库"
- 自动发现远程新增的子 group/org 并追加到配置文件
- 配置更新采用"只增不删"策略，保护用户手动修改的配置
- 防止并发 sync 同时修改配置文件的竞争问题

**Non-Goals:**
- 不自动删除远程已不存在的仓库或 group
- 不自动删除本地已 clone 但远程已删除的仓库
- 不修改 `clone` 和 `pull` 命令的现有行为
- 不引入配置文件版本控制或 diff 预览功能

## Decisions

### 1. 同步粒度：基于 source + group/org 筛选

**选择**: `grepom sync` 支持 `--source`（按 provider+url 匹配 source 索引）和 `--group`/`--org`（按名称匹配），默认同步配置中所有 source 的所有 group/org。

**理由**:
- 与现有 `clone --group` 的过滤模式一致
- 用户可能只想同步某个特定 group，不需要全量同步
- `--source` 用于区分同一 provider 不同 URL 的多个 source

### 2. 配置更新策略：只增不删

**选择**: 对比远程 API 返回的 group/org 列表与配置中已有的列表，只追加配置中不存在的新条目。

**理由**:
- 用户可能手动从配置中移除了某些 group（不想跟踪），自动补回会违背用户意图
- 只增不删是安全的默认行为，避免意外修改配置

### 3. 并发写入保护：进程级文件锁

**选择**: 使用 `os.OpenFile` + `syscall.Flock` 对配置文件加排他锁，在写入完成后释放。

**理由**:
- sync 可能被多个终端同时触发，并发写入会导致数据丢失
- flock 是 Unix 标准方案，Go 标准库直接支持
- 比 mutex 更适合跨进程场景

**替代方案**: 写入临时文件 + rename（原子写入）。但 rename 不能解决"读取-修改-写入"之间的 TOCTOU 问题，仍需加锁。

### 4. 配置更新的实现方式

**选择**: 在 `config` 包中新增 `SyncConfig` 函数，接受配置文件路径和需要追加的 group/org 列表，执行读取-对比-追加-写入的完整流程（在文件锁保护下）。

**理由**:
- 将配置更新逻辑封装在 config 包中，与现有 `AddSource`/`AddRepo` 风格一致
- 锁的获取和释放在同一个函数中完成，避免锁泄漏
- 命令层只需调用该函数，不需要关心并发细节

### 5. 子 group 发现：仅对 recursive=true 的 GitLab group 启用

**选择**: 只有当 GitLab group 配置了 `recursive: true` 时，sync 才会尝试发现并追加新的子 group。非递归 group 不发现子 group。

**理由**:
- 非递归 group 的用户明确不想跟踪子 group，不应自动追加
- 保持与 `ListRepos` 的行为一致
- GitHub org 没有 group 层级概念，不适用

### 6. Provider 接口扩展：新增 `ListSubGroups` 方法

**选择**: 在 `provider.Provider` 接口中新增可选方法（通过接口断言检测），而非修改现有接口。

**具体做法**: 定义 `SubGroupLister` 接口：
```go
type SubGroupLister interface {
    ListSubGroups(ctx context.Context, source config.Source, groupPath string) ([]string, error)
}
```
sync 命令通过类型断言 `if p, ok := provider.(SubGroupLister); ok` 来检测是否支持子 group 发现。

**理由**:
- 不破坏现有 `Provider` 接口，GitHub provider 不需要实现
- 接口隔离原则：只有需要子 group 发现的 provider 才实现该接口
- 避免在每个 provider 中实现空方法

## Risks / Trade-offs

- **[API 调用量增加]** → sync 需要额外的 API 调用来发现子 group。缓解：仅在 `recursive: true` 时才调用，且结果可缓存
- **[并发锁等待]** → flock 在另一个 sync 运行时会阻塞。缓解：设置合理的锁超时，超时后报错提示用户
- **[配置格式变化]** → `yaml.Marshal` 可能重新排序字段或改变格式。缓解：sync 只在确实有新 group/org 时才写文件
- **[只增不删可能导致配置膨胀]** → 如果用户大量使用 sync 而不手动清理。缓解：后续可考虑 `sync --prune` 选项
