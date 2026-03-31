## MODIFIED Requirements

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
- `add`: add source or repository
- `sync`: synchronize repositories and update configuration

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands (including `sync`) and global flags
