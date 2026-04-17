package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type WorktreeStatus struct {
	Modified  int
	Staged    int
	Untracked int
	Ahead     int
	Behind    int
}

func CollectStatus(worktreePath string) (WorktreeStatus, error) {
	var s WorktreeStatus

	// git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = worktreePath
	out, err := cmd.Output()
	if err != nil {
		return s, fmt.Errorf("git status in %s: %w", worktreePath, err)
	}
	s.Modified, s.Staged, s.Untracked = parseStatusPorcelain(out)

	// ahead/behind
	s.Ahead, s.Behind = getAheadBehind(worktreePath)
	return s, nil
}

func parseStatusPorcelain(data []byte) (modified, staged, untracked int) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]
		switch {
		case x == '?' && y == '?':
			untracked++
		default:
			if x != ' ' && x != '?' {
				staged++
			}
			if y != ' ' && y != '?' {
				modified++
			}
		}
	}
	return
}

func getAheadBehind(worktreePath string) (ahead, behind int) {
	// ahead
	cmd := exec.Command("git", "rev-list", "--count", "@{upstream}..HEAD")
	cmd.Dir = worktreePath
	if out, err := cmd.Output(); err == nil {
		ahead, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}
	// behind
	cmd = exec.Command("git", "rev-list", "--count", "HEAD..@{upstream}")
	cmd.Dir = worktreePath
	if out, err := cmd.Output(); err == nil {
		behind, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}
	return
}
