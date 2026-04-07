## Context

当前 `Resource.APIURL()` 硬编码返回 `https://` 前缀的 URL。当自建 GitLab 服务器仅开放 HTTP 端口时，`sync` 和 `group-list` 命令会因连接 443 端口被拒而失败。`ServerURL` 通过 `ListReposParams` / `ListGroupsParams` 传入 provider，provider 直接使用该 URL 发起 HTTP 请求，没有 fallback 机制。

此外，`grepom clone` 在顺序克隆模式（`concurrency=1`）下缺少最终结果摘要，用户无法得知总体成功/失败情况。

## Goals / Non-Goals

**Goals:**
- 支持通过 URL 前缀指定 API 协议（`http://`/`https://`）
- 无前缀时自动 fallback（先 HTTPS 后 HTTP）
- `grepom clone` 顺序模式增加最终结果摘要

**Non-Goals:**
- 不新增独立的 `protocol` 配置字段
- 不实现指数退避重试、熔断等通用重试机制
- 不修改 Git clone 的认证优先级逻辑

## Decisions

### 1. 通过 URL 前缀推断协议，不新增字段

**选择**: 解析 Resource URL 时保留协议前缀信息，`APIURL()` 根据前缀决定协议。无前缀时为 auto 模式。

**理由**: 减少配置复杂度，协议与 URL 是一体的，不需要额外字段。`stripScheme()` 改为 `parseURL()`，返回 `(host, scheme)`，scheme 为空表示 auto。

### 2. auto 模式默认行为

**选择**: URL 无协议前缀时，先尝试 HTTPS，连接失败（`net.OpError`）自动回退 HTTP。

**理由**: 兼顾公网（HTTPS）和内网（HTTP）场景，开箱即用。

### 3. fallback 在调用方（cmd 层）

**选择**: 在 `cmd/sync.go` 和 `cmd/group-list.go` 中实现 fallback 包装。

**理由**: provider 保持纯粹的 API 客户端职责，fallback 是调用层的一次性连接探测。

### 4. clone 结果摘要

**选择**: 在顺序克隆结束后，统计成功/失败数量并输出摘要（与并行模式一致）。

**理由**: 保持两种克隆模式的输出一致性。

## Risks / Trade-offs

- **HTTPS 失败延迟**: auto 模式在 HTTP-only 服务器上会多一次 HTTPS 连接超时（默认 30s）。用户可通过在 URL 中显式指定 `http://` 来避免。
- **安全提醒**: HTTP 传输 token 是明文的，auto fallback 成功时打印警告。
