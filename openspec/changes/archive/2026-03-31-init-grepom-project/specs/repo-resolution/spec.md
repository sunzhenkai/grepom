## ADDED Requirements

### Requirement: Aggregate repos from all sources
The system SHALL collect repos from all configured sources (API sources and explicit repo entries) into a unified list. If a repo appears in multiple sources, the first occurrence SHALL win (no deduplication required for MVP).

#### Scenario: Multiple sources with repos
- **WHEN** config has a GitLab source with 3 repos and a GitHub source with 2 repos
- **THEN** the unified list contains 5 repos

#### Scenario: Explicit repo alongside API sources
- **WHEN** config has an API source and a `repos` section with explicit entries
- **THEN** the unified list contains repos from both sources

### Requirement: Map group/subgroup hierarchy to directory path
The system SHALL map the group/subgroup hierarchy from the API into a relative directory path. The path SHALL be relative to the `base` directory.

#### Scenario: GitLab nested group path
- **WHEN** GitLab API returns a project in group `my-org` > subgroup `frontend` > subgroup `components`
- **THEN** the relative path is `my-org/frontend/components/<project-name>`

#### Scenario: GitHub org path
- **WHEN** GitHub API returns a repo in org `my-org`
- **THEN** the relative path is `my-org/<repo-name>`

#### Scenario: Custom path override for explicit repos
- **WHEN** an explicit repo entry has `path: ./special`
- **THEN** the relative path is `special` (resolved relative to base)

### Requirement: Filter repos by target
The system SHALL support filtering the repo list by:
- Exact repo name
- Group/org name (all repos under that group/org)
- Provider name

#### Scenario: Filter by repo name
- **WHEN** user specifies repo name `web-app`
- **THEN** only repos matching `web-app` are returned

#### Scenario: Filter by group
- **WHEN** user specifies `--group my-org/frontend`
- **THEN** all repos under `my-org/frontend` (including subgroups if recursive) are returned

#### Scenario: Filter by provider
- **WHEN** user specifies `--source gitlab`
- **THEN** only repos from GitLab sources are returned

#### Scenario: No filter (all repos)
- **WHEN** user specifies no filter
- **THEN** all repos from all sources are returned
