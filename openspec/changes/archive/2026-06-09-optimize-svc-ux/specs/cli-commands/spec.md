## ADDED Requirements

### Requirement: completion command
The system SHALL provide a `grepom completion` subcommand as part of the root command set.

#### Scenario: completion listed in help
- **WHEN** the user runs `grepom --help`
- **THEN** the system SHALL list `completion` among available subcommands

#### Scenario: completion help in English
- **WHEN** the user runs `grepom completion --help`
- **THEN** the command description and examples SHALL be in English
