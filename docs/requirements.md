# reposcan — Requirements

## Problem Statement

71+ git repos under ~/Dev/directbook1/ with worktrees scattered as sibling directories. The existing worktree-monitor.sh only works for one repo at a time. Need a single Go CLI tool that monitors all repos, groups worktrees under their parent repos, auto-refreshes in parallel, and collapses clean/inactive repos into a summary line.

---

## Functional Requirements

### REQ-001: Read-Only Dashboard
- Display-only tool — no worktree creation, deletion, or modification
- **AC:** Tool never runs `git worktree add/remove/prune` or any mutating git commands

### REQ-002: Installable to PATH
- Installable via `go install github.com/dpetillo/reposcan@latest`
- Operates on the current working directory (scans repos listed in config)
- **AC:** After `go install`, running `reposcan` from any directory works if config is found

### REQ-003: YAML Config with Group Labels
- Config file lists repos with path and group label
- Tool auto-discovers worktrees for each listed repo
- **AC:** Given a config.yaml with repos and groups, tool scans only listed repos and groups output by label

### REQ-004: Auto-Refresh with Parallel Scanning
- Configurable refresh interval (default: 10 seconds)
- All repos scanned in parallel via goroutines
- **AC:** Dashboard refreshes every N seconds; scan duration shown in footer; parallelism observable via timing

### REQ-005: Worktree Grouping Under Parent Repo
- Worktrees displayed nested/indented under their parent repo
- Parent repo identified via `.git` file → `gitdir:` pointer resolution
- **AC:** Worktrees like `feature-TEC-3272-add-winnfield-qa` appear under `cloudflare-pipeline`

### REQ-006: Clean Repo Collapse
- Repos with no worktrees (just main) and clean status are not rendered individually
- Collapsed into a summary count line at the bottom (e.g., "42 repos clean (no worktrees)")
- **AC:** Clean repos with no worktrees do not appear in the main listing; count shown at bottom

### REQ-007: Per-Worktree Status
- Branch name
- Modified file count
- Staged file count
- Untracked file count
- Ahead/behind remote counts
- **AC:** Each worktree line shows branch, M/S/U counts, and ahead/behind; no-upstream shows 0/0

### REQ-008: Colored Terminal Output
- Group headers in bold white
- Repo names in cyan, branches in blue
- Status colors: red (modified), green (staged), yellow (untracked), green (ahead), red (behind), green checkmark (clean)
- **AC:** Output uses ANSI escape codes; colors render correctly in standard terminals

### REQ-009: `init` Subcommand
- `reposcan init [dir]` scans a directory (default: cwd), discovers git repos, infers groups from parent directory name
- Writes config.yaml to cwd
- Skips worktree directories (`.git` is a file, not a directory)
- **AC:** Running `reposcan init .` in ~/Dev/directbook1 produces a config.yaml with all repos grouped correctly

### REQ-010: Test Coverage
- Unit tests for config parsing, worktree list parsing, status parsing, grouping logic, rendering
- E2e tests using real git repos created in temp directories
- **AC:** `go test ./...` passes; e2e tests create real repos with worktrees and verify full pipeline

---

## Technical Requirements

### TECH-001: Config Resolution
- Look for `config.yaml` in cwd first
- Fall back to `~/.config/reposcan/config.yaml`
- `-c` flag overrides both
- **AC:** Config loaded from correct location based on precedence; error if none found

### TECH-002: Config Struct
```go
type Repo struct {
    Path  string `yaml:"path"`
    Group string `yaml:"group"`
}
type Config struct {
    Repos    []Repo `yaml:"repos"`
    Interval int    `yaml:"interval"` // seconds, default 10
}
```
- **AC:** YAML with repos and interval deserializes correctly; missing interval defaults to 10

### TECH-003: CLI Flags
- `-c <path>`: config file override
- `-once`: single scan, no loop
- `-version`: print version and exit
- **AC:** All flags parsed and applied correctly

### TECH-004: Worktree Discovery
- Run `git worktree list --porcelain` per repo
- Parse output into Worktree structs (path, branch, isMain)
- **AC:** Correctly parses single worktree, multiple worktrees, detached HEAD entries

### TECH-005: Submodule vs Worktree Distinction
- `.git` files with relative `gitdir:` paths (e.g., `../.git/modules/...`) are submodules — skip them
- `.git` files with absolute `gitdir:` paths pointing to `.../.git/worktrees/...` are actual worktrees
- **AC:** Submodule directories are never reported as worktrees

### TECH-006: Status Collection
- `git status --porcelain` for modified/staged/untracked counts
- `git rev-list --count HEAD..@{upstream}` and `@{upstream}..HEAD` for behind/ahead
- Handle no upstream gracefully (0/0)
- **AC:** Correct counts for clean, dirty, and no-upstream scenarios

### TECH-007: Parallel Scanning
- Use `errgroup` to fan out repo scans
- Collect results into grouped `ScanResult`
- **AC:** All repos scanned concurrently; errors from individual repos don't crash the tool

### TECH-008: Terminal Rendering
- ANSI escape codes for colors
- `\033[H\033[2J` to clear screen between refreshes
- `NO_COLOR` env var or `--no-color` flag disables colors
- **AC:** Renders correctly with and without color; screen clears between refreshes

### TECH-009: Module Path
- `module github.com/dpetillo/reposcan`
- **AC:** `go install github.com/dpetillo/reposcan@latest` resolves correctly

### TECH-010: Makefile
- `make build` — compile binary
- `make install` — `go install .`
- `make test` — `go test ./...`
- `make test-e2e` — `go test -tags e2e ./...`
- **AC:** All targets work from repo root
