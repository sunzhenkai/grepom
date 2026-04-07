## MODIFIED Requirements

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
