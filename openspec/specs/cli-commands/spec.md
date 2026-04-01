### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `clone`: clone repositories
- `list`: list discovered repositories
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands and global flags

### Requirement: clone command
系统 SHALL 提供 `grepom clone` 命令，将仓库 clone 到本地文件系统。Group 内 repo 的目标路径通过路径推导公式计算。独立 repo 使用其 local_path。

#### Scenario: Clone all repos
- **WHEN** 用户运行 `grepom clone`（无参数）
- **THEN** 系统从所有 groups 和独立 repos clone 所有仓库到各自推导的本地路径

#### Scenario: Clone single repo by name
- **WHEN** 用户运行 `grepom clone web-app`
- **THEN** 系统仅 clone 名为 `web-app` 的仓库（在所有 groups 和独立 repos 中搜索）

#### Scenario: Clone by group
- **WHEN** 用户运行 `grepom clone --group frontend`
- **THEN** 系统仅 clone group `frontend` 下的所有仓库

#### Scenario: Repo already exists
- **WHEN** 目标目录已包含 git 仓库
- **THEN** 系统跳过 clone 并打印提示（非错误）

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

#### Scenario: Add group
- **WHEN** user runs `grepom add group --name frontend --resource work-gl --path my-org/frontend --local-path ./frontend --recursive`
- **THEN** the system appends a group entry to the config YAML file under `groups`

#### Scenario: Add repo to group
- **WHEN** user runs `grepom add repo --name special --resource work-gl --url https://gitlab.../special.git --group frontend --path my-org/frontend/special`
- **THEN** the system appends a repo entry to group `frontend`'s repos list

#### Scenario: Add independent repo
- **WHEN** user runs `grepom add repo --name dotfiles --resource github --url https://github.com/me/dotfiles.git`
- **THEN** the system appends a repo entry to the top-level `repos` list with default local_path `./dotfiles`

#### Scenario: Add to specific config file
- **WHEN** user runs `grepom -c /path/to/config.yml add resource ...`
- **THEN** the system appends to `/path/to/config.yml`
