package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
	IsMain bool
}

func ScanWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list in %s: %w", repoPath, err)
	}
	return parseWorktreeList(out), nil
}

func parseWorktreeList(data []byte) []Worktree {
	var worktrees []Worktree
	var current Worktree
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "detached":
			current.Branch = "(detached)"
		case line == "":
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
		}
	}
	// Flush last entry if no trailing blank line
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	// Mark the first worktree as main (it's always the main working tree)
	if len(worktrees) > 0 {
		worktrees[0].IsMain = true
	}
	return worktrees
}
