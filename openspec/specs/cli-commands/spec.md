### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands and global flags

### Requirement: clone command
系统 SHALL 提供 `grepom clone` 命令，将仓库 clone 到本地文件系统 `<base>/<path>`，按需创建目录。

#### Scenario: Clone all repos
- **WHEN** 用户运行 `grepom clone`（无参数）
- **THEN** 系统从所有 sources clone 所有仓库到各自 base 下的路径

#### Scenario: Clone single repo
- **WHEN** 用户运行 `grepom clone web-app`
- **THEN** 系统仅 clone 名为 `web-app` 的仓库

#### Scenario: Clone by group
- **WHEN** 用户运行 `grepom clone --group my-org/frontend`
- **THEN** 系统仅 clone `my-org/frontend` 下的所有仓库

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
- **THEN** the system displays all repos with name, path, provider, and clone status

#### Scenario: List single repo
- **WHEN** user runs `grepom list web-app`
- **THEN** the system displays info for only `web-app`

#### Scenario: List with filters
- **WHEN** user runs `grepom list --source gitlab --group my-org/frontend`
- **THEN** the system displays repos matching both filters

### Requirement: status command
The system SHALL provide a `grepom status` command that shows git status for cloned repositories.

#### Scenario: Status of all cloned repos
- **WHEN** user runs `grepom status`
- **THEN** the system shows git status (branch, clean/dirty, ahead/behind) for each cloned repo

#### Scenario: Status of single repo
- **WHEN** user runs `grepom status web-app`
- **THEN** the system shows git status for `web-app` only

#### Scenario: Status of not-yet-cloned repo
- **WHEN** user runs `grepom status` and a repo has not been cloned yet
- **THEN** the system shows "not cloned" for that repo

### Requirement: pull command
The system SHALL provide a `grepom pull` command that runs `git pull` on cloned repositories.

#### Scenario: Pull all cloned repos
- **WHEN** user runs `grepom pull`
- **THEN** the system runs `git pull` on each cloned repo

#### Scenario: Pull single repo
- **WHEN** user runs `grepom pull web-app`
- **THEN** the system runs `git pull` only on `web-app`

#### Scenario: Pull on not-yet-cloned repo
- **WHEN** user runs `grepom pull` and a repo has not been cloned
- **THEN** the system skips that repo and shows "not cloned"

#### Scenario: Pull with local changes
- **WHEN** `git pull` fails due to local changes
- **THEN** the system shows the error and continues with the next repo

### Requirement: add command
The system SHALL provide a `grepom add` command with two subcommands:
- `grepom add source`: append a new API source to the config file
- `grepom add repo`: append a new explicit repo to the config file

#### Scenario: Add gitlab source with group
- **WHEN** user runs `grepom add source --provider gitlab --url https://gitlab.com --group my-org/frontend`
- **THEN** the system appends a source entry to the config YAML file

#### Scenario: Add github source with org
- **WHEN** user runs `grepom add source --provider github --url https://github.com --org my-org`
- **THEN** the system appends a source entry to the config YAML file

#### Scenario: Add repo with custom path
- **WHEN** user runs `grepom add repo --name special --url https://gitlab.com/other/special.git --path ./special`
- **THEN** the system appends a repo entry to the config YAML file

#### Scenario: Add to specific config file
- **WHEN** user runs `grepom -c /path/to/config.yml add source ...`
- **THEN** the system appends to `/path/to/config.yml`
