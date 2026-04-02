## Context

当前 grepom 的 resource `url` 字段要求填写完整 URL（如 `https://g.wii.pub`），系统在验证时自动补全 `https://` 前缀。clone 时通过 `deriveSSHURL()` 从 HTTPS URL 转换得到 SSH URL，形成两种协议的克隆地址。

认证优先级链共 6 级，最后一级"裸 HTTP URL"（无 token 的 HTTPS clone）会在无认证凭据时触发 git 的交互式认证流程，阻塞自动化操作。

## Goals / Non-Goals

**Goals:**
- resource `url` 字段简化为仅 host（可选 port），如 `g.wii.pub` 或 `g.wii.pub:8080`
- 系统自动从 host:port 推导出 SSH、HTTPS、HTTP 三种克隆 URL
- clone 日志打印实际尝试的 URL，便于用户调试
- 移除裸 HTTP 策略，避免交互式认证提示阻塞自动化

**Non-Goals:**
- 不改变 API 调用逻辑（provider 的 `ServerURL` 仍使用 HTTPS）
- 不改变认证优先级链的 token/SSH key 优先级规则
- 不增加用户手动选择协议的选项
- 不处理带路径的 URL（如 `g.wii.pub/git`）

## Decisions

### Decision 1: URL 解析与存储方式

**选择**: config `validate()` 中将 `url` 统一解析为 host:port 格式存储，剥离协议前缀。

**方案**:
- 输入 `https://g.wii.pub` → 存储 `g.wii.pub`
- 输入 `http://g.wii.pub:8080` → 存储 `g.wii.pub:8080`
- 输入 `g.wii.pub` → 存储 `g.wii.pub`（不变）

API 调用时动态拼接 `https://` + host。

**替代方案**: 保留完整 URL 存储，新增 `host` 字段 → 增加配置复杂度，不必要。

### Decision 2: URL 推导逻辑集中化

**选择**: 在 `config.Resource` 上新增方法，按需推导各种协议 URL。

```go
// 在 config 包或独立 urlutil 包中
func (r Resource) Host() string           // 返回 host:port
func (r Resource) APIURL() string          // https://host:port
func (r Resource) HTTPSURL(path) string    // https://host:port/path.git
func (r Resource) HTTPURL(path) string     // http://host:port/path.git
func (r Resource) SSHURL(path) string      // git@host:path.git
```

**替代方案**: 在 resolver 中推导 → URL 推导逻辑分散，不利于 provider 和 git 包复用。

### Decision 3: 移除裸 HTTP 策略

**选择**: 从 `buildAuthStrategies()` 中移除策略 #6（裸 HTTP URL）。

**理由**: 裸 HTTP clone 在无凭据时会触发 git 的交互式认证提示（`Username for 'http://...'`），这在自动化场景中是致命问题。有 token 时已有 token 策略覆盖，无 token 时不应尝试需要认证的协议。

**影响**: 无 token 且无 SSH 访问的仓库将无法克隆，直接报错。这是合理的行为——用户应配置认证信息。

### Decision 4: clone 日志打印实际 URL

**选择**: 在 `Clone()` 函数的日志中，除策略标签外，额外打印实际尝试的 URL。

格式：
```
  [1/4] 尝试 SSH 认证 (默认)... git@g.wii.pub:path/repo.git
  [1/4] 失败: Connection refused
  [2/4] 尝试 token 认证 (resource)... https://oauth2:***@g.wii.pub/path/repo.git
  [2/4] 成功
```

token URL 中的 token 部分用 `***` 替代以避免泄露。

## Risks / Trade-offs

- **[兼容性]** 已有配置文件使用 `https://` 前缀 → 系统自动剥离，完全向后兼容，无破坏性
- **[移除裸 HTTP]** 某些匿名可访问的 HTTP 仓库将无法克隆 → 用户需明确配置 token 或 SSH；实际上匿名 HTTP 仓库极少见，且 SSH 通常也可用
- **[host-only 限制]** 不支持带路径的 URL（如 `company.com/gitlab`）→ 当前用户场景不需要，未来可通过增加 `api_path` 字段支持
