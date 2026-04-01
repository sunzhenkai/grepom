## MODIFIED Requirements

### Requirement: clone command
系统 SHALL 提供 `grepom clone` 命令，将仓库 clone 到本地文件系统。Group 内 repo 的目标路径通过路径推导公式计算。独立 repo 使用其 local_path。

clone 认证优先级链（6 级）：group/repo token → group/repo SSH key → resource token → resource SSH key → 推导 SSH → 裸 HTTP。

clone 过程中 SHALL 输出每种认证方式的尝试日志。

#### Scenario: Clone all repos
- **WHEN** 用户运行 `grepom clone`（无参数）
- **THEN** 系统从所有 groups 和独立 repos clone 所有仓库到各自推导的本地路径，按优先级链尝试认证

#### Scenario: Clone single repo by name
- **WHEN** 用户运行 `grepom clone web-app`
- **THEN** 系统仅 clone 名为 `web-app` 的仓库，按优先级链尝试认证

#### Scenario: Clone by group
- **WHEN** 用户运行 `grepom clone --group frontend`
- **THEN** 系统仅 clone group `frontend` 下的所有仓库

#### Scenario: Repo already exists
- **WHEN** 目标目录已包含 git 仓库
- **THEN** 系统跳过 clone 并打印提示（非错误）

#### Scenario: 使用 group 级别 token 克隆
- **WHEN** group 配置了 token
- **THEN** 系统优先使用 group token 构建 HTTPS 认证 URL 进行 clone

#### Scenario: 使用 resource SSH key 作为回退
- **WHEN** resource token 认证失败，且 resource 配置了 ssh_key
- **THEN** 系统使用 resource 的 SSH key 尝试 SSH clone

#### Scenario: 认证尝试日志输出
- **WHEN** clone 过程中尝试某种认证方式
- **THEN** 系统输出日志 `  [N/M] 尝试 <方式> (<级别>)...`；失败时输出错误摘要；成功时输出 "成功"

### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `clone`: clone repositories（使用 6 级认证优先级链，输出认证尝试日志）
- `list`: list discovered repositories
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)
- `interactive`: 进入交互式操作模式

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands including `interactive` and global flags

### Requirement: add resource command
`grepom add resource` 命令 SHALL 支持新增 `--ssh-key` flag，用于配置 resource 级别的 SSH 密钥文件路径。

#### Scenario: 添加资源时指定 SSH key
- **WHEN** 用户运行 `grepom add resource --name work-gl --provider gitlab --url https://gitlab.com --token ${GL_TOKEN} --ssh-key ~/.ssh/id_work`
- **THEN** 系统将 SSH key 路径写入 resource 的 `ssh_key` 字段

#### Scenario: 添加资源时不指定 SSH key
- **WHEN** 用户运行 `grepom add resource --name work-gl --provider gitlab --url https://gitlab.com --token ${GL_TOKEN}`（不含 --ssh-key）
- **THEN** resource 的 `ssh_key` 字段为空

### Requirement: add group command
`grepom add group` 命令 SHALL 支持新增 `--ssh-key` 和 `--token` flag，用于配置 group 级别的认证覆盖。

#### Scenario: 添加组时指定认证
- **WHEN** 用户运行 `grepom add group --name frontend --resource work-gl --path my-org/frontend --ssh-key ~/.ssh/deploy_fe --token ${FE_TOKEN}`
- **THEN** 系统将 ssh_key 和 token 写入 group 配置

### Requirement: add repo command
`grepom add repo` 命令 SHALL 支持新增 `--ssh-key` 和 `--token` flag，用于配置独立 repo 级别的认证覆盖。

#### Scenario: 添加独立 repo 时指定认证
- **WHEN** 用户运行 `grepom add repo --name dotfiles --resource github --url https://github.com/me/dotfiles.git --ssh-key ~/.ssh/id_personal`
- **THEN** 系统将 ssh_key 写入 repo 配置

## ADDED Requirements

### Requirement: interactive 子命令
系统 SHALL 提供 `grepom interactive` 子命令，启动交互式引导操作模式。该命令不需要任何参数。

#### Scenario: 运行 interactive 命令
- **WHEN** 用户运行 `grepom interactive`
- **THEN** 系统进入交互式操作菜单

#### Scenario: interactive 命令与 config 标志兼容
- **WHEN** 用户运行 `grepom -c custom.yml interactive`
- **THEN** 交互模式使用 `custom.yml` 作为配置文件路径
