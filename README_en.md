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
- **MR/PR creation** — create GitHub Pull Requests, GitLab Merge Requests, or Codeup merge requests from the CLI; returns existing MR/PR address if one is already open
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

### Agent skill (optional)

The repository includes [`skills/grepom-cli/SKILL.md`](./skills/grepom-cli/SKILL.md) to guide agents through grepom multi-repository tasks, config discovery, and safety gates.

Install into the current project:

```bash
npx skills add sunzhenkai/grepom -s grepom-cli -y
```

Install globally for the user:

```bash
npx skills add sunzhenkai/grepom -s grepom-cli -g -y
```

List available skills first:

```bash
npx skills add sunzhenkai/grepom --list
```

The skill only documents CLI usage; configure MCP tools separately with `grepom mcp install <agent>`. Remote installation is available once this skill is present in the remote repository; on the current branch, use `npx skills add . --list` for local discovery validation.

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

resources:                             # canonical format is a map (key = resource name); list+name still loads
  my-gitlab:
    provider: gitlab
    url: gitlab.example.com            # bare host, or with https:// prefix
    token: ${GITLAB_TOKEN}
    ssh_key: ~/.ssh/id_work            # recommended; common default keys are tried if unset

  my-github:
    provider: github
    url: github.com
    token: ${GITHUB_TOKEN}

groups:
  - name: frontend
    resource: my-gitlab
    path: my-org/frontend
    recursive: true
    exclude_repos:                     # optional: exclude specific repos
      - archived-repo

  - name: my-org
    resource: my-github
    path: my-github-org

virtual_groups:                        # optional: can reference real groups and standalone repos
  work:
    groups:
      - frontend
      - my-org
    repos:                             # optional: top-level standalone repo names
      - dotfiles

repos:                                 # standalone repos (not part of any group)
  - name: dotfiles
    resource: my-github
    url: example/dotfiles.git          # relative path: joined with resource.host
  - name: public-demo
    resource: my-github
    url: https://github.com/example/public-demo.git  # absolute HTTPS: used as-is; public repos can clone anonymously

services:                              # optional local development service definitions
  api:
    cwd: ./backend
    command: make dev
  web:
    cwd: ./frontend
    command:
      - pnpm
      - dev
```

#### URL forms and clone protocol

For standalone repos bound to a `resource`:

| `repo.url` form | Behavior |
|-----------------|----------|
| Relative path (e.g. `org/app.git`) | Joined with `resource.url` into HTTPS/SSH; **SSH first** |
| Absolute `https://` / `http://` | **Used as-is**; HTTPS preferred (including anonymous public clones) |
| Absolute `git@` / `ssh://` | **Used as-is for SSH**; never re-concatenated onto resource.host |

If `ssh_key` is unset, after default SSH fails grepom tries `~/.ssh/id_ed25519`, `id_rsa`, `id_ecdsa`, `id_ed25519_sk` when present. Clone failures include sanitized git stderr; use `-v` for full per-step reasons.

> GitLab `group.path` supports both group/subgroup paths (for example, `my-org/frontend`) and personal namespaces (for example, `sunzhenkai`). When a personal namespace is configured, `grepom sync` automatically switches to the user projects API.

> **Recycle-bin / scheduled-deletion repos**: When Codeup (Yunxiao) deletes a repo, it enters a recycle bin and is renamed to include a `deletion_scheduled` marker (for example, `repo-deletion_scheduled-499`); such repos are no longer clonable. `grepom sync` and `grepom list --remote` skip these by default (verbose mode `-v` reports the skip count), and any already saved in config are skipped at runtime too, avoiding misleading authentication errors. Use `--include-deleted` to include them.

### Commands

```bash
# Init & Config
grepom init                         # Initialize config file
grepom example                      # Export complete example config
grepom interactive                  # Start interactive mode

# Sync & Discovery
grepom sync                         # Discover repos and update config metadata
grepom sync --resource my-gitlab     # Sync a specific resource by name
grepom sync --group frontend        # Sync a specific group
grepom sync --vgroup work           # Sync all real groups in a virtual group
grepom sync --include-deleted       # Include recycle-bin (deletion_scheduled) repos

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

# Parallel progress display
# When cloning/pulling concurrently (`-j N`), TTY mode renders a live multi-line
# progress area: the first line is a `[done/total] cloning...` summary, and each
# subsequent line shows an in-flight repo name. The renderer is concurrency-safe:
# the `[N/M]` counter is monotonic non-decreasing, stale repo names are cleared
# with blank lines when the in-flight count shrinks (no overlap/garble/regress),
# and non-TTY mode (pipes, CI) prints one `✓ repo` / `✗ repo: err` line per result.

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
grepom list --remote --include-deleted  # Include recycle-bin repos in remote listing

grepom status                       # Check status of all cloned repos
grepom status web-app               # Status of a specific repo
grepom status --vgroup work         # Status for all real groups in a virtual group

grepom search web                   # Search repos by name (substring match)
grepom search web --group frontend  # Search within a specific group
grepom search web --vgroup work     # Search within all real groups in a virtual group

grepom dir                          # Print config directory path
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

# MR/PR Creation (GitLab / GitHub / Codeup; Codeup requires organization_id in the resource)
grepom mr                           # Auto-detect and create MR/PR (returns existing if already open)
grepom mr --from feat-x --to main   # Specify source and target branches
grepom mr --title "Add dark mode"   # Custom title
grepom mr --draft                   # Create as draft MR/PR
grepom mr --web                     # Open browser to create
grepom pr                           # Alias for 'mr'

# CI/CD Pipelines (GitLab / GitHub; Codeup via Yunxiao Flow, requires organization_id)
grepom watch                        # Auto-detect repo and watch latest pipeline
grepom watch web-app                # Watch a specific repo's latest pipeline
grepom watch --id 1234              # Watch a specific pipeline by ID
grepom pipeline list <repo-name>    # List pipelines for a repo
grepom pipeline watch <repo-name>   # Watch pipeline status in real-time
grepom tag -w                       # Create version tag, then watch that tag's pipeline (bound by commit SHA, waits up to 60s, errors on timeout)

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

# MCP (AI Agent integration)
grepom mcp serve                    # Start stdio MCP Server (for agents)
grepom mcp list-tools               # List all MCP tools
grepom mcp install cursor           # Write MCP server config into Cursor
grepom mcp install claude-code      # Write into Claude Code
grepom mcp install --all            # Write into all supported agents
grepom mcp install cursor --print   # Print config snippet only (dry-run)
```

### MCP Integration

grepom exposes its multi-repo management capabilities to local AI agents (Claude Code, Cursor, Codex, etc.) via the MCP (Model Context Protocol) stdio transport.

**Quick setup:**

```bash
grepom mcp install cursor   # One-command agent config
# Restart the agent; grepom_* tools are now available
```

**Available tools:**

| Tool | Description |
|------|-------------|
| `grepom_list` | List repositories (optional group filter) |
| `grepom_status` | Query repo git status |
| `grepom_search` | Fuzzy search repos by name |
| `grepom_dir` | Get repo local path |
| `grepom_pull` | Pull a specific repo |
| `grepom_clone` | Clone a specific repo |
| `grepom_groups` | List configured groups |
| `grepom_scan` | Secret scanning |

### Token Environment Variables

Token fields support `${ENV_VAR}` placeholder syntax. The actual value is resolved from the environment at runtime, and the placeholder is preserved when writing config files.

```yaml
resources:
  my-gitlab:
    provider: gitlab
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
