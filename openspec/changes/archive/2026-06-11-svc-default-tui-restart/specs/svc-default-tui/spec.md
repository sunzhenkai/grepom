## ADDED Requirements

### Requirement: svc 命令无参数时默认启动 TUI
当用户执行 `grepom svc` 或 `grepom service` 不带任何子命令和参数时（`--shell` 标志除外），系统 SHALL 直接启动 TUI 交互界面，而非打印帮助文本。

#### Scenario: 无参数执行 svc 命令
- **WHEN** 用户执行 `grepom svc`（无子命令、无参数）
- **THEN** 系统 SHALL 启动 TUI 交互界面，行为与 `grepom svc tui` 完全一致

#### Scenario: 无参数执行 service 别名
- **WHEN** 用户执行 `grepom service`（无子命令、无参数）
- **THEN** 系统 SHALL 启动 TUI 交互界面

#### Scenario: 通过 --help 获取帮助
- **WHEN** 用户执行 `grepom svc --help`
- **THEN** 系统 SHALL 正常显示 svc 命令的帮助文本

#### Scenario: --shell 标志仍然可用
- **WHEN** 用户执行 `grepom svc --shell`
- **THEN** 系统 SHALL 打印 gsvc() shell 函数，不启动 TUI

#### Scenario: 无 TTY 环境下的错误处理
- **WHEN** 用户在无 TTY 环境（如管道、CI）中执行 `grepom svc`
- **THEN** 系统 SHALL 返回错误信息 "svc tui requires a TTY"
