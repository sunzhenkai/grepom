## ADDED Requirements

### Requirement: CLI 子命令支持唯一前缀匹配
系统 SHALL 支持用户输入子命令的唯一前缀来执行对应命令，无需输入完整命令名。

#### Scenario: 唯一前缀匹配成功
- **WHEN** 用户运行 `grepom cl`
- **THEN** 系统匹配到 `clone` 命令并执行

#### Scenario: 单字符唯一前缀
- **WHEN** 用户运行 `grepom i`
- **THEN** 系统匹配到 `interactive` 命令并执行

#### Scenario: 前缀有歧义
- **WHEN** 用户运行 `grepom s`（同时匹配 `sync`、`status`、`search`）
- **THEN** 系统报错并提示可能匹配的命令列表

#### Scenario: 完整命令名仍然有效
- **WHEN** 用户运行 `grepom interactive`
- **THEN** 系统正常执行 `interactive` 命令
