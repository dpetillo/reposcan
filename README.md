# reposcan

A read-only Git worktree dashboard. Monitors multiple repos, groups worktrees under their parent repos, and shows branch status with colored output.

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
# Generate config from a directory of repos
cd ~/Dev/directbook1
reposcan init .

# Run the dashboard (refreshes every 10s)
reposcan

# Single scan, no loop
reposcan -once

# Custom config
reposcan -c /path/to/config.yaml
```

## Config

`config.yaml` (looked up in cwd, then `~/.config/reposcan/config.yaml`):

```yaml
interval: 10
repos:
  - path: ~/Dev/directbook1/core-monorepo
    group: core
  - path: ~/Dev/directbook1/ghost-monorepo
    group: ghost
  - path: ~/Dev/directbook1/cloudflare-pipeline
    group: sysops
```

## Flags

| Flag | Description |
|------|-------------|
| `-c <path>` | Config file override |
| `-once` | Scan once and exit |
| `--no-color` | Disable colored output |
| `-version` | Print version |

`NO_COLOR` env var also disables colors.

## Development

```bash
make test        # unit tests
make test-e2e    # e2e tests (creates real git repos)
make build       # compile binary
make install     # go install
```
