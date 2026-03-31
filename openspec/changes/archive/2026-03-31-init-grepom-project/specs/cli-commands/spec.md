## ADDED Requirements

### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands and global flags

### Requirement: init command
The system SHALL provide a `grepom init` command that clones repositories to the local filesystem. Repositories SHALL be cloned to `<base>/<path>`, creating directories as needed.

#### Scenario: Clone all repos
- **WHEN** user runs `grepom init` with no arguments
- **THEN** the system clones all repos from all sources to their respective paths under `base`

#### Scenario: Clone single repo
- **WHEN** user runs `grepom init web-app`
- **THEN** the system clones only the repo named `web-app`

#### Scenario: Clone by group
- **WHEN** user runs `grepom init --group my-org/frontend`
- **THEN** the system clones all repos under `my-org/frontend`

#### Scenario: Repo already exists
- **WHEN** the target directory already contains a git repository
- **THEN** the system skips cloning and prints a message (no error)

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
