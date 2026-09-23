---
name: grepom-cli
description: "Use the grepom CLI to inspect and manage multiple Git repositories from a .grepom.yml configuration: discover, clone, pull, list, search, status, scan, push, create MR/PR, watch pipelines, manage local services, maintain config, and set up grepom MCP tools. Use when a task involves grepom or coordinating many GitLab/GitHub repositories; do not use for generic git operations that do not involve grepom configuration."
---

# grepom CLI

Use `grepom` as the orchestration layer for repositories declared in `.grepom.yml`. Prefer its filters and safety checks over reconstructing multi-repo shell loops.

## Establish the context

1. Confirm `grepom` is available with `grepom version`; use `grepom --help` or `grepom <command> --help` when an unfamiliar option is needed.
2. Resolve the active config before acting. With no `-c/--config`, grepom searches the current directory and then parent directories for `.grepom.yml`, like git searches for `.git`. Use `-c/--config <path>` when the target config is elsewhere or the upward search could select the wrong config.
3. Use `grepom dir` to print the config directory and `grepom dir <repo>` for one repo path. Use `grepom search <keyword>` for case-insensitive name matching and `grepom status [name]` for branch, dirty, ahead, and behind state.
4. Token fields support `${ENV_VAR}` placeholders. Keep placeholders in config and command examples; never write a resolved token, credential, or other secret into the repository, command history, output, or notes.

Common filters are `-g/--group`, `-V/--vgroup`, and `-R/--resource`. A virtual group expands to its member groups; combining `--group` and `--vgroup` selects their union, while `--resource` further narrows the selected set.

## Choose the workflow

### Configure and discover

```bash
grepom init
grepom example -o .grepom.yml
grepom add resource --name my-gitlab --provider gitlab --url https://gitlab.example.com --token '${GITLAB_TOKEN}'
grepom add group --name frontend --resource my-gitlab --path my-org/frontend --recursive
grepom add repo --name public-demo --resource my-github --url https://github.com/example/public-demo.git
grepom sync                         # Discover remote repos and update config metadata only
grepom sync --group frontend
grepom sync --resource my-gitlab
```

`sync` adds newly discovered repos to config; it does not clone or pull, and it does not remove existing entries. Follow it with `clone` for new repos and `pull` for existing repos.

### Clone and update

```bash
grepom clone                        # All configured repos, default 4 workers
grepom clone web-app
grepom clone --group frontend --concurrency 8
grepom pull                         # Safe pull: default branch and clean tree only
grepom pull web-app
grepom pull --group frontend
```

`pull -f/--force` skips the default-branch and clean-worktree safety checks. Use it only after checking `grepom status` or when the requested task explicitly authorizes the risk.

### Inspect and query

```bash
grepom list                         # Repos needing attention: unpushed or dirty
grepom list --all
grepom list --no-push
grepom list --no-commit
grepom list groups
grepom list resources
grepom list --remote --group frontend
grepom status
grepom search api --group backend
grepom dir --group frontend
```

`list --remote` queries the provider API instead of local config. `list` without `--all` intentionally hides clean repos; do not interpret that as “all repos are clean.”

### Scan and protect pushes

```bash
grepom scan                         # Workspace scan of configured cloned repos
grepom scan -p /path/to/project     # Scan one directory without config
grepom scan --history               # Include git history
grepom scan --format json
grepom scan --output results.txt
grepom push                         # Scan current repo, then git push only if clean
grepom push -- origin main          # Pass remaining args through to git push
```

`push` rejects the push when findings exist. `-f/--force` still pushes and only prints a warning, so use it only with explicit user authorization and a reviewed result. `scan --history` and `--output` can expose historical data; choose the narrowest scope needed.

### Create MR/PR and watch CI/CD

```bash
grepom mr                           # Auto-detect source/target/provider/title
grepom mr --from feat-x --to main
grepom mr --title "Add dark mode" --body-file changes.md
grepom mr --draft
grepom pr                           # Alias for mr
grepom watch                        # Watch latest pipeline for current repo
grepom watch web-app --id 1234
grepom pipeline list web-app -n 10
grepom pipeline watch web-app
```

`mr` and `pr` can create a remote merge request. Before running either, show the source branch, target branch, title, and whether commits are already pushed; ask when the target or publication scope is ambiguous. `watch` polls until a terminal state, so use it when blocking on CI is intended.

### Manage local services

```bash
grepom svc run -- make dev
grepom svc run api
grepom svc list
grepom svc status api
grepom svc logs -n 200 api
grepom svc restart api
grepom svc kill api
grepom svc clean
```

`svc run` starts a background process and `svc kill` stops it. Confirm service names and working directories before starting or replacing a process, especially with `run --force`.

### Maintain config safely

```bash
grepom dedup                        # Dry-run duplicate analysis
grepom dedup --group core-team
grepom dedup --apply                # Write deduplication changes
grepom prune                        # Dry-run deletion of excluded cloned repos
grepom prune --apply                # Delete excluded repos when safe
grepom prune --apply --force        # Also delete dirty or ahead repos
```

`dedup` and `prune` are dry-run by default. Read the proposed changes first; require explicit authorization before `--apply`. Treat `prune --force` as destructive because it can remove dirty or unpushed work.

### MCP integration

```bash
grepom mcp list-tools
grepom mcp serve                    # stdio server for an MCP-capable agent
grepom mcp install cursor --print   # Preview config only
grepom mcp install cursor           # Write config with .bak backup
grepom mcp install codex --scope project
```

Supported install targets include `claude-code`, `claude-desktop`, `cursor`, `codex`, `zcode`, `kimi`, and `pi`. Installing MCP config changes the target agent configuration; preview with `--print` and obtain authorization before writing it. The skill supplies CLI guidance; it does not install MCP configuration by itself.

## Safety and verification

- Prefer read-only commands (`list`, `status`, `search`, `dir`, `scan`) before mutating commands.
- Show the exact config path, filters, and intended target before `sync`, `clone`, `pull --force`, `push -f`, `dedup --apply`, `prune --apply`, `mr`, `pr`, `tag`, or service replacement.
- Use `grepom tag --dry-run` to preview a version tag. `grepom tag -p` pushes to all remotes; treat tag creation and pushing as online actions requiring explicit authorization.
- Use `-v/--verbose` only when needed for diagnosis, and redact credentials from any captured output.
- Verify the result after execution with the matching read-only command, for example `grepom status <repo>` after `pull`, `grepom svc status <name>` after `run`, or `grepom list --remote --group frontend` after `sync`.
