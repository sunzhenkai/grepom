## ADDED Requirements

### Requirement: example 命令导出示例配置
系统 SHALL 提供 `grepom example` 子命令，将包含全部功能和字段注释的完整示例 YAML 配置输出到 stdout。

#### Scenario: 默认输出到 stdout
- **WHEN** 用户运行 `grepom example`
- **THEN** 系统将完整示例配置输出到 stdout，包含 github、gitlab、generic 三种 provider 的 resource 示例、group 示例、独立 repo 示例，以及每个字段的 YAML 注释说明

#### Scenario: 输出到文件
- **WHEN** 用户运行 `grepom example --output my-config.yml`
- **THEN** 系统将示例配置写入 `my-config.yml` 文件，并输出提示信息

#### Scenario: 输出到文件使用短标志
- **WHEN** 用户运行 `grepom example -o my-config.yml`
- **THEN** 系统行为与 `--output` 完全一致

#### Scenario: 输出文件已存在
- **WHEN** 用户运行 `grepom example -o existing.yml`，且 `existing.yml` 已存在
- **THEN** 系统报错提示文件已存在，不覆盖

### Requirement: example 命令示例配置内容完整性
示例配置 SHALL 包含所有支持的配置字段，每个字段附带 YAML 注释说明其用途和可选值。

#### Scenario: 示例配置包含所有 provider 类型
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出的示例配置包含 `github`、`gitlab`、`generic` 三种 provider 的 resource 示例

#### Scenario: 示例配置包含所有可选字段
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出包含 `ssh_key`、`enabled`、`exclude_repos`、`recursive`、`local_path`、`token`（group/repo 级别覆盖）等所有可选字段
