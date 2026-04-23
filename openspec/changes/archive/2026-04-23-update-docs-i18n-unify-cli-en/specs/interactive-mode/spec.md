## MODIFIED Requirements

### Requirement: interactive 交互式命令入口
系统 SHALL 提供 `grepom interactive` 命令，启动交互式操作模式，显示操作菜单供用户选择。所有交互式界面的用户可见文本（菜单选项、提示信息、错误信息、确认信息、进度输出）SHALL 使用英文。

#### Scenario: 启动交互模式
- **WHEN** 用户运行 `grepom interactive`
- **THEN** 系统显示英文操作菜单，包含以下选项：Initialize config、Add resource、Add group、Add repo、Sync remote repos、Clone repos、Pull updates、Check status、Exit

#### Scenario: 选择操作后执行
- **WHEN** 用户从菜单中选择 "Initialize config"
- **THEN** 系统进入 init 交互引导流程，所有提示信息为英文

#### Scenario: 退出交互模式
- **WHEN** 用户从菜单中选择 "Exit"
- **THEN** 系统输出 "Bye!" 并正常退出

#### Scenario: 非 TTY 环境运行
- **WHEN** 用户在非交互式终端中运行 `grepom interactive`
- **THEN** 系统输出英文错误信息 "interactive mode requires a TTY" 并以非零状态码退出

### Requirement: 交互式初始化配置
交互模式 SHALL 引导用户完成配置文件初始化。所有提示信息 SHALL 使用英文。

#### Scenario: 完整初始化流程
- **WHEN** 用户在交互模式中选择 "Initialize config"
- **THEN** 系统依次以英文提示：Config file path（默认 `.grepom.yml`）、Base directory（默认 `~/projects`）、是否添加资源、Provider type（选择 gitlab/github/generic）、API URL、token、是否配置 SSH key（可选）

### Requirement: 交互式添加资源
交互模式 SHALL 引导用户添加新的 resource 配置。所有提示信息和确认信息 SHALL 使用英文。

#### Scenario: 添加资源流程
- **WHEN** 用户在交互模式中选择 "Add resource"
- **THEN** 系统依次以英文提示：Resource name、Provider type（从 gitlab/github/generic 中选择）、API URL、Token（提示支持 `${ENV_VAR}` 语法）、是否配置 SSH key（可选）

#### Scenario: 无已配置资源时提示
- **WHEN** 用户选择 "Add group" 但尚未配置任何资源
- **THEN** 系统输出英文提示 "No resources configured. Please add a resource first." 并返回主菜单

### Requirement: 交互式同步和克隆
交互模式 SHALL 支持执行 sync 和 clone 操作。范围选择选项和进度输出 SHALL 使用英文。

#### Scenario: 交互式同步范围选择
- **WHEN** 用户选择 "Sync remote repos"
- **THEN** 系统提示选择同步范围（All / By group / By resource），选项均为英文

### Requirement: 交互式拉取更新
交互模式 SHALL 支持执行 pull 操作。进度输出 SHALL 使用英文。

#### Scenario: 交互式拉取
- **WHEN** 用户选择 "Pull updates"
- **THEN** 系统提示选择拉取范围（All / By group / By resource），选项均为英文，跳过和进度信息为英文

### Requirement: 交互式查看状态
交互模式 SHALL 支持查看仓库状态。状态输出 SHALL 使用英文。

#### Scenario: 交互式状态查看
- **WHEN** 用户选择 "Check status"
- **THEN** 系统提示选择查看范围（All / By group / By resource），状态信息为英文（如 "not cloned"、"not a git repo"）
