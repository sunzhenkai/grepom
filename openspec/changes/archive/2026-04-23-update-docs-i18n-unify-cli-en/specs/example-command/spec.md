## MODIFIED Requirements

### Requirement: example 命令导出示例配置
系统 SHALL 提供 `grepom example` 子命令。命令的 cobra 描述（Short、Long）、flag 帮助文本和输出的示例配置中的 YAML 注释 SHALL 全部使用英文。

#### Scenario: example 命令帮助信息为英文
- **WHEN** 用户运行 `grepom example --help`
- **THEN** 系统显示英文的命令描述和 flag 说明

#### Scenario: 示例配置注释为英文
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出的示例 YAML 配置中所有注释均为英文，描述每个字段的用途和可选值

#### Scenario: 输出到文件时的提示为英文
- **WHEN** 用户运行 `grepom example --output my-config.yml`
- **THEN** 系统将示例配置写入文件并输出英文提示信息
