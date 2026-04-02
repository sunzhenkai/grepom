### Requirement: list resources 命令
系统 SHALL 提供 `grepom list --type resources` 命令，列出所有已配置的资源。输出以表格形式显示 NAME、PROVIDER、URL、SSH_KEY 列。Token 字段不在输出中显示原始值。

#### Scenario: 列出所有资源
- **WHEN** 用户运行 `grepom list --type resources`
- **THEN** 系统以表格形式输出所有已配置的 resource，每行包含名称、provider（gitlab/github）、url 和 ssh_key（未配置显示 `-`）

#### Scenario: 配置中无 resource
- **WHEN** 用户运行 `grepom list --type resources` 且配置文件中 `resources` 为空
- **THEN** 系统输出 `No resources found.`

#### Scenario: 资源有 SSH key 配置
- **WHEN** 配置中某 resource 配置了 `ssh_key: ~/.ssh/id_work`
- **THEN** 表格中 SSH_KEY 列显示 `~/.ssh/id_work`

#### Scenario: 资源无 SSH key 配置
- **WHEN** 配置中某 resource 未配置 `ssh_key` 字段
- **THEN** 表格中 SSH_KEY 列显示 `-`

#### Scenario: --type resources 忽略位置参数
- **WHEN** 用户运行 `grepom list --type resources some-name`
- **THEN** 系统忽略位置参数 `some-name`，仍列出所有 resources

#### Scenario: --type resources 忽略过滤标志
- **WHEN** 用户运行 `grepom list --type resources --group frontend --resource work-gl`
- **THEN** 系统忽略 `--group` 和 `--resource` 标志，仍列出所有 resources
