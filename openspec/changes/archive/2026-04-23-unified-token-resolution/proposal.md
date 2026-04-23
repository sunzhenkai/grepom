## Why

上一轮改动将 token 环境变量解析从 `config.Load()` 延迟到 `Resolver`，解决了"不用的 resource 环境变量未设置导致整个程序无法启动"的问题。但 Resolver 只是 token 的使用方之一——`sync`、`list --remote`、`pipeline`、`interactive sync` 等命令直接从 `cfg.Resources[name].Token` 取值，绕过了 Resolver，拿到的仍是 `${GITLAB_TOKEN}` 原始占位符字符串。这导致延迟解析只修了一半路径，另一半反而让错误信息变得更难理解（provider 拿 `${GITLAB_TOKEN}` 字符串当 token 去调 API，返回 401 而非清晰的"环境变量未设置"报错）。

此外，用户在 YAML 或 CLI 参数中书写 token 时可能用单引号或双引号包裹（如 `token: '${GITLAB_TOKEN}'` 或 `--token "'${GITLAB_TOKEN}'"`），当前 `ResolveToken` 不会去除引号，导致匹配失败。

## What Changes

- 在 `Resource` 上新增 `ResolvedToken() (string, error)` 方法，封装"去引号 + 解析环境变量"两步操作，作为获取 token 实际值的唯一推荐入口
- 在 `ResolveToken()` 内部增加引号清理逻辑：当 token 值被成对的单引号或双引号包裹时，自动去除首尾引号
- 修复 `cmd/sync.go`、`cmd/list.go`（两处）、`cmd/pipeline.go`、`cmd/interactive.go` 中直接使用 `res.Token` 的问题，统一改为调用 `res.ResolvedToken()`
- 将 `repo/resolver.go` 中直接调用 `config.ResolveToken(token)` 改为通过 `Resource.ResolvedToken()` 或对 override token 调用 `config.ResolveToken()`（已内置引号清理）

## Capabilities

### New Capabilities
- `unified-token-resolution`: 所有命令通过统一的 `Resource.ResolvedToken()` 方法获取已解析的 token，消除绕过 Resolver 直接使用原始 token 的路径

### Modified Capabilities
- `token-env-placeholder`: REQUIREMENTS 变更——`ResolveToken()` 增加 `stripQuotes` 行为，兼容引号包裹的占位符
- `lazy-token-resolution`: REQUIREMENTS 变更——延迟解析不再仅限于 Resolver，而是通过 `Resource.ResolvedToken()` 在所有使用点生效
- `sync-command`: REQUIREMENTS 变更——sync 命令使用 `Resource.ResolvedToken()` 获取 token 而非直接读取 `res.Token`
- `list-remote-repos`: REQUIREMENTS 变更——list --remote 使用 `Resource.ResolvedToken()` 获取 token
- `pipeline-list`: REQUIREMENTS 变更——pipeline 命令使用 `Resource.ResolvedToken()` 获取 token

## Impact

- **核心代码**：`config/config.go`（新增 `ResolvedToken()` 方法、`ResolveToken()` 增加引号清理）、`repo/resolver.go`（统一调用方式）
- **CLI 命令**：`cmd/sync.go`、`cmd/list.go`、`cmd/pipeline.go`、`cmd/interactive.go`（`res.Token` → `res.ResolvedToken()`）
- **测试**：新增 `ResolvedToken()` 测试、引号清理测试、修改现有 sync/pipeline 相关测试
- **向后兼容**：配置文件格式不变，纯行为修复
