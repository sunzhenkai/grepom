## ADDED Requirements

### Requirement: Provider interface
The system SHALL define a `Provider` interface with a `ListRepos(ctx context.Context, source Source) ([]Repo, error)` method. Each provider implementation (GitLab, GitHub) SHALL implement this interface.

#### Scenario: Select provider by config
- **WHEN** config source has `provider: gitlab`
- **THEN** the system uses the GitLab provider implementation

#### Scenario: Unknown provider
- **WHEN** config source has `provider: bitbucket`
- **THEN** the system returns an error indicating the unsupported provider

### Requirement: GitLab API - list group repositories
The system SHALL call the GitLab API v4 endpoint `GET /groups/:id/projects` to list repositories under a group. It SHALL support pagination to retrieve all results.

#### Scenario: Group with multiple pages of projects
- **WHEN** a GitLab group has more than 20 projects (default page size)
- **THEN** the system fetches all pages and returns the complete list

#### Scenario: Empty group
- **WHEN** a GitLab group has no projects
- **THEN** the system returns an empty list without error

#### Scenario: Invalid token
- **WHEN** the GitLab API returns 401 Unauthorized
- **THEN** the system returns a clear authentication error message

### Requirement: GitLab API - recursive subgroup traversal
When a group config has `recursive: true`, the system SHALL recursively fetch all subgroups and their projects using `GET /groups/:id/subgroups` and `GET /groups/:id/projects`.

#### Scenario: Group with nested subgroups
- **WHEN** group `my-org/frontend` has subgroup `components` with projects `ui-lib` and `icons`
- **THEN** the system returns repos with paths preserving the nested structure:
  - `my-org/frontend/web-app` (direct project)
  - `my-org/frontend/components/ui-lib`
  - `my-org/frontend/components/icons`

#### Scenario: Recursive disabled
- **WHEN** group config has `recursive: false` (or the field is absent)
- **THEN** the system only fetches direct projects of the group, not subgroups

### Requirement: GitHub API - list organization repositories
The system SHALL call the GitHub REST API endpoint `GET /orgs/:org/repos` to list repositories under an organization. It SHALL support pagination.

#### Scenario: Organization with multiple pages of repos
- **WHEN** a GitHub org has more than 30 repos (default page size)
- **THEN** the system fetches all pages and returns the complete list

#### Scenario: Invalid token
- **WHEN** the GitHub API returns 401 Unauthorized
- **THEN** the system returns a clear authentication error message

### Requirement: Repo model
Each resolved repo SHALL contain:
- `Name`: repository name
- `CloneURL`: HTTP clone URL
- `SSHURL`: SSH clone URL
- `Path`: relative path from base directory (reflecting group/subgroup hierarchy)
- `Provider`: source provider name

#### Scenario: GitLab repo path structure
- **WHEN** a repo `web-app` is under group `my-org/frontend`
- **THEN** the repo's `Path` is `my-org/frontend/web-app`

#### Scenario: GitHub repo path structure
- **WHEN** a repo `api-server` is under org `my-org`
- **THEN** the repo's `Path` is `my-org/api-server`

### Requirement: Respect API rate limits
The system SHALL respect rate limit headers returned by GitLab and GitHub APIs. When rate limited, the system SHALL wait and retry.

#### Scenario: GitLab rate limit
- **WHEN** GitLab API returns 429 Too Many Requests with `Retry-After` header
- **THEN** the system waits for the specified duration and retries the request

#### Scenario: GitHub rate limit
- **WHEN** GitHub API returns 403 with rate limit headers
- **THEN** the system prints a message about the rate limit and when it resets
