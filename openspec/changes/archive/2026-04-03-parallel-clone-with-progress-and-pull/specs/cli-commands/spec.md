## MODIFIED Requirements

### Requirement: clone command
系统 SHALL 提供 `grepom clone` 命令，将仓库 clone 到本地文件系统。Group 内 repo 的目标路径通过路径推导公式计算。独立 repo 使用其 local_path。

clone 认证优先级链（5 级，SSH 优先）：group/repo SSH key → group/repo token → resource SSH key → 推导 SSH → resource token。

clone 过程中 SHALL 输出每种认证方式的尝试日志。并行模式下，认证日志 SHALL 被收集到结果中而非直接输出，完成后按仓库分组展示。

clone 命令 SHALL 支持 `--concurrency` 参数（默认 4）控制并行克隆的仓库数量。当 `--concurrency` 为 1 时保持原有顺序行为。

#### Scenario: Clone all repos
- **WHEN** 用户运行 `grepom clone`（无参数）
- **THEN** 系统从所有 groups 和独立 repos 并行克隆所有仓库到各自推导的本地路径，按优先级链尝试认证，完成后输出操作摘要

#### Scenario: Clone single repo by name
- **WHEN** 用户运行 `grepom clone web-app`
- **THEN** 系统仅 clone 名为 `web-app` 的仓库，按优先级链尝试认证

#### Scenario: Clone by group
- **WHEN** 用户运行 `grepom clone --group frontend`
- **THEN** 系统仅 clone group `frontend` 下的所有仓库

#### Scenario: Repo already exists
- **WHEN** 目标目录已包含 git 仓库
- **THEN** 系统跳过 clone 并在 verbose 模式下打印提示

#### Scenario: 使用 group 级别 SSH key 克隆（最高优先级）
- **WHEN** group 配置了 ssh_key
- **THEN** 系统优先使用 group 的 SSH key 进行 SSH clone

#### Scenario: 认证尝试日志输出
- **WHEN** clone 过程中尝试某种认证方式（顺序模式）
- **THEN** 系统输出日志 `  [N/M] 尝试 <方式> (<级别>)...`；失败时输出错误摘要；成功时输出 "成功"

#### Scenario: 并行模式下认证日志收集
- **WHEN** 并行克隆（`--concurrency > 1`）过程中尝试某种认证方式
- **THEN** 系统将认证尝试日志收集到结果中，完成后按仓库分组展示

#### Scenario: 指定并行度
- **WHEN** 用户运行 `grepom clone --concurrency 8`
- **THEN** 系统使用 8 个 worker 并发克隆仓库

### Requirement: pull command
The system SHALL provide a `grepom pull` command that runs `git pull` on cloned repositories.

pull 命令 SHALL 默认启用安全检查：仅对"已克隆 + 在默认分支 + clean"的仓库执行 pull。使用 `--force` 标志可跳过安全检查，恢复无条件 pull 行为。

pull 命令 SHALL 支持 `--concurrency` 参数（默认 4）控制并行 pull 的仓库数量。

#### Scenario: Pull all cloned repos
- **WHEN** user runs `grepom pull`
- **THEN** the system runs `git pull` on each cloned repo that is on its default branch and has a clean working tree, across all groups and independent repos

#### Scenario: Pull by group
- **WHEN** user runs `grepom pull --group frontend`
- **THEN** the system runs `git pull` only on repos in group `frontend` that satisfy safety checks

#### Scenario: Pull on not-yet-cloned repo
- **WHEN** user runs `grepom pull` and a repo has not been cloned
- **THEN** the system skips that repo and shows "not cloned"

#### Scenario: Pull on dirty repo
- **WHEN** user runs `grepom pull` and a repo has uncommitted changes
- **THEN** the system skips that repo and shows "dirty working tree"

#### Scenario: Pull on non-default branch
- **WHEN** user runs `grepom pull` and a repo is on a feature branch
- **THEN** the system skips that repo and shows the current branch name

#### Scenario: Pull with local changes using --force
- **WHEN** user runs `grepom pull --force` and a repo has local changes
- **THEN** the system runs `git pull` on that repo; if it fails, shows the error

#### Scenario: Pull with --concurrency
- **WHEN** user runs `grepom pull --concurrency 4`
- **THEN** the system uses 4 workers to pull eligible repos in parallel

#### Scenario: Pull on single repo
- **WHEN** user runs `grepom pull web-app`
- **THEN** the system only pulls `web-app` (with safety checks applied)
