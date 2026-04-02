## Why

当前 clone 认证优先级链中，resource 级别的 token 排在 resource SSH 和 default SSH 之前，导致 SSH 配置可用时仍会优先尝试 token 认证。实际场景中，SSH 认证（尤其是系统默认 SSH agent）通常比 token 更稳定可靠，应优先使用。例如用户执行 `grepom clone` 时，即便本机已配置好 SSH key，仍然会先尝试 resource token（可能因 token 过期而失败），体验不佳。

## What Changes

- **调整 resource 级别认证优先级**：将 resource SSH key 认证和 default SSH 认证都提升到 resource token 认证之前
- **新增优先级级别**：在 group/repo 认证和 resource token 之间插入 "resource SSH" 和 "default SSH" 两个级别
- **修改 `buildAuthStrategies()` 函数**：调整策略构建顺序，并移除 resource SSH 对 `HasGroupSSHKey` 的互斥条件

新的优先级链：
1. group/repo SSH key
2. group/repo token
3. resource SSH key
4. default SSH（系统默认 SSH agent）
5. resource token
6. bare HTTP

## Capabilities

### New Capabilities

（无新增能力）

### Modified Capabilities

- `clone-auth-priority`: 调整克隆认证优先级链，resource SSH 和 default SSH 优先于 resource token

## Impact

- `git/git.go`：`buildAuthStrategies()` 函数的构建逻辑需重写
- `git/git.go`：`CloneOptions` 中 `HasGroupSSHKey` 和 `HasGroupToken` 的使用方式需调整
- `repo/resolver.go`：resolver 合并逻辑可能需要调整，以支持 group SSH 和 resource token 同时存在的场景
- 现有 `clone-auth-priority` spec 需要更新优先级描述
