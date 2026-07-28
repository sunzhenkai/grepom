## Why

首次配置 grepom 时，文档/示例与实现多处脱节：`resources` 旧 list 格式直接报错、generic 会错误拼接完整 URL、HTTPS 示例实际强行走 SSH、认证失败几乎无法诊断、`virtual_groups` 无法覆盖独立 repos。需要用兼容逻辑兜住旧写法与直觉写法，并同步 README / example 说明，降低试错成本。

## What Changes

- **resources list 兼容**：加载配置时同时接受 map（正式格式）与带 `name` 字段的 list（README 旧格式），list 在内存中归一为 map；文档统一改为 map，并注明 list 仍可解析。
- **repo.url 智能解析**：相对路径继续与 `resource.url` 拼成 `git@host:path` / `https://host/path`；若 `repo.url` 已是完整 `https://` / `http://` / `git@` / `ssh://` URL，则**不再二次拼接**，直接使用（修正「写完整 URL 反而拼坏」）。
- **尊重声明协议**：完整 HTTPS/HTTP URL 优先走 HTTPS 克隆策略（含无 token 的匿名 HTTPS，用于公开仓）；完整 SSH URL 走 SSH；相对路径保持现有 SSH→token 优先级链。
- **克隆错误诊断**：`git clone`/`pull` 失败时保留并输出原始 stderr（脱敏后），`-v` 展示每步策略的完整失败原因，不再只剩 `all authentication methods failed`。
- **默认 SSH 身份兜底**：未配置 `ssh_key` 且 default SSH 失败时，依次尝试常见默认私钥文件（如 `~/.ssh/id_ed25519`、`~/.ssh/id_rsa`），减少「agent 未起就必须手写 ssh_key」的摩擦；仍推荐显式配置 `ssh_key`。
- **virtual_groups 支持 repos**：`virtual_groups.<name>` 新增可选 `repos` 列表，可引用顶层独立仓库名；`--vgroup` 批量操作同时覆盖成员 groups 与 repos。
- **文档与 example**：更新 `README.md` / `README_en.md` / `grepom example` 注释，说明 map 格式、URL 写法、协议行为、`ssh_key` 建议、vgroup 的 repos；示例仅使用虚构 host/path，**不写入任何真实私密仓库信息**。

## Capabilities

### New Capabilities

- `clone-error-diagnostics`: 克隆/拉取失败时透出脱敏后的 git 原始错误，改善定位体验

### Modified Capabilities

- `resource-management`: resources 加载兼容 list→map 归一
- `generic-provider`: 明确相对路径 vs 完整 URL 的解析与克隆行为；HTTPS 完整 URL 可匿名克隆
- `clone-auth-priority`: 按 URL 协议选择策略；增加匿名 HTTPS；未配置 ssh_key 时尝试默认私钥文件
- `virtual-groups`: 支持引用顶层独立 repos
- `example-command`: 示例注释补充 URL/协议/ssh_key/vgroup repos 说明

## Impact

- **配置加载**：`config/config.go`（resources 反序列化兼容）
- **URL 解析 / Resolver**：`repo/resolver.go`（`ExtractRemotePath` / `deriveSSHURL` / 独立 repo 绑定 resource 时的 URL 处理）
- **克隆认证**：`git/git.go`（策略构建、匿名 HTTPS、默认 key 兜底、stderr 透出）
- **虚拟分组**：`config/virtual_groups.go` 及 `--vgroup` 过滤相关命令（list/status/clone/pull/sync 等）
- **文档**：`README.md`、`README_en.md`、`cmd/example.go`
- **测试**：config / resolver / git / virtual_groups 单测与回归
- **无 BREAKING**：旧 map 配置与相对路径写法行为保持；仅放宽兼容与修正明显错误路径
