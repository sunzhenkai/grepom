## ADDED Requirements

### Requirement: interactive 交互式命令入口
系统 SHALL 提供 `grepom interactive` 命令，启动交互式操作模式，显示操作菜单供用户选择。

#### Scenario: 启动交互模式
- **WHEN** 用户运行 `grepom interactive`
- **THEN** 系统显示操作菜单，包含以下选项：初始化配置、添加资源、添加组、添加仓库、同步远程仓库、克隆仓库、查看状态、退出

#### Scenario: 选择操作后执行
- **WHEN** 用户从菜单中选择"初始化配置"
- **THEN** 系统进入 init 交互引导流程

#### Scenario: 退出交互模式
- **WHEN** 用户从菜单中选择"退出"
- **THEN** 系统正常退出，不执行任何操作

#### Scenario: 非 TTY 环境运行
- **WHEN** 用户在非交互式终端（如 CI/CD 管道）中运行 `grepom interactive`
- **THEN** 系统提示"interactive 模式需要交互式终端"并以非零状态码退出

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

### Requirement: 交互式添加组
交互模式 SHALL 引导用户添加新的 group 配置，包括可选的认证信息。

#### Scenario: 添加组流程
- **WHEN** 用户在交互模式中选择"添加组"
- **THEN** 系统依次提示：组名称、关联资源（从已配置的资源列表中选择）、远程路径、本地路径（默认 `./<name>`）、是否递归（GitLab）、是否配置独立 SSH key（可选）、是否配置独立 token（可选）

#### Scenario: 添加组时配置 SSH key
- **WHEN** 用户选择配置独立 SSH key 并输入 `~/.ssh/deploy_frontend`
- **THEN** 系统将该 SSH key 路径写入 group 配置的 `ssh_key` 字段

#### Scenario: 添加组时配置 token
- **WHEN** 用户选择配置独立 token 并输入 `${FRONTEND_TOKEN}`
- **THEN** 系统将该 token 写入 group 配置的 `token` 字段

#### Scenario: 无已配置资源时提示
- **WHEN** 用户选择"添加组"但尚未配置任何资源
- **THEN** 系统提示"请先添加资源"并返回主菜单

### Requirement: 交互式添加仓库
交互模式 SHALL 引导用户添加独立仓库或组内仓库，包括可选的认证信息。

#### Scenario: 添加独立仓库
- **WHEN** 用户选择"添加仓库"并选择"独立仓库"
- **THEN** 系统依次提示：仓库名称、关联资源、clone URL、本地路径（默认 `./<name>`）、是否配置独立 SSH key（可选）、是否配置独立 token（可选）

#### Scenario: 添加组内仓库
- **WHEN** 用户选择"添加仓库"并选择"添加到组"
- **THEN** 系统提示选择目标组（从已配置组列表中选择），然后依次提示：仓库名称、clone URL、远程路径（组内仓库继承组的认证配置，不单独提示认证）

#### Scenario: 添加独立仓库时配置认证
- **WHEN** 用户在添加独立仓库时选择配置 SSH key 和 token
- **THEN** 系统将认证信息写入 repo 配置的 `ssh_key` 和 `token` 字段

### Requirement: 交互式同步和克隆
交互模式 SHALL 支持执行 sync 和 clone 操作。

#### Scenario: 交互式同步
- **WHEN** 用户选择"同步远程仓库"
- **THEN** 系统提示选择同步范围（全部/按组/按资源），确认后执行同步

#### Scenario: 交互式克隆
- **WHEN** 用户选择"克隆仓库"
- **THEN** 系统提示选择克隆范围（全部/按组/按资源），确认后执行克隆（使用认证优先级链和认证尝试日志）

### Requirement: 交互式查看状态
交互模式 SHALL 支持查看仓库状态。

#### Scenario: 交互式状态查看
- **WHEN** 用户选择"查看状态"
- **THEN** 系统提示选择查看范围（全部/按组/按资源），然后显示仓库状态
