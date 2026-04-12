## Why

当前 group 和 standalone repo 都**强制绑定**一个 resource，但实际使用中存在不需要远程 API 的场景：用户可能只想用 grepom 管理一组手动维护的仓库，进行 clone/pull 等本地操作，而不需要通过 GitLab/GitHub API 自动发现仓库。强制绑定 resource 增加了不必要的配置负担。

## What Changes

- Group 的 `resource` 字段变为**可选**：不绑定 resource 时，repos 列表由用户手动维护，sync 等依赖 API 的操作跳过执行
- Standalone repo 的 `resource` 字段变为**可选**：不绑定 resource 时，必须通过 `url` 字段提供完整的 clone URL
- 依赖 resource 的操作（sync 远程发现、list remote 等）在缺少 resource 时优雅跳过，并输出提示信息
- `add group` 和 `add repo` 命令的 `--resource` 参数变为可选

## Capabilities

### New Capabilities

- `optional-resource-binding`: 定义 group 和 repo 的 resource 字段为可选时的行为规则，包括配置验证、操作跳过逻辑和用户提示

### Modified Capabilities

- `sync-command`: sync 操作需要处理 group 未绑定 resource 的情况，跳过远程发现并提示用户
- `group-management`: group 的增删改查需支持无 resource 的场景
- `resource-management`: resource 查找逻辑需要安全处理缺失 resource 的引用
- `config-path-resolution`: 配置验证需要移除 resource 必填约束

## Impact

- **config/config.go**: `validate()` 移除 resource 必填检查，`Resource` 字段语义变更
- **cmd/add.go**: `--resource` 参数变为可选
- **cmd/sync.go**: 需要处理无 resource 的 group，跳过远程发现
- **repo/resolver.go**: 需要处理 resource 不存在时的解析逻辑
- **cmd/list.go**: `list --remote` 需要跳过无 resource 的 group
- **git/git.go**: clone/pull 需要在无认证信息时使用默认系统 SSH/HTTPS
