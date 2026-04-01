## MODIFIED Requirements

### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `clone`: clone repositories（使用 6 级认证优先级链，SSH 优先，输出认证尝试日志）
- `list`: list discovered repositories
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)
- `search`: search repositories by name (substring match)
- `interactive`: 进入交互式操作模式

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands including `search` and `interactive` and global flags

### Requirement: clone command
系统 SHALL 提供 `grepom clone` 命令，将仓库 clone 到本地文件系统。Group 内 repo 的目标路径通过路径推导公式计算。独立 repo 使用其 local_path。

clone 认证优先级链（6 级，SSH 优先）：group/repo SSH key → group/repo token → resource SSH key → resource token → 推导 SSH → 裸 HTTP。

clone 过程中 SHALL 输出每种认证方式的尝试日志。

#### Scenario: Clone all repos
- **WHEN** 用户运行 `grepom clone`（无参数）
- **THEN** 系统从所有 groups 和独立 repos clone 所有仓库到各自推导的本地路径，按 SSH 优先优先级链尝试认证

#### Scenario: Clone single repo by name
- **WHEN** 用户运行 `grepom clone web-app`
- **THEN** 系统仅 clone 名为 `web-app` 的仓库，按优先级链尝试认证

#### Scenario: Clone by group
- **WHEN** 用户运行 `grepom clone --group frontend`
- **THEN** 系统仅 clone group `frontend` 下的所有仓库

#### Scenario: Repo already exists
- **WHEN** 目标目录已包含 git 仓库
- **THEN** 系统跳过 clone 并打印提示（非错误）

#### Scenario: 使用 group 级别 SSH key 克隆（最高优先级）
- **WHEN** group 配置了 ssh_key
- **THEN** 系统优先使用 group 的 SSH key 进行 SSH clone

#### Scenario: 认证尝试日志输出
- **WHEN** clone 过程中尝试某种认证方式
- **THEN** 系统输出日志 `  [N/M] 尝试 <方式> (<级别>)...`；失败时输出错误摘要；成功时输出 "成功"
