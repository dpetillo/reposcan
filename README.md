# reposcan

Zero-config Git worktree dashboard. Scans a directory tree for repos, shows worktrees with branch status.

## Install

```bash
go install github.com/dpetillo/reposcan@latest
```

## Usage

```bash
# Scan current directory
reposcan

# Scan a specific directory
reposcan ~/Dev/directbook1

# Single scan, no refresh loop
reposcan -once .

# Custom refresh interval
reposcan -interval 5 .
```

No config files needed. Repos are discovered by walking the directory tree for `.git` directories. Worktrees (`.git` files) and submodules are skipped. Groups are inferred from the first subdirectory level.

## Flags

| Flag | Description |
|------|-------------|
| `-once` | Scan once and exit |
| `-interval N` | Refresh interval in seconds (default: 10) |
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
