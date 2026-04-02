## Why

当前配置文件的 `url` 字段要求包含完整协议前缀（如 `https://g.wii.pub`），既冗余又不够灵活。同时 clone 时的认证优先级链中，最后一个"裸 HTTP"策略（策略 #6）会在无 token 的情况下触发 git 的交互式认证提示（`Username for 'http://...'`），导致自动化流程被阻塞，用户体验很差。

需要两个改进：
1. **简化配置 URL**：`url` 只需填写 host（可选 port），系统自动推导 SSH/HTTPS/HTTP 三种协议的 URL
2. **移除裸 HTTP 回退**：去掉会导致交互式认证提示的裸 HTTP 策略，避免 clone 流程被阻塞

## What Changes

- **BREAKING** 移除认证优先级链中的"裸 HTTP URL"策略（原策略 #6），避免无 token 时触发 git 交互式认证
- 修改 resource `url` 的语义：仅表示 host（可选 port），不再需要包含协议前缀
- 新增 URL 推导逻辑：从 host:port 自动生成 SSH、HTTPS、HTTP 三种 URL
- clone 日志中打印实际尝试的 URL（而非仅策略标签），便于调试
- 更新 `deriveSSHURL` 使其基于 host:port 推导，而非从 HTTPS URL 转换
- 更新 provider API 调用逻辑，确保 `ServerURL` 仍使用 HTTPS（API 调用始终需要协议前缀）

## Capabilities

### New Capabilities
- `auto-protocol-url`: 从 host:port 自动推导 SSH/HTTPS/HTTP 三种克隆 URL，并打印实际尝试的 URL

### Modified Capabilities
- `resource-management`: resource `url` 字段语义变更——仅表示 host（可选 port），不再要求包含协议前缀，API 调用由系统自动补全 HTTPS
- `clone-auth-priority`: 移除"裸 HTTP URL"策略（原策略 #6），clone 日志改为打印实际 URL

## Impact

- **配置文件格式**: 已有配置中 `url: https://xxx` 的写法仍兼容（系统识别并剥离协议前缀提取 host），但推荐简化为 `url: xxx`
- **config/config.go**: 修改 `validate()` 中 URL 前缀补全逻辑，改为提取 host 并存储；新增从 host 推导 API URL 的方法
- **git/git.go**: `buildAuthStrategies()` 移除裸 HTTP 策略；`Clone()` 日志打印实际 URL
- **repo/resolver.go**: `deriveSSHURL()` 改为从 host:port 推导，而非从 HTTPS URL 转换
- **provider/gitlab.go, provider/github.go**: `ListRepos` 中 `ServerURL` 使用从 host 推导的 HTTPS URL
- **测试文件**: 更新相关单元测试
