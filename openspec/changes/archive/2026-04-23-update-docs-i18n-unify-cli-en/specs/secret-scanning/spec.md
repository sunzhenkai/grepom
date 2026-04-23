## MODIFIED Requirements

### Requirement: 用户可以执行 scan 命令扫描仓库敏感信息
系统 SHALL 提供 `grepom scan` 命令。命令的 cobra 描述（Short、Long、Example）、flag 帮助文本、错误信息和用户可见输出 SHALL 全部使用英文。

#### Scenario: scan 命令帮助信息为英文
- **WHEN** 用户运行 `grepom scan --help`
- **THEN** 系统显示英文的命令描述、flag 说明和示例

#### Scenario: scan 错误信息为英文
- **WHEN** scan 过程中发生错误（如扫描失败、无法创建输出文件、JSON 序列化失败）
- **THEN** 系统输出英文错误信息

#### Scenario: 无配置文件时扫描当前目录
- **WHEN** 用户执行 `grepom scan` 且当前目录及默认路径下没有配置文件
- **THEN** 系统扫描当前工作目录下的所有文件，在 stderr 输出英文提示信息 "Scanning current directory (no config file found)..."
