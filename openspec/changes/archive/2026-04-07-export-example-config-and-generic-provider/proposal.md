## Why

当前 `grepom init` 仅生成最小化的空配置文件，用户需要查阅文档才能了解所有可用的配置字段和功能。新增 `example` 命令可导出一份包含全部功能的完整示例配置，帮助用户快速理解和上手。

此外，现有 resource provider 仅支持 GitHub 和 GitLab，无法管理不依赖特定平台 API 的纯 Git 仓库。新增通用 Git URL provider（`generic`）可以让用户通过直接指定 Git URL 来管理任何 Git 仓库，无需平台 API 支持。

## What Changes

- 新增 `grepom example` 子命令，输出一份包含全部功能的完整示例 YAML 配置到 stdout 或写入文件
- 示例配置包含所有 provider 类型（github、gitlab、generic）的 resource、group、独立 repo 的完整配置，附带中文注释说明每个字段
- 新增 `generic` provider 类型，支持通过纯 Git URL 管理仓库，不依赖任何平台 API
- `generic` provider 不支持 `ListRepos` 和 `ListGroups` API 调用（因为没有远程 API 可查询），仅用于认证和克隆
- 更新配置验证逻辑，允许 `generic` 作为合法的 provider 值
- 更新 git 认证逻辑，为 `generic` provider 添加 token 用户名映射

## Capabilities

### New Capabilities
- `example-command`: 新增 `grepom example` 子命令，导出包含全部功能的完整示例 YAML 配置
- `generic-provider`: 新增通用 Git URL provider，支持不依赖平台 API 的纯 Git 仓库管理

### Modified Capabilities
- `resource-management`: 新增 `generic` 作为合法的 provider 类型，更新验证规则
- `cli-commands`: 新增 `example` 子命令到命令列表

## Impact

- **代码**：`provider/` 目录新增 `generic.go`；`cmd/` 目录新增 `example.go`；`config/config.go` 更新验证逻辑；`git/git.go` 更新认证映射
- **API**：Provider 接口无变更，`generic` provider 实现返回空列表
- **依赖**：无新增外部依赖
- **兼容性**：完全向后兼容，现有配置无需修改
