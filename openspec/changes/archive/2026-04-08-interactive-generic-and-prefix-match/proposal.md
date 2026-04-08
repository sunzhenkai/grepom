## Why

interactive 模式和 `add resource` 命令中 provider 选项硬编码为 `["gitlab", "github"]`，新增的 `generic` provider 无法被选择。此外，CLI 子命令名称较长（如 `interactive`、`status`），用户希望通过前缀匹配（如 `grepom i` → `interactive`）快速调用。

## What Changes

- 将 interactive 模式和 `add resource` 命令中所有硬编码的 provider 列表替换为动态获取（包含 `generic`）
- 更新 `interactiveInit` 中根据 provider 自动推导资源名称的逻辑，支持 `generic`
- 更新 `interactiveInit` 中根据 provider 推导默认 URL 的逻辑，`generic` 不设默认 URL
- 启用 Cobra 的 `ValidArgs` 或自定义前缀匹配，允许用户输入子命令的唯一前缀即可执行

## Capabilities

### New Capabilities
- `command-prefix-match`: 支持 CLI 子命令前缀匹配，用户输入唯一前缀即可执行对应命令

### Modified Capabilities
- `interactive-mode`: provider 选择列表新增 `generic`，自动推导资源名称和默认 URL 支持 `generic`
- `add-command-validation`: `add resource` 命令的 provider 验证支持 `generic`

## Impact

- `cmd/interactive.go`：`interactiveInit`、`interactiveAddResource` 中的 provider 选项列表和自动推导逻辑
- `cmd/add.go`：`addResourceCmd` 中的 provider 验证和帮助文本
- `cmd/root.go`：启用前缀匹配
