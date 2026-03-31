## ADDED Requirements

### Requirement: Clone repository
The system SHALL clone a repository using `git clone`. It SHALL create the parent directory structure as needed and clone into the final directory.

#### Scenario: Clone to nested directory
- **WHEN** target path is `my-org/frontend/web-app` and only `my-org/` exists
- **THEN** the system creates `my-org/frontend/` directory and clones into `web-app/`

#### Scenario: Clone with SSH URL
- **WHEN** SSH URL is available and preferred
- **THEN** the system uses the SSH URL for cloning

#### Scenario: Clone with HTTP URL
- **WHEN** only HTTP URL is available or SSH fails
- **THEN** the system falls back to HTTP URL for cloning

#### Scenario: Clone failure
- **WHEN** `git clone` fails (e.g., network error, auth failure)
- **THEN** the system returns the error with context (repo name, URL)

### Requirement: Pull repository
The system SHALL run `git pull` in an existing cloned repository directory.

#### Scenario: Successful pull
- **WHEN** `git pull` succeeds with updates
- **THEN** the system returns success

#### Scenario: Already up to date
- **WHEN** `git pull` reports "Already up to date"
- **THEN** the system returns success without treating it as an error

#### Scenario: Pull failure
- **WHEN** `git pull` fails (e.g., merge conflict, detached HEAD)
- **THEN** the system returns the error with the repo name for context

### Requirement: Get repository status
The system SHALL run `git status --porcelain=v2 --branch` in a cloned repository directory and parse the output to determine:
- Current branch name
- Whether the working tree is clean or dirty
- Ahead/behind count relative to upstream

#### Scenario: Clean repository on main branch
- **WHEN** repo is clean and on `main` with no remote changes
- **THEN** the system reports branch `main`, clean, up to date

#### Scenario: Dirty repository with uncommitted changes
- **WHEN** repo has modified files
- **THEN** the system reports the branch name and indicates dirty state with file count

#### Scenario: Ahead of upstream
- **WHEN** repo has 2 commits ahead of origin/main
- **THEN** the system reports branch `main`, ahead 2

#### Scenario: Not a git repository
- **WHEN** the target directory exists but is not a git repository
- **THEN** the system returns an error indicating it is not a git repo

### Requirement: Check if repository is cloned
The system SHALL determine whether a repository has been cloned by checking if the target path contains a `.git` directory.

#### Scenario: Repository is cloned
- **WHEN** path `base/my-org/frontend/web-app/.git` exists
- **THEN** the system reports the repo as cloned

#### Scenario: Repository is not cloned
- **WHEN** path `base/my-org/frontend/web-app/` does not exist or has no `.git`
- **THEN** the system reports the repo as not cloned
