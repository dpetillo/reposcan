package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
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

	cleanCounts := make(map[string]int)
	groupNames := sortedKeys(result.Groups)

	for _, groupName := range groupNames {
		repos := result.Groups[groupName]
		for _, repo := range repos {
			if isFullyClean(repo) {
				cleanCounts[groupName]++
				continue
			}
			// Repo line: [group] repo-name  main-status
			mainStatus := mainWorktreeStatus(repo, c)
			fmt.Fprintf(w, "%s %s %s\n", c(dim, "["+groupName+"]"), c(cyan, repo.Name), mainStatus)

			// Non-main worktrees on one line, skip clean ones
			var wtParts []string
			for _, wt := range repo.Worktrees {
				if wt.IsMain {
					continue
				}
				s := compactWorktree(wt, c)
				if s != "" {
					wtParts = append(wtParts, s)
				}
			}
			if len(wtParts) > 0 {
				fmt.Fprintf(w, "  %s\n", strings.Join(wtParts, c(dim, " │ ")))
			}
		}
	}

	// Clean summary — one line
	var cleanParts []string
	for _, g := range groupNames {
		if n := cleanCounts[g]; n > 0 {
			cleanParts = append(cleanParts, fmt.Sprintf("%s:%d", g, n))
		}
	}
	if len(cleanParts) > 0 {
		fmt.Fprintf(w, "%s\n", c(green, "✓ clean "+strings.Join(cleanParts, " ")))
	}

	// Footer
	fmt.Fprintf(w, "%s %s\n",
		c(dim, result.Duration.Round(time.Millisecond).String()),
		c(dim, time.Now().Format("15:04:05")))
}

func mainWorktreeStatus(repo RepoResult, c func(string, string) string) string {
	for _, wt := range repo.Worktrees {
		if wt.IsMain {
			return compactStatus(wt.Status, c)
		}
	}
	return ""
}

func compactWorktree(wt WorktreeResult, c func(string, string) string) string {
	s := wt.Status
	// Skip clean worktrees with no ahead/behind
	if s.Modified == 0 && s.Staged == 0 && s.Untracked == 0 && s.Ahead == 0 && s.Behind == 0 {
		return ""
	}
	branch := shortBranch(wt.Branch)
	return c(blue, branch) + " " + compactStatus(s, c)
}

func compactStatus(s WorktreeStatus, c func(string, string) string) string {
	if s.Modified == 0 && s.Staged == 0 && s.Untracked == 0 && s.Ahead == 0 && s.Behind == 0 {
		return c(green, "✓")
	}
	var parts []string
	if s.Modified > 0 {
		parts = append(parts, c(red, fmt.Sprintf("M%d", s.Modified)))
	}
	if s.Staged > 0 {
		parts = append(parts, c(green, fmt.Sprintf("S%d", s.Staged)))
	}
	if s.Untracked > 0 {
		parts = append(parts, c(yellow, fmt.Sprintf("U%d", s.Untracked)))
	}
	if s.Ahead > 0 {
		parts = append(parts, c(green, fmt.Sprintf("↑%d", s.Ahead)))
	}
	if s.Behind > 0 {
		parts = append(parts, c(red, fmt.Sprintf("↓%d", s.Behind)))
	}
	return strings.Join(parts, " ")
}

// shortBranch strips common prefixes to save space
func shortBranch(branch string) string {
	for _, prefix := range []string{"feature-", "feat-", "bugfix-", "hotfix-", "release-"} {
		if strings.HasPrefix(branch, prefix) {
			return branch[len(prefix):]
		}
	}
	return branch
}

// isFullyClean returns true if repo has only main worktree and it's clean
func isFullyClean(repo RepoResult) bool {
	if len(repo.Worktrees) != 1 {
		return false
	}
	wt := repo.Worktrees[0]
	return wt.IsMain && wt.Status == (WorktreeStatus{})
}

func isCleanNoWorktrees(repo RepoResult) bool {
	return isFullyClean(repo)
}

func sortedKeys(m map[string][]RepoResult) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
