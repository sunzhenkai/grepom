## MODIFIED Requirements

### Requirement: 交互式初始化配置
交互模式 SHALL 引导用户完成配置文件初始化，逐步输入 base 目录、provider 类型、URL、token 和可选的 SSH key。

#### Scenario: 完整初始化流程
- **WHEN** 用户在交互模式中选择"初始化配置"
- **THEN** 系统依次提示：配置文件路径（默认 `.grepom.yml`）、base 目录（默认 `~/projects`）、是否添加资源、provider 类型（选择 gitlab/github/generic）、API URL、token、是否配置 SSH key（可选）

#### Scenario: 选择 generic provider 时无默认 URL
- **WHEN** 用户在初始化流程中选择 provider 为 `generic`
- **THEN** 系统提示输入 URL 时不设默认值，用户必须手动输入

#### Scenario: 选择 generic provider 时默认资源名
- **WHEN** 用户在初始化流程中选择 provider 为 `generic`
- **THEN** 系统使用 `generic` 作为默认资源名称

#### Scenario: 使用默认值
- **WHEN** 系统提示输入 base 目录，用户直接按回车
- **THEN** 系统使用默认值 `~/projects`

### Requirement: 交互式添加资源
交互模式 SHALL 引导用户添加新的 resource 配置，包括 token 和可选的 SSH key。

#### Scenario: 添加资源流程
- **WHEN** 用户在交互模式中选择"添加资源"
- **THEN** 系统依次提示：资源名称、provider 类型（从 gitlab/github/generic 中选择）、API URL、token（支持 `${ENV_VAR}` 语法提示）、是否配置 SSH key（可选）

#### Scenario: 选择 generic provider 时无默认 URL
- **WHEN** 用户在添加资源流程中选择 provider 为 `generic`
- **THEN** 系统提示输入 URL 时不设默认值

#### Scenario: 添加资源时配置 SSH key
- **WHEN** 用户选择配置 SSH key 并输入 `~/.ssh/id_work`
- **THEN** 系统将该 SSH key 路径写入 resource 配置的 `ssh_key` 字段

#### Scenario: 添加资源后确认
- **WHEN** 用户完成所有输入
- **THEN** 系统显示输入摘要并请求确认，确认后写入配置文件
