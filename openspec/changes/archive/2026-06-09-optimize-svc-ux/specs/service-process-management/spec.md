## MODIFIED Requirements

### Requirement: Service Registry Storage
The system SHALL store runtime service metadata outside `.grepom.yml` in a machine-local state location under the XDG state directory.

#### Scenario: Keep runtime data out of config
- **WHEN** a service is started, stopped, or cleaned
- **THEN** `.grepom.yml` SHALL NOT be modified with PID, status, log path, or other runtime-only fields

#### Scenario: Isolate registry by config or directory
- **WHEN** services are managed from different config files or standalone directories
- **THEN** the system SHALL keep their registry records isolated to avoid same-name collisions across scopes

#### Scenario: Store state under XDG state home
- **WHEN** the system resolves the service state directory
- **THEN** it SHALL use `$XDG_STATE_HOME/grepom/services/<scope>/` when `XDG_STATE_HOME` is set
- **AND** it SHALL use `~/.local/state/grepom/services/<scope>/` when `XDG_STATE_HOME` is not set

#### Scenario: No migration from legacy state directory
- **WHEN** a user upgrades after this change
- **THEN** the system SHALL NOT read or migrate registry data from the previous `UserConfigDir`-based location

### Requirement: Service List and Status
The system SHALL list managed services in a compact table by default and provide verbose output on request.

#### Scenario: List services as compact table
- **WHEN** the user runs `grepom svc list`
- **THEN** the system SHALL print a table containing service name, status, PID, and working directory path

#### Scenario: List services with verbose output
- **WHEN** the user runs `grepom svc list -v` or `grepom svc list --verbose`
- **THEN** the system SHALL print a table containing service name, status, PID, path, command, and log path

#### Scenario: Show service path in list
- **WHEN** a service appears in `grepom svc list`
- **THEN** the service working directory SHALL be displayed in the table

#### Scenario: Shorten home directory in displayed paths
- **WHEN** a service working directory is under the user home directory
- **THEN** the list output SHALL display the path with a `~` prefix instead of the full home path when formatting for display

#### Scenario: Refresh process status during list
- **WHEN** the user runs `grepom svc list` or `grepom svc status`
- **THEN** the system SHALL check the recorded process and display whether the service is running, exited, or stale

#### Scenario: Show single service status with full metadata
- **WHEN** the user runs `grepom svc status <name>`
- **THEN** the system SHALL display the selected service status and full metadata including command and log path
