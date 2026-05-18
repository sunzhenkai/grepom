## Why

在使用 `grepom mr` 创建 Merge Request 时，如果仓库的 remote URL 包含嵌入凭据（如 `https://oauth2:TOKEN@gitlab.company.com/org/repo.git`，这是 GitLab CI 等场景的常见模式），provider 检测会失败，报错 "cannot determine provider for host"。同时，`grepom sync` 在同一组内存在重复仓库的问题 —— 当 provider 返回的仓库列表包含重复条目，或 `newGroupRepos` 批次内部未做去重时，同一仓库的 repo 记录会被多次写入配置文件。

## What Changes

- **修复 `extractHost()` 中 URL 凭据导致 provider 检测失败**：在 `cmd/mr.go` 的 `extractHost()` 函数中，去除 URL scheme 后、在提取 host 之前，先剥离 `userinfo@` 部分（即 `@` 之前的所有内容）。同时，`ssh://` 协议的 host 提取路径也会一并修复，以处理 `ssh://user@host/path` 格式。
- **修复 `grepom sync` 同一组内仓库重复**：在 `cmd/sync.go` 构建 `newGroupRepos` 列表时，增加对列表内部 URL 的去重检查；同时在 `config.SyncGroupRepos()` 中也增加对传入 `newRepos` 批次的自去重保护，形成双重保障。

## Capabilities

### New Capabilities

- `url-userinfo-stripping`: 从 git remote URL 中提取 host 时，正确剥离 `userinfo@` 部分（支持 `https://user:token@host/path`、`ssh://user@host/path` 格式）。

### Modified Capabilities

- `sync-command`: `grepom sync` 在同一组内写入仓库列表时，增加对新增批次内部的自去重逻辑，防止同一 URL 的仓库在单次 sync 中被重复添加。

## Impact

- **`cmd/mr.go`** — `extractHost()` 函数需要修改 HTTPS 和 SSH 两种协议路径的 host 提取逻辑
- **`cmd/sync.go`** — `newGroupRepos` 构建循环需要增加 URL 去重
- **`config/config.go`** — `SyncGroupRepos()` 需要增加对 `newRepos` 参数的自去重
- 无配置文件格式变化，无 API 变化，向后完全兼容
