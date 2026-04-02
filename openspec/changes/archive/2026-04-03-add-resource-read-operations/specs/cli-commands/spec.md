## MODIFIED Requirements

### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `clone`: clone repositories（使用 5 级认证优先级链，SSH 优先，输出认证尝试日志）
- `list`: list resources, groups, or repositories（通过 `--type` 标志切换，默认列出 repos）
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)
- `search`: search repositories by name (substring match)
- `interactive`: 进入交互式操作模式

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands including `search` and `interactive` and global flags

## MODIFIED Requirements

### Requirement: list command
系统 SHALL 提供 `grepom list` 命令，支持通过 `--type` 标志切换列出目标。`--type` 支持三个值：`repos`（默认）、`resources`、`groups`。

当 `--type` 为 `repos`（默认）时，行为与原有 list 命令一致：列出所有已发现的仓库，支持位置参数 `[name]` 按名称过滤、`--group` 按分组过滤、`--resource` 按 resource 过滤。

当 `--type` 为 `resources` 或 `groups` 时，位置参数和过滤标志不生效。

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

#### Scenario: --type repos 等同默认行为
- **WHEN** user runs `grepom list --type repos`
- **THEN** 系统行为与 `grepom list` 完全一致，列出所有 repos

#### Scenario: --type 列出 resources
- **WHEN** user runs `grepom list --type resources`
- **THEN** 系统列出所有已配置的 resources，输出包含名称、provider、url 和 ssh_key

#### Scenario: --type 列出 groups
- **WHEN** user runs `grepom list --type groups`
- **THEN** 系统列出所有已配置的 groups，输出包含名称、关联 resource、路径、recursive 和 repo 数量
