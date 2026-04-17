# reposcan

A read-only Git worktree dashboard. Monitors multiple repos organized by groups, shows worktrees nested under their parent repos with branch status and colored output.

## Install

```bash
go install github.com/dpetillo/reposcan@latest
```

Or from source:

```bash
make install
```

## Quick Start

```bash
# Generate config from GitLab group hierarchy
cd ~/Dev/directbook1
reposcan init . -g directbook1

# Or generate from local filesystem structure
reposcan init .

# Run the dashboard (refreshes every 10s)
reposcan

# Single scan, no loop
reposcan -once
```

## Config

`config.yaml` (looked up in cwd, then `~/.config/reposcan/config.yaml`):

```yaml
gitlab_group: directbook1
groups:
  - name: Core Platform
    path: ~/Dev/directbook1/core
  - name: SysOps
    path: ~/Dev/directbook1/sysops
  - name: Test
    path: ~/Dev/directbook1/test
  - name: (root)
    path: ~/Dev/directbook1
interval: 10
```

Each group points to a directory containing repos. Repos are flat within the group directory. Worktrees are siblings of their parent repo at the same level — they're auto-discovered via `git worktree list`, not listed in config.

## Flags

| Flag | Description |
|------|-------------|
| `-c <path>` | Config file override |
| `-once` | Scan once and exit |
| `--no-color` | Disable colored output |
| `-version` | Print version |

### init subcommand

| Flag | Description |
|------|-------------|
| `-g <group>` | GitLab group path (uses `glab` API to discover hierarchy) |

`NO_COLOR` env var also disables colors.

## Development

```bash
make test        # unit tests
make test-e2e    # e2e tests (creates real git repos)
make build       # compile binary
make install     # go install
```
