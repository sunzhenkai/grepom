## Context

当前 `exclude_repos` 在 `repo/resolver.go` 的 `isExcluded()` 函数中实现，仅作为运行时过滤器，覆盖通过 Resolver 的命令（clone、pull、status、list 本地、search）。但 `sync` 和 `list --remote` 直接遍历 config groups 并调用 provider API，不经过 Resolver，导致被排除的仓库仍出现在同步结果和远程列表中。

**当前实现关键路径：**
- `cmd/sync.go:122-128`：发现远程仓库后直接转换为 `GroupRepo` 写入配置，无 `exclude_repos` 过滤
- `cmd/list.go:243-255`（`runListRemoteRepos`）：查询远程仓库后直接追加到展示列表，无 `exclude_repos` 过滤
- `repo/resolver.go:132-140`（`isExcluded`）：当前为包内私有函数，sync/list 无法复用

## Goals / Non-Goals

**Goals:**
- `sync` 发现仓库时跳过 `exclude_repos` 中匹配的仓库，不写入配置
- `list --remote` 默认过滤被排除的仓库，与 `list`（本地）行为一致
- `list --remote --all` 可查看包含被排除仓库的完整远程列表
- 导出 `isExcluded` 函数供 sync 和 list 复用，避免逻辑重复

**Non-Goals:**
- 不修改 `exclude_repos` 的匹配逻辑（仍为精确名称匹配）
- 不为 Resource 或独立 Repo 增加 `exclude_repos` 字段
- 不修改 `DetectPathConflicts()` 中的排除逻辑
- 不改变 `exclude_repos` 的配置格式

## Decisions

### 1. 导出 `isExcluded` 为包级公共函数

**选择**：将 `repo/resolver.go` 中的 `isExcluded()` 导出为 `IsExcluded()`，供 `cmd/sync.go` 和 `cmd/list.go` 调用。

**备选方案**：
- A) 在 `config` 包中新增 `IsExcluded` 方法 → 职责不清，config 包应只管数据
- B) 在 sync/list 中各自内联实现 → 逻辑重复，违反 DRY
- C) 新建 `filter` 包 → 过度设计，仅一个简单函数

**理由**：`isExcluded` 本质是 repo 过滤逻辑，属于 `repo` 包的职责。导出是最小改动，且 Resolver 已是 `repo` 包的一部分。

### 2. sync 过滤时机：写入前过滤

**选择**：在 `sync.go` 中，将远程发现的 repos 转为 `GroupRepo` 列表后、调用 `SyncGroupRepos` 写入前，过滤掉匹配 `exclude_repos` 的条目。

**备选方案**：
- A) 在 provider API 层过滤 → provider 不应了解业务逻辑
- B) 在 `SyncGroupRepos` 内部过滤 → 函数职责变复杂，且配置函数不应包含运行时过滤逻辑

**理由**：sync 命令是唯一调用者，在调用侧过滤最清晰。同时需要在 verbose 模式下输出被跳过的仓库数量，便于用户确认。

### 3. list --remote 复用 listAll flag

**选择**：`list --remote` 复用已有的 `--all`（`listAll`）flag，当设置时不过滤 `exclude_repos`。

**备选方案**：
- A) 为 `--remote` 新增独立 flag → 用户心智负担增加
- B) 始终过滤，无例外 → 用户无法查看完整远程列表

**理由**：`--all` 在 `list`（本地）中已有"包含被排除条目"的语义，远程查询复用同一 flag 保持行为一致性。

### 4. list --remote 同时跳过禁用的 group/resource

**选择**：`list --remote` 在遍历 groups 时增加 `enabled` 检查，跳过禁用的 group 和 resource（除非 `--all`）。

**理由**：当前 `runListRemoteRepos` 不检查 `enabled` 状态，与 sync 不一致。统一增加 enabled 检查使行为更加一致。

## Risks / Trade-offs

- **sync 跳过被排除仓库后，取消排除需重新 sync** → 用户从 `exclude_repos` 移除仓库名后，需要重新运行 `grepom sync` 才能将该仓库加入配置。这是合理的行为，且 sync 原本就是增量追加的。
- **list --remote 过滤改变现有行为** → 之前 `list --remote` 显示所有仓库，现在默认过滤。用户可通过 `--all` 恢复原行为。这是行为优化而非 breaking change。
