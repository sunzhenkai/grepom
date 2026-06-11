## Why

当用户执行 `grepom svc` 不带任何子命令时，当前行为仅打印帮助文本，没有实际功能。对于日常管理本地开发服务的场景，用户更希望直接进入 TUI 交互界面来快速查看和管理服务。同时，TUI 中缺少 restart 操作（虽然 CLI 的 `svc restart` 已可用），导致用户必须退出 TUI 才能重启服务，影响使用体验。

## What Changes

- **默认 TUI 模式**：`grepom svc` 不带参数时，从打印帮助文本改为直接启动 TUI 交互界面；可通过 `grepom svc --help` 获取帮助信息
- **TUI restart 快捷键**：在 TUI 列表视图中新增 `R`（大写）快捷键，调用已有的 `Manager.Restart()` 方法重启选中的服务
- **TUI 底栏提示更新**：在底栏键位提示中增加 restart 的说明

## Capabilities

### New Capabilities
- `svc-default-tui`：svc 命令无参数时默认启动 TUI 交互界面
- `tui-restart`：TUI 列表视图中支持通过快捷键重启选中的服务

### Modified Capabilities

## Impact

- **`cmd/svc.go`**：修改 `svcCmd.RunE` 逻辑，无参数时调用 TUI Run 函数而非打印帮助
- **`service/tui/tui.go`**：Update 方法中新增 `R` 键处理，调用 `mgr.Restart()`
- **`service/tui/model.go`**：底栏帮助文本增加 restart 提示
- **`README.md`** 及多语言文档：更新 svc 命令说明，反映默认 TUI 行为和 TUI 快捷键变更
- **`cmd/svc_test.go`**：需要更新或新增测试用例覆盖无参数 TUI 启动逻辑
