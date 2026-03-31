## ADDED Requirements

### Requirement: Parse YAML configuration file
The system SHALL parse YAML configuration files with the following structure:
- `base`: string, the root directory for all cloned repositories (support `~` expansion)
- `sources`: array of source definitions
- `repos`: array of explicit repository definitions (optional, for repos not discoverable via API)

Each source SHALL have:
- `provider`: string, one of `gitlab` or `github`
- `url`: string, the API base URL (e.g., `https://gitlab.com`)
- `token`: string, the API token (supports `${ENV_VAR}` syntax)
- `groups`: array of group definitions (GitLab-specific)
- `orgs`: array of organization definitions (GitHub-specific)

#### Scenario: Valid configuration file with GitLab source
- **WHEN** user provides a YAML file with a gitlab source containing groups
- **THEN** the system parses it into the internal config structure with all fields populated

#### Scenario: Configuration file with environment variable token
- **WHEN** YAML contains `token: ${GITLAB_TOKEN}` and env var `GITLAB_TOKEN=glpat-xxx`
- **THEN** the system resolves the token to `glpat-xxx`

#### Scenario: Missing required field
- **WHEN** a source entry is missing the `provider` field
- **THEN** the system returns a clear validation error indicating the missing field

### Requirement: Support multiple configuration files
The system SHALL accept a configuration file path via the `-c` / `--config` global flag. When not provided, the system SHALL look for `.grepom.yml` in the current directory.

#### Scenario: Explicit config file via flag
- **WHEN** user runs `grepom -c ~/work-repos.yml list`
- **THEN** the system loads configuration from `~/work-repos.yml`

#### Scenario: Default config file
- **WHEN** user runs `grepom list` and `.grepom.yml` exists in current directory
- **THEN** the system loads configuration from `.grepom.yml`

#### Scenario: No config file found
- **WHEN** user runs `grepom list` with no `-c` flag and no `.grepom.yml` in current directory
- **THEN** the system prints an error message suggesting how to specify a config file

### Requirement: Expand tilde in base path
The system SHALL expand `~` to the user's home directory in the `base` field.

#### Scenario: Base path with tilde
- **WHEN** config contains `base: ~/projects`
- **THEN** the system resolves `base` to `/home/<user>/projects`

### Requirement: Expand environment variables in token fields
The system SHALL expand `${VAR_NAME}` syntax in any string field to the corresponding environment variable value.

#### Scenario: Multiple env vars in config
- **WHEN** config contains `${GITLAB_TOKEN}` and `${GITHUB_TOKEN}` in different source entries
- **THEN** each is resolved to its respective environment variable value

#### Scenario: Undefined environment variable
- **WHEN** config references `${UNDEFINED_VAR}` and the env var does not exist
- **THEN** the system returns an error indicating the undefined variable

### Requirement: Write to configuration file
The system SHALL support appending new entries to an existing YAML configuration file while preserving formatting and comments.

#### Scenario: Add a new source
- **WHEN** user runs `grepom add source --provider gitlab --url https://gitlab.com --group my-org`
- **THEN** the system appends a new source entry to the YAML file

#### Scenario: Add a new explicit repo
- **WHEN** user runs `grepom add repo --name special --url https://gitlab.com/other/special.git --path ./special`
- **THEN** the system appends a new repo entry to the YAML file

#### Scenario: Config file does not exist yet
- **WHEN** user runs `grepom add source ...` and the target config file does not exist
- **THEN** the system creates the file with a minimal valid YAML structure and the new entry
