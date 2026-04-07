## Why

`grepom sync` 在调用 GitLab/GitHub API 时硬编码使用 HTTPS 协议，当自建 GitLab 服务器仅开放 HTTP（如内网环境 `g.wii.pub:80`）时，API 请求会因连接 443 端口失败而报错。用户需要一种方式为 resource 指定 API 协议，并支持自动 fallback。此外，`grepom clone` 在顺序克隆模式下缺少最终结果摘要。

## What Changes

- Resource URL 支持显式协议前缀（`http://`、`https://`），无前缀时自动 fallback（先 HTTPS 后 HTTP）
- URL 无前缀时，API 调用先尝试 HTTPS，连接失败自动回退 HTTP
- URL 带 `http://` 前缀时，仅使用 HTTP
- URL 带 `https://` 前缀时，仅使用 HTTPS
- `grepom clone` 顺序克隆模式增加最终结果摘要

## Capabilities

### New Capabilities

_(无新增 capability)_

### Modified Capabilities

- `auto-protocol-url`: API URL 推导逻辑从"始终 HTTPS"改为从 URL 前缀推断协议，无前缀时 auto fallback
- `resource-management`: Resource URL 解析逻辑更新，保留协议前缀信息用于 API 调用

## Impact

- **config 包**: `Resource` 结构体内部记录协议信息，`APIURL()` 方法逻辑变更
- **cmd 包**: sync/group-list 命令接入 fallback 逻辑，clone 命令增加结果摘要
- **CLI 输出**: sync 命令在 fallback 时提示用户；clone 命令显示最终摘要
