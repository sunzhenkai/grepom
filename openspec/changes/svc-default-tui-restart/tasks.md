## 1. svc 命令默认启动 TUI

- [x] 1.1 修改 `cmd/svc.go` 中 `svcCmd.RunE`：在非 `--shell` 模式下调用 `runSvcTui` 函数替代 `cmd.Help()`，同时确认 `--help` 由 Cobra 自动处理不受影响
- [x] 1.2 验证 `serviceCmd`（alias）自动继承新行为，无需额外修改

## 2. TUI restart 快捷键

- [x] 2.1 在 `service/tui/model.go` 中新增 `restart()` 方法：选中服务后调用 `m.mgr.Restart()`，成功时设置 `m.message` 并调用 `m.refresh()` 刷新列表，失败时设置错误信息
- [x] 2.2 在 `service/tui/tui.go` 的 `Update` 方法中，为 `viewList` 模式添加 `R` 键处理分支，调用 `m.restart()` 方法
- [x] 2.3 更新 `service/tui/model.go` 中 `listView()` 的底栏帮助文本，在键位提示中添加 "R restart"

## 3. 测试

- [x] 3.1 为 `model.restart()` 方法添加单元测试：覆盖成功重启、无选中服务、重启失败等场景（在 `service/tui/model_test.go` 中）
- [x] 3.2 手动验证：`grepom svc` 无参数直接启动 TUI，`grepom svc --help` 正常显示帮助
- [x] 3.3 手动验证：TUI 中 `R` 键可重启服务，底栏显示提示，列表自动刷新

## 4. 文档更新

- [x] 4.1 更新 README.md（中文版）中 svc 命令说明，反映默认 TUI 行为和 TUI 快捷键 `R`
- [x] 4.2 更新 README.md（英文版）中对应内容
