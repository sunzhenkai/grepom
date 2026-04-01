## Why

当前 grepom 在 clone 仓库时仅支持 SSH → HTTP 的简单回退策略，未能利用已配置的 resource token 进行认证克隆。许多企业环境中 SSH 密钥并非普遍可用，但 API token 是已配置好的（用于 sync 等功能），优先使用 token 认证可以减少 clone 失败和手动干预。

同时，同一 resource 下的不同 group/repo 可能需要不同的认证方式。例如，某个 repo 需要专用的 SSH 密钥，某个 group 需要使用特定的 deploy token。当前配置模型不支持在 group/repo 级别覆盖认证信息，只能使用 resource 级别的全局 token。

Resource 级别目前也只支持 token，不支持 SSH key。但实际场景中，一个 resource 可能只配置了 SSH key（无 token），此时 clone 也应优先使用 resource 级别的 SSH key。

此外，随着功能不断增长（init、add resource/group/repo、sync、clone、pull、status 等），命令行 flag 组合越来越复杂。新手用户需要查阅文档才能完成基本配置。引入交互式引导模式可以显著降低使用门槛。

在 clone 过程中，用户无法感知系统正在尝试哪种认证方式，调试困难。应该在每次尝试认证方式时输出日志，让用户了解当前进度。

## What Changes

- 新增克隆认证优先级链：group/repo 级别 token → group/repo 级别 SSH key → resource token → resource SSH key → 推导的 SSH → 裸 HTTP 克隆
- Resource 配置新增可选的 `ssh_key` 字段，作为 token 之后的二级认证
- Group 和 Repo 配置新增可选的 `ssh_key` 和 `token` 字段，允许覆盖 resource 级别的认证
- `git.Clone` 函数扩展，使用 `CloneOptions` 结构体接收 token 和 SSH key 参数
- clone 尝试每种认证方式时打印日志（如 "尝试 token 认证..."、"token 认证失败，尝试 SSH key..."）
- 新增 `grepom interactive` 交互式命令，提供逐步引导的配置和操作流程
- 交互模式支持：初始化配置、添加资源、添加组、添加仓库、同步和克隆等常见操作
- 交互模式自动提供默认值和选项提示，减少用户记忆负担
- `add resource`、`add group`、`add repo` 命令新增 `--ssh-key` flag；`add group` 和 `add repo` 新增 `--token` flag

## Capabilities

### New Capabilities
- `clone-auth-priority`: 克隆仓库时的认证优先级策略，支持多级别认证覆盖和有序回退，包含认证尝试日志
- `interactive-mode`: 交互式命令模式，为 init、add、sync、clone 等操作提供逐步引导的用户体验

### Modified Capabilities
- `cli-commands`: 新增 `interactive` 子命令；`clone` 命令内部认证逻辑变更并增加日志；`add resource/group/repo` 新增 `--ssh-key` 和 `--token` flag
- `group-management`: Group 结构新增可选 `ssh_key` 和 `token` 字段
- `resource-management`: Resource 结构新增可选 `ssh_key` 字段；独立 Repo 结构新增可选 `ssh_key` 和 `token` 字段

## Impact

- **config/config.go**: `Resource` 结构新增 `ssh_key` 字段；`Group`、`Repo`、`GroupRepo` 结构新增 `ssh_key` 和 `token` 可选字段；配置校验和 token 解析逻辑需更新
- **git/git.go**: `Clone` 函数使用 `CloneOptions` 结构体重构，支持 token 认证 URL 构建、SSH key 指定、认证尝试日志
- **cmd/clone.go**: clone 命令需传递完整认证信息到 Clone 函数
- **cmd/add.go**: `add resource` 新增 `--ssh-key` flag；`add group` 和 `add repo` 新增 `--ssh-key` 和 `--token` flag
- **cmd/interactive.go**: 新增文件，实现交互式命令
- **repo/resolver.go**: Resolver 需要合并各级别认证信息（token、SSH key）到 resolved repo 结构中
- **provider/provider.go**: `Repo` 结构新增 `Token` 和 `SSHKey` 字段
- **依赖**: 需要引入 `github.com/AlecAivazis/survey/v2` 交互式 prompt 库
