# grepom

English | [简体中文](./README.md)

Git Repository Orchestrator & Manager — manage multiple git repositories across GitLab groups and GitHub organizations from a single YAML config.

## Features

- **Declarative config** — define GitLab groups and GitHub orgs in YAML, grepom discovers repos automatically
- **Bulk operations** — clone, pull, and check status across all repos at once
- **Hierarchical layout** — preserves group/subgroup directory structure locally
- **Multi-provider** — works with GitLab, GitHub, Codeup, and Generic APIs
- **Flexible filtering** — filter by name, group, virtual group, or provider
- **Virtual groups** — organize multiple real groups into named sets for batch operations via `--vgroup`
- **Secret scanning** — built-in gitleaks engine with workspace and git history scanning
- **Push guard** — automatically detect secrets before pushing
- **Interactive mode** — menu-driven interactive UI
- **MR/PR creation** — create GitHub Pull Requests or GitLab Merge Requests from the CLI; returns existing MR/PR address if one is already open
- **Service process management** — start local dev services in the background, inspect status/logs, stop processes, and manage them via TUI

## Install

### One-line install (recommended)

Download a prebuilt binary from GitHub Releases:

```bash
curl -fsSL https://raw.githubusercontent.com/sunzhenkai/grepom/master/scripts/install.sh | bash
```

By default this installs to `~/.local/bin`. To install to a system directory:

```bash
curl -fsSL https://raw.githubusercontent.com/sunzhenkai/grepom/master/scripts/install.sh | sudo INSTALL_DIR=/usr/local/bin bash
```

Install a specific version (including pre-releases):

```bash
VERSION=v0.2.0-rc.1 curl -fsSL https://raw.githubusercontent.com/sunzhenkai/grepom/master/scripts/install.sh | bash
```

> `latest` installs only the newest stable release, not `-rc` or `-beta` pre-releases; set `VERSION` explicitly for pre-releases.

### Install via Go

```bash
go install github.com/wii/grepom@latest
```

### Build from source

```bash
make install
```

## Quick Start

```bash
grepom init                     # Initialize config file
grepom example -o .grepom.yml   # Export example config (with all field descriptions)
grepom add resource ...         # Add an auth resource
grepom add group ...            # Add a remote group
grepom sync                     # Discover repos and update config
grepom clone                    # Clone all repos
```

## Usage

Create a config file (default: `.grepom.yml`). grepom automatically searches parent directories for the config file (similar to how git finds `.git`), so you can run commands from any subdirectory.

```yaml
base: ~/projects

resources:
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_TOKEN}
    ssh_key: ~/.ssh/id_work        # optional

  - name: my-github
    provider: github
    url: https://github.com
    token: ${GITHUB_TOKEN}

groups:
  - name: frontend
    resource: my-gitlab
    path: my-org/frontend
    recursive: true
    exclude_repos:                 # optional: exclude specific repos
      - archived-repo

  - name: my-org
    resource: my-github
    path: my-github-org

virtual_groups:                    # optional: named collections of real groups
  work:
    groups:
      - frontend
      - my-org

repos:                             # standalone repos (not part of any group)
  - name: dotfiles
    resource: my-github
    url: https://github.com/me/dotfiles.git

services:                          # optional local development service definitions
  api:
    cwd: ./backend
    command: make dev
  web:
    cwd: ./frontend
    command:
      - pnpm
      - dev
```

### Commands

```bash
# Init & Config
grepom init                         # Initialize config file
grepom example                      # Export complete example config
grepom interactive                  # Start interactive mode

# Sync & Discovery
grepom sync                         # Discover repos and update config metadata
grepom sync --source my-gitlab      # Sync a specific resource by name
grepom sync --group frontend        # Sync a specific group
grepom sync --vgroup work           # Sync all real groups in a virtual group

# Clone & Pull
grepom clone                        # Clone all discovered repos
grepom clone web-app                # Clone a specific repo
grepom clone --group frontend       # Clone all repos in a group
grepom clone --vgroup work          # Clone all repos in a virtual group
grepom clone --concurrency 8        # Clone with 8 parallel workers

grepom pull                         # Pull updates for all cloned repos
grepom pull web-app                 # Pull a specific repo
grepom pull --vgroup work           # Pull all repos in a virtual group
grepom pull --force                 # Skip safety checks and force pull
grepom pull --concurrency 8         # Pull with 8 parallel workers

# Query & Filter
grepom list                         # List repos needing attention (unpushed/uncommitted)
grepom list --all                   # List all repos with status
grepom list --no-push               # Only show repos with unpushed commits
grepom list --no-commit             # Only show repos with uncommitted changes
grepom list --group frontend        # Filter by real group
grepom list --vgroup work           # Filter by virtual group (expands to real groups)
grepom list --group infra --vgroup work  # --group and --vgroup are unioned
grepom list --resource my-gitlab    # Filter by resource (intersects with group selection)
grepom list groups                  # List configured real groups and virtual groups
grepom list resources               # List configured resources
grepom list --remote                # List remote repos from provider API
grepom list --remote --vgroup work  # Query remote repos for all real groups in a virtual group
grepom list --remote --type groups  # List remote groups from provider API

grepom status                       # Check status of all cloned repos
grepom status web-app               # Status of a specific repo
grepom status --vgroup work         # Status for all real groups in a virtual group

grepom search web                   # Search repos by name (substring match)
grepom search web --group frontend  # Search within a specific group
grepom search web --vgroup work     # Search within all real groups in a virtual group

grepom dir                          # Print base directory path
grepom dir web-app                  # Print a repo's local path
grepom dir web --group fe           # Search within a group and print path
cd "$(grepom dir web-app)"          # Quickly jump to a repo directory

# Secret Scanning
grepom scan                         # Scan workspace of all cloned repos
grepom scan -p /path/to/project     # Scan a specific directory directly (no config needed)
grepom scan --group frontend        # Scan only the frontend group
grepom scan --vgroup work           # Scan only real groups in a virtual group
grepom scan --history               # Scan workspace + git history
grepom scan --format json           # Output in JSON format
grepom scan --output results.txt    # Write results to file
grepom scan --gitleaks-config rules.toml  # Use custom rules

# Push Guard
grepom push                         # Scan and push (if no secrets found)
grepom push -f                      # Force push even if secrets found
grepom push -- origin main          # Pass arguments through to git push

# MR/PR Creation
grepom mr                           # Auto-detect and create MR/PR (returns existing if already open)
grepom mr --from feat-x --to main   # Specify source and target branches
grepom mr --title "Add dark mode"   # Custom title
grepom mr --draft                   # Create as draft MR/PR
grepom mr --web                     # Open browser to create
grepom pr                           # Alias for 'mr'

# CI/CD Pipelines
grepom watch                        # Auto-detect repo and watch latest pipeline
grepom watch web-app                # Watch a specific repo's latest pipeline
grepom watch --id 1234              # Watch a specific pipeline by ID
grepom pipeline list <repo-name>    # List pipelines for a repo
grepom pipeline watch <repo-name>   # Watch pipeline status in real-time
grepom tag -w                       # Create version tag, then watch pipeline status

# Service process management
grepom svc run -- make dev         # Start a service in the current directory (default name = dirname)
grepom svc run api                  # Start configured service from .grepom.yml
grepom svc                          # Open TUI directly (default when no subcommand given)
grepom svc list                     # Compact table: name, status, PID, path
grepom svc list -v                  # Full table: also shows command and log path
grepom svc status api               # Show full status for one service
grepom svc logs -f api              # Follow service logs
grepom svc logs --open api          # Open log file in editor
grepom svc kill api                 # Stop a service
grepom svc kill -9 api              # Force stop a service
grepom svc restart api              # Restart a service
grepom svc clean                    # Remove records for exited services
grepom svc dir api                  # Print service working directory
grepom svc tui                      # Explicitly open interactive service management UI
eval "$(grepom svc --shell)"        # Enable gsvc helper for cd to service directories

# TUI keybindings (list view)
# j/k or ↑/↓  Move cursor
# l            View service logs
# s            Stop service (SIGTERM)
# S            Force stop service (SIGKILL)
# R            Restart service
# c            Clean exited service records
# p            Show service path
# r            Refresh list
# d            Show service details
# q            Quit

# Shell completion
eval "$(grepom completion bash)"    # bash completion (or source <(grepom completion bash))
eval "$(grepom completion zsh)"     # zsh completion

# Maintenance
grepom prune                        # Remove cloned repos not in config
grepom prune --vgroup work          # Prune only real groups in a virtual group
grepom dedup                        # Check all groups for intra-group dupes and cross-group warnings
grepom dedup --group core-team      # Check only core-team group
grepom dedup --vgroup work          # Check only real groups in a virtual group
grepom dedup --group core-team --reference infra-team  # Also exclude by name against infra-team
grepom dedup --apply                # Apply changes

# Add resources/groups/repos
grepom add resource --name my-gl --provider gitlab --url https://gitlab.com --token '${GITLAB_TOKEN}'
grepom add group --name frontend --resource my-gl --path my-org/frontend --recursive
grepom add repo --name special --url https://gitlab.com/other/special.git
```

### Token Environment Variables

Token fields support `${ENV_VAR}` placeholder syntax. The actual value is resolved from the environment at runtime, and the placeholder is preserved when writing config files.

```yaml
resources:
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
| `--config` | `-c` | auto-detect | Path to config file (default: searches for `.grepom.yml` upward) |
| `--verbose` | `-v` | `false` | Enable verbose output |

### Service state directory

Service registry and logs are stored under the XDG state directory:

- When `XDG_STATE_HOME` is set: `$XDG_STATE_HOME/grepom/services/<scope>/`
- Otherwise: `~/.local/state/grepom/services/<scope>/`

Data from the legacy `Application Support` (macOS) or `UserConfigDir` location is not migrated automatically. If `grepom svc list` appears empty after upgrading, stop or clean old services and run them again.

## Build

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run vet and format check
make install  # Build and install to ~/.local/bin
make clean    # Remove binary
```
