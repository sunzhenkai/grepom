## Why

当前 `config.Load()` 在加载配置时立即解析所有 resource/group/repo 的 `${ENV_VAR}` token。如果任何一个环境变量未设置，整个加载过程就会失败——即使该 resource 被标记为 `enabled: false`，或者当前命令根本不需要使用它。这导致用户在只有部分 resource 可用的环境下（如只配置了 GitHub 但没配 GitLab token）无法使用任何功能，即使他们只需要操作其中一个 provider。

## What Changes

- **延迟 token 解析**：`config.Load()` 不再在加载时调用 `resolveToken()`，改为保留原始 `${ENV_VAR}` 字符串，仅在需要实际使用 token 时（如 clone、sync、list 等操作）才解析
- **按需校验**：在 `repo.Resolver` 解析阶段或 provider 调用阶段，只对实际启用的、当前命令涉及的 resource 进行环境变量解析
- **移除 `loadExistingConfig()` 变通方案**：`cmd/add.go` 中的 `loadExistingConfig()` 是为了绕过加载时校验而存在的，延迟检查后不再需要

## Capabilities

### New Capabilities
- `lazy-token-resolution`: token 的环境变量解析从配置加载时延迟到实际使用时，只在需要时解析和校验

### Modified Capabilities
- `token-env-placeholder`: REQUIREMENTS 变更——解析时机从"加载时立即解析"变为"使用时按需解析"，保留 write-back 行为不变
- `resource-management`: REQUIREMENTS 变更——resource 校验不再包含 token 环境变量检查，该检查移至使用时

## Impact

- **核心代码**：`config/config.go`（`Load()` 函数移除 eager token 解析）、`repo/resolver.go`（在解析时增加 token 解析逻辑）
- **CLI 命令**：`cmd/add.go`（移除 `loadExistingConfig()` 变通方案）
- **测试**：`config/config_test.go`、`repo/resolver_test.go` 需要更新测试用例
- **向后兼容**：配置文件格式不变，纯行为变更，不影响已正常工作的用户
