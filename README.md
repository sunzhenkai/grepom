# grepom

Git Repository Orchestrator & Manager — manage multiple git repositories across GitLab groups and GitHub organizations from a single YAML config.

## Features

- **Declarative config** — define GitLab groups and GitHub orgs in YAML, grepom discovers repos automatically
- **Bulk operations** — clone, pull, and check status across all repos at once
- **Hierarchical layout** — preserves group/subgroup directory structure locally
- **Multi-provider** — works with both GitLab and GitHub APIs
- **Flexible filtering** — filter by name, group, or provider

## Install

```bash
go install github.com/wii/grepom@latest
```

Or build from source:

```bash
make install
```

## Usage

Create a config file (default: `.grepom.yml`):

```yaml
base: ~/projects

sources:
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_TOKEN}
    groups:
      - path: my-org/frontend
        recursive: true

  - name: my-github
    provider: github
    url: https://github.com
    token: ${GITHUB_TOKEN}
    orgs:
      - name: my-org
```

### Commands

```bash
grepom sync                            # Discover repos and update config metadata
grepom sync --source my-gitlab         # Sync a specific source by name
grepom sync --source 0                 # Sync a specific source by index

grepom clone                           # Clone all discovered repos
grepom clone web-app                   # Clone a specific repo
grepom clone --group my-org/frontend   # Clone all repos in a group

grepom list                            # List all discovered repos
grepom list --source gitlab            # Filter by provider
grepom list --group my-org/frontend    # Filter by group

grepom status                          # Check status of all cloned repos
grepom status web-app                  # Status of a specific repo

grepom pull                            # Pull updates for all cloned repos
grepom pull web-app                    # Pull a specific repo
```

### Add sources/repos interactively

```bash
grepom add source --name my-gitlab --provider gitlab --url https://gitlab.com --group my-org/backend --recursive --token '${GITLAB_TOKEN}'
grepom add source --provider github --url https://github.com --org my-org
grepom add repo --name special --url https://gitlab.com/other/special.git
```

### Token environment variables

Token fields support `${ENV_VAR}` placeholder syntax. The actual value is resolved from the environment at runtime, and the placeholder is preserved when writing config files.

```yaml
sources:
  - provider: gitlab
    token: ${GITLAB_TOKEN}   # Resolved from $GITLAB_TOKEN at runtime
```

```bash
export GITLAB_TOKEN=glpat-xxxxxxxxxxxx
grepom sync   # uses the resolved token value
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `.grepom.yml` | Path to config file |
| `--verbose` | `-v` | `false` | Enable verbose output |

## Build

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run vet and format check
make install  # Build and install to ~/.local/bin
make clean    # Remove binary
```
