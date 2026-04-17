package main

import (
	"fmt"
	"io"
	"sort"
	"time"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	blue      = "\033[34m"
	cyan      = "\033[36m"
	boldWhite = "\033[1;37m"
	clearScr  = "\033[H\033[2J"
)

func Render(result ScanResult, w io.Writer, noColor bool) {
	c := func(code, text string) string {
		if noColor {
			return text
		}
		return code + text + reset
	}

	if !noColor {
		fmt.Fprint(w, clearScr)
	}

	cleanCount := 0
	groupNames := sortedKeys(result.Groups)

	for _, groupName := range groupNames {
		repos := result.Groups[groupName]
		var activeRepos []RepoResult
		for _, repo := range repos {
			if isCleanNoWorktrees(repo) {
				cleanCount++
			} else {
				activeRepos = append(activeRepos, repo)
			}
		}
		if len(activeRepos) == 0 {
			continue
		}

		fmt.Fprintf(w, "\n%s\n", c(boldWhite, "═══ "+groupName+" ═══"))
		for _, repo := range activeRepos {
			fmt.Fprintf(w, "  %s\n", c(cyan, repo.Name))
			for _, wt := range repo.Worktrees {
				prefix := "  ├─"
				if wt.IsMain {
					prefix = "  ├─"
				}
				branchStr := c(blue, wt.Branch)
				statusStr := formatStatus(wt.Status, c)
				fmt.Fprintf(w, "  %s %s %s\n", prefix, branchStr, statusStr)
			}
		}
	}

	// Clean repos summary
	if cleanCount > 0 {
		fmt.Fprintf(w, "\n%s\n", c(green, fmt.Sprintf("✓ %d repos clean (no worktrees)", cleanCount)))
	}

	// Footer
	fmt.Fprintf(w, "\n%s  %s\n",
		c(bold, fmt.Sprintf("Scanned in %s", result.Duration.Round(time.Millisecond))),
		time.Now().Format("15:04:05"))
}

func formatStatus(s WorktreeStatus, c func(string, string) string) string {
	if s.Modified == 0 && s.Staged == 0 && s.Untracked == 0 && s.Ahead == 0 && s.Behind == 0 {
		return c(green, "✓ clean")
	}
	var parts []string
	if s.Modified > 0 {
		parts = append(parts, c(red, fmt.Sprintf("M:%d", s.Modified)))
	}
	if s.Staged > 0 {
		parts = append(parts, c(green, fmt.Sprintf("S:%d", s.Staged)))
	}
	if s.Untracked > 0 {
		parts = append(parts, c(yellow, fmt.Sprintf("U:%d", s.Untracked)))
	}
	if s.Ahead > 0 {
		parts = append(parts, c(green, fmt.Sprintf("+%d", s.Ahead)))
	}
	if s.Behind > 0 {
		parts = append(parts, c(red, fmt.Sprintf("-%d", s.Behind)))
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

func isCleanNoWorktrees(repo RepoResult) bool {
	if len(repo.Worktrees) != 1 {
		return false
	}
	wt := repo.Worktrees[0]
	return wt.IsMain && wt.Status.Modified == 0 && wt.Status.Staged == 0 &&
		wt.Status.Untracked == 0 && wt.Status.Ahead == 0 && wt.Status.Behind == 0
}

func sortedKeys(m map[string][]RepoResult) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
