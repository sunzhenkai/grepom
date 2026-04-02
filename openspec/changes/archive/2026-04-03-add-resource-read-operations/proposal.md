## Why

当前 grepom 的 `list` 命令仅能列出仓库（repos），无法查看已配置的资源（resources）和分组（groups）信息。用户需要打开 YAML 配置文件才能确认 resource 和 group 的配置情况，使用体验不佳。新增 resource 和 group 的读操作，让用户能通过 CLI 快速查看配置概览，提升日常使用效率。

## What Changes

- 新增 `grepom list resources` 子命令，列出所有已配置的资源，显示名称、provider、url 等关键信息
- 新增 `grepom list groups` 子命令，列出所有已配置的分组，显示名称、关联 resource、远端路径、本地路径、recursive 等信息
- 新增 `grepom list groups <name>` 命令，列出指定 group 下的 repo 列表，包含名称、远端路径、克隆状态等

## Capabilities

### New Capabilities
- `resource-list`: 资源列表读取功能，支持列出所有已配置的 resource 并展示关键字段
- `group-list`: 分组列表读取功能，支持列出所有 group 及其属性，以及查看指定 group 下的 repo 列表

### Modified Capabilities
- `cli-commands`: 在 `list` 命令下新增 `resources` 和 `groups` 子命令，扩展 list 命令的读取范围

## Impact

- **代码影响**: `cmd/list.go` 需要重构为支持子命令（`repos`、`resources`、`groups`），保持原有 `grepom list`（等同于 `grepom list repos`）的向后兼容
- **配置层**: 可能需要在 `config` 包中新增辅助方法（如按 resource 统计 group 数量），但不修改配置结构
- **依赖**: 无新增外部依赖
