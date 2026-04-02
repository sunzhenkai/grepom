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

#### Scenario: 使用 group 级别 SSH key 克隆（最高优先级）
- **WHEN** group 配置了 ssh_key
- **THEN** 系统优先使用 group 的 SSH key 进行 SSH clone

#### Scenario: 认证尝试日志输出
- **WHEN** clone 过程中尝试某种认证方式
- **THEN** 系统输出日志 `  [N/M] 尝试 <方式> (<级别>)...`；失败时输出错误摘要；成功时输出 "成功"

### Requirement: clone 命令兼容提示
当用户运行 `grepom init [name]`（带位置参数）时，系统 SHALL 提示用户 "did you mean `grepom clone`?" 并退出，帮助已有用户迁移。

#### Scenario: 用户误用 init clone 语法
- **WHEN** 用户运行 `grepom init web-app`（带位置参数）
- **THEN** 系统提示 "did you mean `grepom clone`?" 并以非零状态码退出

### Requirement: list command
The system SHALL provide a `grepom list` command that displays discovered repositories.

#### Scenario: List all repos
- **WHEN** user runs `grepom list`
- **THEN** the system displays all repos from all groups and independent repos, with name, path, provider, and clone status

#### Scenario: List single repo
- **WHEN** user runs `grepom list web-app`
- **THEN** the system displays info for only `web-app`

#### Scenario: List by group
- **WHEN** user runs `grepom list --group frontend`
- **THEN** the system displays repos only from group `frontend`

#### Scenario: List by resource
- **WHEN** user runs `grepom list --resource work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`

### Requirement: status command
The system SHALL provide a `grepom status` command that shows git status for cloned repositories.

#### Scenario: Status of all cloned repos
- **WHEN** user runs `grepom status`
- **THEN** the system shows git status for each cloned repo across all groups and independent repos

#### Scenario: Status by group
- **WHEN** user runs `grepom status --group frontend`
- **THEN** the system shows git status only for repos in group `frontend`

### Requirement: pull command
The system SHALL provide a `grepom pull` command that runs `git pull` on cloned repositories.

#### Scenario: Pull all cloned repos
- **WHEN** user runs `grepom pull`
- **THEN** the system runs `git pull` on each cloned repo across all groups and independent repos

#### Scenario: Pull by group
- **WHEN** user runs `grepom pull --group frontend`
- **THEN** the system runs `git pull` only on repos in group `frontend`

#### Scenario: Pull on not-yet-cloned repo
- **WHEN** user runs `grepom pull` and a repo has not been cloned
- **THEN** the system skips that repo and shows "not cloned"

#### Scenario: Pull with local changes
- **WHEN** `git pull` fails due to local changes
- **THEN** the system shows the error and continues with the next repo

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
