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
- `sync`: synchronize repository metadata and update configuration (does NOT clone or pull)

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands (including `sync`) and global flags

### Requirement: add command
The system SHALL provide a `grepom add` command with two subcommands:
- `grepom add source`: append a new API source to the config file
- `grepom add repo`: append a new explicit repo to the config file

`grepom add source` SHALL support an optional `--name` flag to assign a name identifier to the source.

#### Scenario: Add gitlab source with group and name
- **WHEN** user runs `grepom add source --name my-gitlab --provider gitlab --url https://gitlab.com --group my-org/frontend`
- **THEN** the system appends a source entry with `name: my-gitlab` to the config YAML file

#### Scenario: Add github source with org
- **WHEN** user runs `grepom add source --provider github --url https://github.com --org my-org`
- **THEN** the system appends a source entry (without name field) to the config YAML file

#### Scenario: Add repo with custom path
- **WHEN** user runs `grepom add repo --name special --url https://gitlab.com/other/special.git --path ./special`
- **THEN** the system appends a repo entry to the config YAML file

#### Scenario: Add to specific config file
- **WHEN** user runs `grepom -c /path/to/config.yml add source ...`
- **THEN** the system appends to `/path/to/config.yml`
