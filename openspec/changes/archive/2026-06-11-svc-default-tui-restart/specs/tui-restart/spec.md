## ADDED Requirements

### Requirement: TUI 列表视图支持 restart 快捷键
TUI 列表视图中 SHALL 支持通过 `R`（大写）快捷键重启当前选中的服务。系统 SHALL 调用已有的 `Manager.Restart()` 方法执行重启操作。

#### Scenario: 重启运行中的服务
- **WHEN** 用户在 TUI 列表视图中选中一个 running 状态的服务并按下 `R` 键
- **THEN** 系统 SHALL 调用 `Manager.Restart()` 停止并重新启动该服务
- **THEN** 系统 SHALL 在底栏显示类似 "restarted <name>" 的消息
- **THEN** 系统 SHALL 自动刷新服务列表

#### Scenario: 重启已退出的服务
- **WHEN** 用户在 TUI 列表视图中选中一个 exited 状态的服务并按下 `R` 键
- **THEN** 系统 SHALL 调用 `Manager.Restart()` 重新启动该服务
- **THEN** 系统 SHALL 在底栏显示成功消息并刷新列表

#### Scenario: 重启无命令信息的服务
- **WHEN** 用户在 TUI 列表视图中选中一个没有记录命令信息的服务并按下 `R` 键
- **THEN** 系统 SHALL 在底栏显示错误信息（如 "service has no command to restart"）
- **THEN** 服务列表 SHALL 保持不变

#### Scenario: 无选中服务时按下 R 键
- **WHEN** 服务列表为空时用户按下 `R` 键
- **THEN** 系统 SHALL 在底栏显示 "no service selected" 错误信息

### Requirement: TUI 底栏显示 restart 快捷键提示
TUI 列表视图的底栏帮助文本 SHALL 包含 `R` 快捷键的 restart 操作提示，使用户能够发现此功能。

#### Scenario: 底栏包含 restart 提示
- **WHEN** 用户查看 TUI 列表视图
- **THEN** 底栏帮助行 SHALL 包含 "R restart" 或等价的提示文本
