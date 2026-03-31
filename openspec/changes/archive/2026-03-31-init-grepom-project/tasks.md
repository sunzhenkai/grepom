## 1. Project Skeleton

- [x] 1.1 Initialize Go module (`go mod init`) and create `main.go` entry point
- [x] 1.2 Set up cobra root command in `cmd/root.go` with `-c/--config` and `--verbose` global flags
- [x] 1.3 Create directory structure: `cmd/`, `config/`, `provider/`, `repo/`, `git/`
- [x] 1.4 Add dependencies: `github.com/spf13/cobra`, `gopkg.in/yaml.v3`

## 2. Config Management

- [x] 2.1 Define config Go structs in `config/config.go`: Config, Source, GroupSource, OrgSource, RepoEntry
- [x] 2.2 Implement YAML parsing with `os.ExpandEnv` for `${VAR}` token resolution
- [x] 2.3 Implement tilde `~` expansion for the `base` field
- [x] 2.4 Implement config file discovery: `-c` flag > `.grepom.yml` in current directory
- [x] 2.5 Implement config file writing (append source/repo entries while preserving existing content)

## 3. Provider Interface & Repo Model

- [x] 3.1 Define `Provider` interface and `Repo` struct in `provider/provider.go`
- [x] 3.2 Implement provider registry/factory to select provider by config name
- [x] 3.3 Define path resolution logic: group/subgroup hierarchy → relative directory path

## 4. GitLab Provider

- [x] 4.1 Implement GitLab HTTP client with auth header (Private-Token)
- [x] 4.2 Implement `GET /groups/:id/projects` with pagination
- [x] 4.3 Implement `GET /groups/:id/subgroups` for subgroup discovery
- [x] 4.4 Implement recursive subgroup traversal (BFS/DFS) when `recursive: true`
- [x] 4.5 Implement rate limit handling (429 / Retry-After)
- [x] 4.6 Map GitLab API response to `Repo` struct with path hierarchy

## 5. GitHub Provider

- [x] 5.1 Implement GitHub HTTP client with auth header (Authorization: Bearer)
- [x] 5.2 Implement `GET /orgs/:org/repos` with pagination (Link header)
- [x] 5.3 Implement rate limit handling (403 + X-RateLimit-Reset)
- [x] 5.4 Map GitHub API response to `Repo` struct with path hierarchy

## 6. Repo Resolution & Aggregation

- [x] 6.1 Implement repo list aggregation from all sources (API sources + explicit repos)
- [x] 6.2 Implement filtering: by name, by group, by provider
- [x] 6.3 Implement path resolution with base directory joining

## 7. Git Operations

- [x] 7.1 Implement `git clone` wrapper with directory creation and SSH/HTTP URL fallback
- [x] 7.2 Implement `git pull` wrapper with error handling
- [x] 7.3 Implement `git status --porcelain=v2 --branch` output parsing
- [x] 7.4 Implement clone check (directory + `.git` existence)

## 8. CLI Commands

- [x] 8.1 Implement `list` command: fetch repos from providers, display table
- [x] 8.2 Implement `init` command: clone repos (all/single/group), skip existing
- [x] 8.3 Implement `status` command: show git status for cloned repos
- [x] 8.4 Implement `pull` command: git pull on cloned repos, skip uncloned
- [x] 8.5 Implement `add source` subcommand: append source entry to YAML
- [x] 8.6 Implement `add repo` subcommand: append repo entry to YAML

## 9. Polish & Validation

- [x] 9.1 Add `--help` documentation for all commands and flags
- [x] 9.2 Add `--verbose` logging throughout
- [x] 9.3 Write unit tests for config parsing, path resolution, git status parsing
- [x] 9.4 Write integration tests for provider API calls (with mocked HTTP)
- [x] 9.5 Verify end-to-end flow: config → fetch repos → init → status → pull
