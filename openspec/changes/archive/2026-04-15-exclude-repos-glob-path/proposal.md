## Why

当前 `exclude_repos` 只支持按 repo 的 `name` 字段精确匹配排除。当用户不想克隆某个路径前缀下的所有仓库（如 `my-org/frontend/*`）或想按路径模式批量排除时，只能逐个列出 repo 名称，维护成本高且容易遗漏。

## What Changes

- `IsExcluded` 函数增加 glob 模式匹配能力：当 `exclude_repos` 中的条目包含通配符（`*`、`?`、`[`）时，改为对 repo 的**远端路径**（remote path）进行 `filepath.Match` 匹配
- 不含通配符的条目保持原有的 repo name 精确匹配行为，完全向后兼容

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `exclusion-toggle`: `exclude_repos` 的匹配方式从"仅支持 name 精确匹配"扩展为"name 精确匹配 + 含通配符时对远端路径 glob 匹配"

## Impact

- `repo/resolver.go`：`IsExcluded` 函数签名和逻辑变更，两处调用点需传入远端路径
- `cmd/list.go`：`runListRemoteRepos` 中 `IsExcluded` 调用需传入远端路径
- 无新依赖，使用标准库 `path/filepath.Match`
- 无配置文件结构变更，无 CLI flag 变更，无破坏性变更
