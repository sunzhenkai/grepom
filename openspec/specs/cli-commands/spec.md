### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `example`: export a complete example configuration with all features
- `clone`: clone repositories (uses 5-level auth priority chain, SSH first, outputs auth attempt logs)
- `list`: list resources, groups, or repositories (switch via --type flag, defaults to repos)
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)
- `search`: search repositories by name (substring match)
- `interactive`: enter interactive operation mode

All cobra command descriptions (Short, Long, Example) and flag help text SHALL be in English.

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands including `example`, `search` and `interactive` and global flags, all in English

#### Scenario: Mixed language output no longer occurs
- **WHEN** user runs any grepom command that produces output
- **THEN** all user-facing output (descriptions, help text, progress, errors) SHALL be in English only

### Requirement: clone command
The system SHALL provide a `grepom clone` command that clones repositories to the local filesystem. The target path for repos within a group is calculated via path derivation formula. Standalone repos use their local_path.

Clone auth priority chain (5 levels, SSH first): group/repo SSH key → group/repo token → resource SSH key → derived SSH → resource token.

During clone, the system SHALL output logs for each auth method attempt. In parallel mode, auth logs SHALL be collected into results rather than output directly, and displayed grouped by repository after completion.

The clone command SHALL support `--concurrency` parameter (default 4) to control the number of parallel clone workers. When `--concurrency` is 1, sequential behavior is preserved.

All auth strategy labels and progress output SHALL be in English.

#### Scenario: Clone all repos
- **WHEN** user runs `grepom clone` (no arguments)
- **THEN** the system clones all repos from all groups and standalone repos in parallel to their derived local paths, attempts auth via priority chain, and outputs an operation summary in English

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

#### Scenario: Auth attempt log output
- **WHEN** an auth method is attempted during clone (sequential mode)
- **THEN** the system outputs log `  [N/M] trying <method> (<level>)...`; on failure outputs error summary; on success outputs "ok"

#### Scenario: Parallel mode auth log collection
- **WHEN** during parallel clone (`--concurrency > 1`) an auth method is attempted
- **THEN** the system collects auth attempt logs into results and displays them grouped by repository after completion

#### Scenario: 指定并行度
- **WHEN** 用户运行 `grepom clone --concurrency 8`
- **THEN** 系统使用 8 个 worker 并发克隆仓库

### Requirement: clone 命令兼容提示
当用户运行 `grepom init [name]`（带位置参数）时，系统 SHALL 提示用户 "did you mean `grepom clone`?" 并退出，帮助已有用户迁移。

#### Scenario: 用户误用 init clone 语法
- **WHEN** 用户运行 `grepom init web-app`（带位置参数）
- **THEN** 系统提示 "did you mean `grepom clone`?" 并以非零状态码退出

### Requirement: list command
The system SHALL provide a `grepom list` command that supports switching list target via `--type` flag. `--type` supports three values: `repos` (default), `resources`, `groups`.

When `--type` is `repos` (default), behavior matches original list command: list all discovered repos, supports positional `[name]` filter, `--group` filter, `--resource` filter.

When `--type` is `resources` or `groups`, positional args and filter flags are ignored.

The `list` command positional arg SHALL support keywords `groups` and `resources`. When positional arg is `groups` it is equivalent to `--type groups`; when `resources` equivalent to `--type resources`. Positional keyword has lower priority than `--type` flag (i.e., `grepom list groups --type repos` uses `--type repos`).

The `--remote` flag SHALL support `--type groups`, querying remote groups/orgs list via provider API. `--remote` does not support `--type resources`.

The `list` command SHALL support `--no-push` flag to filter repos with unpushed commits, and `--no-commit` flag to filter repos with uncommitted changes. Both flags only apply to local repo listing (`--type repos` default mode), silently ignored in `--remote` mode.

list command flags SHALL support short aliases: `-t` (`--type`), `-r` (`--remote`), `-g` (`--group`), `-R` (`--resource`).

When output contains informational messages (e.g., skipping a group with no bound resource), those messages SHALL be in English.

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
- **THEN** the system displays repos only from group `frontend`, behavior identical to `--group`

#### Scenario: List by resource
- **WHEN** user runs `grepom list --resource work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`

#### Scenario: List by resource using short flag
- **WHEN** user runs `grepom list -R work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`, behavior identical to `--resource`

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

#### Scenario: List with unbound resource warning
- **WHEN** user runs `grepom list` and a group has no bound resource
- **THEN** the system outputs an English message indicating the group was skipped due to no bound resource

#### Scenario: 使用混合短别名
- **WHEN** user runs `grepom list -r -t groups -R work-gl`
- **THEN** the system queries only resource `work-gl` remote groups, behavior identical to `grepom list --remote --type groups --resource work-gl`

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
