# service-process-management Specification

## Purpose

在 `.grepom.yml` 中定义本地开发服务，并通过 `grepom svc` 后台启动、查看状态、管理日志与进程。
## Requirements
### Requirement: Service Definitions
The system SHALL support optional service definitions in `.grepom.yml` under a top-level `services` field.

#### Scenario: Load configured service
- **WHEN** the config contains a service with `cwd` and `command`
- **THEN** `grepom svc run <name>` SHALL resolve the service directory and command from the config

#### Scenario: Resolve relative service directory
- **WHEN** a configured service uses a relative `cwd`
- **THEN** the system SHALL resolve it relative to the directory containing `.grepom.yml`

#### Scenario: Support string and array commands
- **WHEN** a configured service command is a string or a string array
- **THEN** the system SHALL accept both forms and preserve the intended command execution semantics

### Requirement: Service Startup
The system SHALL allow users to start a service from an explicit command or from a configured service definition.

#### Scenario: Start service from command line
- **WHEN** the user runs `grepom svc run -- <command> [args...]`
- **THEN** the system SHALL start the command in the current directory using the current directory name as the default service name

#### Scenario: Start named service from command line
- **WHEN** the user runs `grepom svc run <name> -- <command> [args...]`
- **THEN** the system SHALL start the command using `<name>` as the service name

#### Scenario: Start service from config
- **WHEN** the user runs `grepom svc run <name>` and `<name>` exists in the service config
- **THEN** the system SHALL start the configured command in the configured service directory

#### Scenario: Record service runtime metadata
- **WHEN** a service starts successfully
- **THEN** the system SHALL record the service name, PID, process group when available, working directory, command, log path, start time, and config scope in the service registry

#### Scenario: Prevent duplicate running service
- **WHEN** a service with the same name is already running in the same registry scope
- **THEN** the system SHALL refuse to start another service with that name unless an explicit replacement option is provided

### Requirement: Service Registry Storage
The system SHALL store runtime service metadata outside `.grepom.yml` in a machine-local state location.

#### Scenario: Keep runtime data out of config
- **WHEN** a service is started, stopped, or cleaned
- **THEN** `.grepom.yml` SHALL NOT be modified with PID, status, log path, or other runtime-only fields

#### Scenario: Isolate registry by config or directory
- **WHEN** services are managed from different config files or standalone directories
- **THEN** the system SHALL keep their registry records isolated to avoid same-name collisions across scopes

### Requirement: Service List and Status
The system SHALL list managed services in a table and include their current runtime status.

#### Scenario: List services as table
- **WHEN** the user runs `grepom svc list`
- **THEN** the system SHALL print a table containing at least service name, status, PID, path, command, and log path

#### Scenario: Show service path in list
- **WHEN** a service appears in `grepom svc list`
- **THEN** the service working directory SHALL be displayed in the table

#### Scenario: Refresh process status during list
- **WHEN** the user runs `grepom svc list` or `grepom svc status`
- **THEN** the system SHALL check the recorded process and display whether the service is running, exited, or stale

#### Scenario: Show single service status
- **WHEN** the user runs `grepom svc status <name>`
- **THEN** the system SHALL display the selected service status and relevant metadata

### Requirement: Service Logs
The system SHALL capture and expose service logs.

#### Scenario: Write stdout and stderr to log
- **WHEN** a service is started
- **THEN** the system SHALL append the service stdout and stderr to the recorded log file

#### Scenario: View recent logs
- **WHEN** the user runs `grepom svc logs <name>`
- **THEN** the system SHALL print recent log lines for the selected service

#### Scenario: Select log line count
- **WHEN** the user runs `grepom svc logs -n <count> <name>`
- **THEN** the system SHALL print at most the requested number of trailing log lines

#### Scenario: Follow logs
- **WHEN** the user runs `grepom svc logs -f <name>`
- **THEN** the system SHALL continue printing new log lines until interrupted

#### Scenario: Open logs in editor
- **WHEN** the user runs `grepom svc logs --open <name>`
- **THEN** the system SHALL open the service log with the configured editor or platform opener, or print the log path if no opener is available

### Requirement: Service Termination
The system SHALL support graceful and forceful service termination.

#### Scenario: Gracefully kill service
- **WHEN** the user runs `grepom svc kill <name>`
- **THEN** the system SHALL send a graceful termination signal to the service process group when available, otherwise to the recorded PID

#### Scenario: Force kill service
- **WHEN** the user runs `grepom svc kill -9 <name>`
- **THEN** the system SHALL send a forceful termination signal to the service process group when available, otherwise to the recorded PID

#### Scenario: Report missing process
- **WHEN** the recorded process for a service no longer exists
- **THEN** the system SHALL report that the service is not running and mark or leave it eligible for cleanup

### Requirement: Service Cleanup
The system SHALL clean records for services that are no longer running.

#### Scenario: Clean exited services
- **WHEN** the user runs `grepom svc clean`
- **THEN** the system SHALL remove registry records for exited or stale services and preserve their log files by default

#### Scenario: Clean logs explicitly
- **WHEN** the user runs cleanup with an explicit log deletion option
- **THEN** the system SHALL remove logs associated with cleaned service records

#### Scenario: Keep running services during cleanup
- **WHEN** cleanup encounters a running service
- **THEN** the system SHALL keep the service record unchanged

### Requirement: Service Directory Output
The system SHALL print a service working directory for shell integration.

#### Scenario: Print configured service directory
- **WHEN** the user runs `grepom svc dir <name>`
- **THEN** the system SHALL print the service working directory to stdout

#### Scenario: Print shell helper
- **WHEN** the user runs `grepom svc --shell`
- **THEN** the system SHALL print a shell helper that can cd to selected service directories

### Requirement: Service TUI
The system SHALL provide a terminal UI for managing services.

#### Scenario: Open service TUI
- **WHEN** the user runs `grepom svc tui`
- **THEN** the system SHALL open an interactive terminal interface listing managed services

#### Scenario: Display status and path in TUI
- **WHEN** the TUI lists services
- **THEN** it SHALL show each service name, status, PID, path, command, and log path or equivalent detail view

#### Scenario: Manage selected service in TUI
- **WHEN** the user selects a service in the TUI
- **THEN** the TUI SHALL allow the user to view logs, refresh status, stop the service, force stop the service, and show or copy the service path

#### Scenario: Clean exited services in TUI
- **WHEN** the user invokes cleanup from the TUI
- **THEN** the TUI SHALL clean exited or stale service records using the same behavior as `grepom svc clean`

#### Scenario: Reuse service manager behavior
- **WHEN** the TUI performs list, status, logs, kill, clean, or dir actions
- **THEN** it SHALL use the same underlying service management behavior as the CLI commands

### Requirement: Service Command Aliases and Help
The system SHALL expose discoverable command names and help for service management.

#### Scenario: Use service alias
- **WHEN** the user runs `grepom service <subcommand>`
- **THEN** the system SHALL behave the same as `grepom svc <subcommand>`

#### Scenario: Show service help
- **WHEN** the user runs `grepom svc --help`
- **THEN** the system SHALL describe available service management commands, configured-service usage, direct command usage, list table fields, logs, cleanup, and TUI entry

