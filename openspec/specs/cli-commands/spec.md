### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `example`: export a complete example configuration with all features
- `clone`: clone repositories（使用 5 级认证优先级链，SSH 优先，输出认证尝试日志）
- `list`: list resources, groups, or repositories（通过 --type 标志切换，默认列出 repos）
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)
- `search`: search repositories by name (substring match)
- `interactive`: 进入交互式操作模式

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands including `example`, `search` and `interactive` and global flags

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

### Requirement: clone 命令兼容提示
当用户运行 `grepom init [name]`（带位置参数）时，系统 SHALL 提示用户 "did you mean `grepom clone`?" 并退出，帮助已有用户迁移。

#### Scenario: 用户误用 init clone 语法
- **WHEN** 用户运行 `grepom init web-app`（带位置参数）
- **THEN** 系统提示 "did you mean `grepom clone`?" 并以非零状态码退出

### Requirement: list command
系统 SHALL 提供 `grepom list` 命令，支持通过 `--type` 标志切换列出目标。`--type` 支持三个值：`repos`（默认）、`resources`、`groups`。

当 `--type` 为 `repos`（默认）时，行为与原有 list 命令一致：列出所有已发现的仓库，支持位置参数 `[name]` 按名称过滤、`--group` 按分组过滤、`--resource` 按 resource 过滤。

当 `--type` 为 `resources` 或 `groups` 时，位置参数和过滤标志不生效。

`list` 命令的位置参数 SHALL 支持关键字 `groups` 和 `resources`，当位置参数为 `groups` 时等价于 `--type groups`，当位置参数为 `resources` 时等价于 `--type resources`。位置参数关键字优先级低于 `--type` 标志（即 `grepom list groups --type repos` 以 `--type repos` 为准）。

`--remote` 标志 SHALL 支持 `--type groups`，通过 provider API 查询远程 groups/orgs 列表。`--remote` 不支持 `--type resources`。

`list` 命令 SHALL 支持 `--no-push` 标志筛选有未推送提交的仓库，以及 `--no-commit` 标志筛选有未提交更改的仓库。两个标志仅对本地仓库列表生效（`--type repos` 默认模式），在 `--remote` 模式下静默忽略。

list 命令的 flag SHALL 支持短别名：`-t`（`--type`）、`-r`（`--remote`）、`-g`（`--group`）、`-R`（`--resource`）。

#### Scenario: List all repos
- **WHEN** user runs `grepom list`
- **THEN** the system displays all repos from all groups and independent repos, with name, path, provider, and clone status

#### Scenario: List single repo
- **WHEN** user runs `grepom list web-app`
- **THEN** the system displays info for only `web-app`

#### Scenario: List by group
- **WHEN** user runs `grepom list --group frontend`
- **THEN** the system displays repos only from group `frontend`

#### Scenario: List by group using short flag
- **WHEN** user runs `grepom list -g frontend`
- **THEN** the system displays repos only from group `frontend`，行为与 `--group` 完全一致

#### Scenario: List by resource
- **WHEN** user runs `grepom list --resource work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`

#### Scenario: List by resource using short flag
- **WHEN** user runs `grepom list -R work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`，行为与 `--resource` 完全一致

#### Scenario: --type repos 等同默认行为
- **WHEN** user runs `grepom list --type repos`
- **THEN** 系统行为与 `grepom list` 完全一致，列出所有 repos

#### Scenario: --type repos 使用短别名
- **WHEN** user runs `grepom list -t repos`
- **THEN** 系统行为与 `grepom list --type repos` 完全一致

#### Scenario: --type 列出 resources
- **WHEN** user runs `grepom list --type resources`
- **THEN** 系统列出所有已配置的 resources，输出包含名称、provider、url 和 ssh_key

#### Scenario: --type 列出 groups
- **WHEN** user runs `grepom list --type groups`
- **THEN** 系统列出所有已配置的 groups，输出包含名称、关联 resource、路径、recursive 和 repo 数量

#### Scenario: 位置参数 groups 等价 --type groups
- **WHEN** user runs `grepom list groups`
- **THEN** 系统行为与 `grepom list --type groups` 完全一致，列出所有已配置的 groups

#### Scenario: 位置参数 resources 等价 --type resources
- **WHEN** user runs `grepom list resources`
- **THEN** 系统行为与 `grepom list --type resources` 完全一致，列出所有已配置的 resources

#### Scenario: --type 标志优先于位置参数关键字
- **WHEN** user runs `grepom list groups --type repos`
- **THEN** 系统以 `--type repos` 为准，列出所有 repos（`groups` 位置参数被忽略）

#### Scenario: --remote --type groups 远程列出 groups
- **WHEN** user runs `grepom list --remote --type groups`
- **THEN** 系统通过 provider API 查询远程 groups/orgs 列表并输出

#### Scenario: --remote --type groups 使用短别名
- **WHEN** user runs `grepom list -r -t groups`
- **THEN** 系统通过 provider API 查询远程 groups/orgs 列表并输出，行为与长 flag 版本完全一致

#### Scenario: --remote 不支持 --type resources
- **WHEN** user runs `grepom list --remote --type resources`
- **THEN** 系统输出错误信息 "--remote is not supported with --type resources"

#### Scenario: 使用混合短别名
- **WHEN** user runs `grepom list -r -t groups -R work-gl`
- **THEN** 系统仅查询 resource `work-gl` 的远程 groups，行为与 `grepom list --remote --type groups --resource work-gl` 完全一致

### Requirement: status command
系统 SHALL 提供 `grepom status` 命令，显示已克隆仓库的 git 状态概要和每个仓库的精简状态。

输出分为两部分：
1. **概要表格**：以表格形式统计各状态 repo 数量，表格包含 STATUS 和 COUNT 两列。状态行包括：clean、dirty、ahead、behind、not cloned（仅显示数量 > 0 的状态行）。表格下方显示总 repo 数。
2. **仓库列表**：每个 repo 一行，包含名称、状态标记、本地路径，三列对齐显示

状态标记优先级（仅显示最高优先级）：not cloned > dirty (N) > ahead N > behind N > clean

#### Scenario: Status of all repos with概要
- **WHEN** 用户运行 `grepom status`
- **THEN** 系统先输出概要表格统计各状态数量，然后列出每个 repo 的名称、状态标记和本地路径

#### Scenario: Status by group
- **WHEN** 用户运行 `grepom status --group frontend`
- **THEN** 系统仅显示 group `frontend` 下的 repo，概要表格也仅统计该 group 的 repo

#### Scenario: Status of not-yet-cloned repo
- **WHEN** 某 repo 未克隆
- **THEN** 系统在列表中显示该 repo，状态标记为 `not cloned`，不调用 git status

#### Scenario: 所有 repo 均为 clean
- **WHEN** 所有 repo 均 clean，无 ahead/behind
- **THEN** 概要表格仅显示 clean 行，无 dirty/ahead/behind 行

#### Scenario: 无仓库
- **WHEN** 过滤后无仓库匹配
- **THEN** 系统输出 `No repositories found.`

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

### Requirement: add command
The system SHALL provide a `grepom add` command with three subcommands:
- `grepom add resource`: append a new resource to the config file
- `grepom add group`: append a new group to the config file
- `grepom add repo`: append a new explicit repo to the config file

#### Scenario: Add resource
- **WHEN** user runs `grepom add resource --name work-gl --provider gitlab --url https://gitlab.mycompany.com --token ${GITLAB_TOKEN}`
- **THEN** the system appends a resource entry `work-gl` to the config YAML file under `resources`

#### Scenario: Add resource with SSH key
- **WHEN** user runs `grepom add resource --name work-gl --provider gitlab --url https://gitlab.com --token ${GL_TOKEN} --ssh-key ~/.ssh/id_work`
- **THEN** the system writes the SSH key path to the resource's `ssh_key` field

#### Scenario: Add resource without SSH key
- **WHEN** user runs `grepom add resource --name work-gl --provider gitlab --url https://gitlab.com --token ${GL_TOKEN}` (no --ssh-key)
- **THEN** the resource's `ssh_key` field is empty

#### Scenario: Add group
- **WHEN** user runs `grepom add group --name frontend --resource work-gl --path my-org/frontend --local-path ./frontend --recursive`
- **THEN** the system appends a group entry to the config YAML file under `groups`

#### Scenario: Add group with auth override
- **WHEN** user runs `grepom add group --name frontend --resource work-gl --path my-org/frontend --ssh-key ~/.ssh/deploy_fe --token ${FE_TOKEN}`
- **THEN** the system writes ssh_key and token to the group config

#### Scenario: Add repo to group
- **WHEN** user runs `grepom add repo --name special --resource work-gl --url https://gitlab.../special.git --group frontend --path my-org/frontend/special`
- **THEN** the system appends a repo entry to group `frontend`'s repos list

#### Scenario: Add independent repo
- **WHEN** user runs `grepom add repo --name dotfiles --resource github --url https://github.com/me/dotfiles.git`
- **THEN** the system appends a repo entry to the top-level `repos` list with default local_path `./dotfiles`

#### Scenario: Add independent repo with auth override
- **WHEN** user runs `grepom add repo --name dotfiles --resource github --url https://github.com/me/dotfiles.git --ssh-key ~/.ssh/id_personal`
- **THEN** the system writes ssh_key to the repo config

#### Scenario: Add to specific config file
- **WHEN** user runs `grepom -c /path/to/config.yml add resource ...`
- **THEN** the system appends to `/path/to/config.yml`

### Requirement: interactive 子命令
系统 SHALL 提供 `grepom interactive` 子命令，启动交互式引导操作模式。该命令不需要任何参数。

#### Scenario: 运行 interactive 命令
- **WHEN** 用户运行 `grepom interactive`
- **THEN** 系统进入交互式操作菜单

#### Scenario: interactive 命令与 config 标志兼容
- **WHEN** 用户运行 `grepom -c custom.yml interactive`
- **THEN** 交互模式使用 `custom.yml` 作为配置文件路径

### Requirement: search 命令
系统 SHALL 提供 `grepom search <keyword>` 命令，按名称模糊搜索仓库。搜索使用大小写不敏感的子串匹配。

#### Scenario: 搜索匹配的仓库
- **WHEN** 用户运行 `grepom search web`
- **THEN** 系统显示所有名称包含 "web"（大小写不敏感）的仓库，包括 group 内 repo 和 standalone repo

#### Scenario: 搜索无匹配结果
- **WHEN** 用户运行 `grepom search xyz`
- **THEN** 系统输出 "no matching repos found"

#### Scenario: 搜索关键字为空
- **WHEN** 用户运行 `grepom search`（无参数）
- **THEN** 系统报错提示需要提供搜索关键字

### Requirement: search 结合 group 过滤器
`search` 命令 SHALL 支持 `--group` 过滤器，仅在指定 group 的范围内搜索仓库。

#### Scenario: 在指定 group 内搜索
- **WHEN** 用户运行 `grepom search web --group frontend`
- **THEN** 系统仅在 group `frontend` 的 repos 中搜索名称包含 "web" 的仓库

#### Scenario: 指定的 group 不存在
- **WHEN** 用户运行 `grepom search web --group nonexistent`
- **THEN** 系统输出 "no matching repos found"

### Requirement: search 结合 resource 过滤器
`search` 命令 SHALL 支持 `--resource` 过滤器，仅在引用指定 resource 的仓库中搜索。

#### Scenario: 在指定 resource 范围内搜索
- **WHEN** 用户运行 `grepom search web --resource work-gl`
- **THEN** 系统仅在引用 resource `work-gl` 的仓库中搜索名称包含 "web" 的仓库

### Requirement: search 输出格式
`search` 命令的输出格式 SHALL 与 `list` 命令保持一致，以表格形式显示仓库名称、路径、group、resource 和克隆状态。

#### Scenario: search 输出包含完整信息
- **WHEN** 用户运行 `grepom search web` 且找到匹配仓库
- **THEN** 输出包含仓库名称、本地路径、所属 group、关联 resource 和克隆状态，格式与 `grepom list` 一致
