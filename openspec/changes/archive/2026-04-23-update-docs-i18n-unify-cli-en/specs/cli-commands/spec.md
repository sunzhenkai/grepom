## MODIFIED Requirements

### Requirement: Root command with global flags
The system SHALL provide a root command `grepom` with the following global flags:
- `-c, --config`: path to the YAML configuration file (optional, default: `.grepom.yml`)
- `--verbose`: enable verbose output

The system SHALL also provide the following subcommands:
- `init`: initialize configuration file
- `example`: export a complete example configuration with all features
- `clone`: clone repositories (uses 5-level auth priority chain, SSH first, outputs auth attempt logs)
- `list`: list resources, groups, or repositories (switch via --type flag, defaults to repos)
- `status`: show git status
- `pull`: pull updates
- `add`: add resource, group, or repository
- `sync`: synchronize repository metadata from groups (does NOT clone or pull)
- `search`: search repositories by name (substring match)
- `interactive`: enter interactive operation mode

All cobra command descriptions (Short, Long, Example) and flag help text SHALL be in English.

#### Scenario: Show help
- **WHEN** user runs `grepom --help`
- **THEN** the system displays available commands including `example`, `search` and `interactive` and global flags, all in English

#### Scenario: Mixed language output no longer occurs
- **WHEN** user runs any grepom command that produces output
- **THEN** all user-facing output (descriptions, help text, progress, errors) SHALL be in English only

### Requirement: clone command
The system SHALL provide a `grepom clone` command that clones repositories to the local filesystem. The target path for repos within a group is calculated via path derivation formula. Standalone repos use their local_path.

Clone auth priority chain (5 levels, SSH first): group/repo SSH key → group/repo token → resource SSH key → derived SSH → resource token.

During clone, the system SHALL output logs for each auth method attempt. In parallel mode, auth logs SHALL be collected into results rather than output directly, and displayed grouped by repository after completion.

The clone command SHALL support `--concurrency` parameter (default 4) to control the number of parallel clone workers. When `--concurrency` is 1, sequential behavior is preserved.

All auth strategy labels and progress output SHALL be in English.

#### Scenario: Clone all repos
- **WHEN** user runs `grepom clone` (no arguments)
- **THEN** the system clones all repos from all groups and standalone repos in parallel to their derived local paths, attempts auth via priority chain, and outputs an operation summary in English

#### Scenario: Auth attempt log output
- **WHEN** an auth method is attempted during clone (sequential mode)
- **THEN** the system outputs log `  [N/M] trying <method> (<level>)...`; on failure outputs error summary; on success outputs "ok"

#### Scenario: Parallel mode auth log collection
- **WHEN** during parallel clone (`--concurrency > 1`) an auth method is attempted
- **THEN** the system collects auth attempt logs into results and displays them grouped by repository after completion

### Requirement: list command
The system SHALL provide a `grepom list` command that supports switching list target via `--type` flag. `--type` supports three values: `repos` (default), `resources`, `groups`.

When output contains informational messages (e.g., skipping a group with no bound resource), those messages SHALL be in English.

#### Scenario: List with unbound resource warning
- **WHEN** user runs `grepom list` and a group has no bound resource
- **THEN** the system outputs an English message indicating the group was skipped due to no bound resource
