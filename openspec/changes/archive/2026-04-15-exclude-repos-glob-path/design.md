## Context

`exclude_repos` 是 group 级别的排除机制，当前实现在 `repo/resolver.go` 的 `IsExcluded` 函数中，仅通过 `==` 对 repo `Name` 做精确匹配。用户在管理大型 GitLab 组（如 `recursive: true`）时，经常需要按远端路径层级批量排除仓库，而非逐个列名称。

当前 `resolveInternal` 中 `GroupRepo.Path`（远端路径，如 `my-org/frontend/web-app`）已经存在，但 `IsExcluded` 只接收 `repoName`，远端路径未传入。

## Goals / Non-Goals

**Goals:**

- `exclude_repos` 中的条目支持 glob 通配符（`*`、`?`、`[`），匹配 repo 的远端路径
- 不含通配符的条目保持原有 Name 精确匹配行为，向后兼容
- 改动范围最小化：仅修改 `IsExcluded` 函数及其调用点

**Non-Goals:**

- 不引入 `**`（递归通配符）支持 — 标准 `filepath.Match` 不支持，暂不引入第三方库
- 不新增配置字段（如 `exclude_paths`），复用 `exclude_repos`
- 不修改 CLI flags 或 `Filter` 结构体
- 不修改 `provider.Repo` 结构体 — 远端路径直接从 `GroupRepo.Path` 传入 `IsExcluded`

## Decisions

### D1: 复用 `exclude_repos` 而非新增配置字段

**选择**: 在现有 `exclude_repos` 字段上扩展匹配能力。

**理由**: 无通配符的条目保持精确匹配 Name 的行为，含通配符时自动切换为 glob 匹配远端路径。用户无需学习新字段，配置文件无需结构变更。

**备选**: 新增 `exclude_paths` 字段。放弃因为：两个独立列表增加用户认知负担，且隔离后反而无法在同一个列表中混用 Name 精确匹配和 Path glob 匹配。

### D2: 使用 `strings.ContainsAny(pattern, "*?[")` 区分匹配模式

**选择**: 通过检测通配符字符来决定走精确匹配还是 glob 匹配。

**理由**: 简单直观。repo Name 不含这些字符（合法 repo 名不含 `*?[]`），不会产生误判。

**备选**: 无条件对所有 pattern 都执行 glob 匹配。放弃因为：`filepath.Match("deprecated-app", "deprecated-app")` 虽然也能命中，但 glob 匹配的语义是"对路径匹配"，而非"对名称匹配"。区分两种模式更清晰。

### D3: Glob 匹配目标是远端路径，使用 `filepath.Match`

**选择**: 对 `GroupRepo.Path`（远端路径，如 `my-org/frontend/web-app`）执行 `filepath.Match`。

**理由**: 远端路径天然是层级结构，glob 模式语义清晰。`filepath.Match` 是标准库，零依赖。

### D4: 远端路径不需要存入 `provider.Repo`

**选择**: `IsExcluded` 直接在 `resolveInternal` 内部调用时传入 `gr.Path`，不存入 `provider.Repo`。

**理由**: 远端路径仅在排除判断时需要，不影响后续 clone/pull/status 等操作。避免修改 `provider.Repo` 结构体。

## Risks / Trade-offs

- **[风险] `filepath.Match` 不支持 `**`** → 缓解：用户可用 `*/*/*` 等多层级写法达到类似效果；如后续需求强烈可引入 `github.com/gobwas/glob`
- **[风险] 大量 exclude_repos 条目时的性能** → 缓解：`filepath.Match` 很轻量，且排除列表通常很短（< 20 条），无需担心
- **[权衡] glob 匹配的是远端路径而非 repo Name** → 这是刻意选择：远端路径才是层级化的，适合 glob 模式匹配
