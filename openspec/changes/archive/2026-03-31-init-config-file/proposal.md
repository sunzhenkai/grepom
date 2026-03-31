## Why

当前 `grepom init` 的语义是 "clone 仓库"，这与 CLI 工具的惯例不符——`init` 通常意味着"初始化项目/配置文件"。同时，grepom 缺少一个真正意义上的初始化命令：用户首次使用时需要手动创建 `.grepom.yml` 配置文件，无法通过工具本身完成从零开始的配置闭环。应该将 `init` 重新定义为初始化配置文件，并将原 clone 逻辑迁移到更语义化的命令（如 `clone`）。

## What Changes

- 将当前 `init` 命令重命名为 `clone`，保持原有 clone 仓库功能不变
- 新增 `init` 命令，负责在当前目录创建 `.grepom.yml` 配置文件
  - 支持 `--base` 参数指定仓库根目录（默认 `~/projects`）
  - 交互式引导用户添加第一个 source（provider、url、group/org）
  - 配置文件不存在时创建，已存在时提示
- `init` 命令与现有 `add source` / `add repo` 配合，实现配置文件的完整闭环管理

## Capabilities

### New Capabilities

- `init-command`: `grepom init` 命令实现，创建配置文件并可选引导用户添加第一个 source

### Modified Capabilities

- `cli-commands`: `init` 命令语义变更（clone → 配置初始化），新增 `clone` 命令

## Impact

- **命令行接口**: `grepom init` 行为变更（**BREAKING**），原 `init` 功能移至 `grepom clone`
- **代码**: `cmd/init.go` 重写为配置初始化逻辑，新增 `cmd/clone.go`（原 init 逻辑迁移）
- **文档**: help 文本和 example 需要更新
- **用户**: 已有用户需要将 `grepom init` 改为 `grepom clone`
