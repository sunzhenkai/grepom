# shell-completion Specification

## Purpose
TBD - created by archiving change optimize-svc-ux. Update Purpose after archive.
## Requirements
### Requirement: Shell completion command
The system SHALL provide a `grepom completion` subcommand that generates shell completion scripts for `bash`, `zsh`, and `fish`.

#### Scenario: Generate bash completion
- **WHEN** the user runs `grepom completion bash`
- **THEN** the system SHALL print a bash completion script to stdout suitable for `eval "$(grepom completion bash)"`

#### Scenario: Generate zsh completion
- **WHEN** the user runs `grepom completion zsh`
- **THEN** the system SHALL print a zsh completion script to stdout suitable for `eval "$(grepom completion zsh)"`

#### Scenario: Generate fish completion
- **WHEN** the user runs `grepom completion fish`
- **THEN** the system SHALL print a fish completion script to stdout

#### Scenario: Reject unknown shell
- **WHEN** the user runs `grepom completion` with an unsupported shell argument
- **THEN** the system SHALL return a non-zero exit code and an error message

### Requirement: Root command completion coverage
The generated completion script SHALL provide completion for the `grepom` root command, its global flags, and all registered subcommands and their flags.

#### Scenario: Complete subcommands
- **WHEN** the user types `grepom <TAB>` in a shell with completion enabled
- **THEN** the shell SHALL offer available top-level subcommands including `svc` and `completion`

#### Scenario: Complete global flags
- **WHEN** the user types `grepom -<TAB>` in a shell with completion enabled
- **THEN** the shell SHALL offer global flags such as `--config` and `--verbose`

### Requirement: Service name completion for svc commands
The system SHALL provide dynamic shell completion for service name arguments on `grepom svc` and `grepom service` subcommands.

#### Scenario: Complete service names for logs
- **WHEN** the user types `grepom svc logs <TAB>` and completion is enabled
- **THEN** the system SHALL offer service names from the current registry scope and configured `services` keys in `.grepom.yml`, merged and deduplicated

#### Scenario: Complete service names for kill
- **WHEN** the user types `grepom svc kill <TAB>` and completion is enabled
- **THEN** the system SHALL offer the same merged service name set as for `logs`

#### Scenario: Complete service names for dir
- **WHEN** the user types `grepom svc dir <TAB>` and completion is enabled
- **THEN** the system SHALL offer the same merged service name set

#### Scenario: Complete service names for status
- **WHEN** the user types `grepom svc status <TAB>` and completion is enabled
- **THEN** the system SHALL offer the same merged service name set

#### Scenario: Complete configured names for run
- **WHEN** the user types `grepom svc run <TAB>` and completion is enabled
- **THEN** the system SHALL offer configured service names from `.grepom.yml` and names from the current registry scope

#### Scenario: Respect config scope for completion
- **WHEN** the user runs completion with `grepom -c /path/to/other.yml svc kill <TAB>`
- **THEN** the system SHALL resolve service names for the registry scope of that config file

#### Scenario: Completion fails silently
- **WHEN** service name completion cannot resolve a manager or registry
- **THEN** the system SHALL return no completions without printing errors to the terminal

