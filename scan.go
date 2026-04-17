package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type WorktreeResult struct {
	Worktree
	Status WorktreeStatus
}

type RepoResult struct {
	Name      string
	Path      string
	Worktrees []WorktreeResult
}

type ScanResult struct {
	Groups   map[string][]RepoResult
	Duration time.Duration
}

const defaultInterval = 10

func RunScan(root string) ScanResult {
	start := time.Now()
	repos := discoverRepos(root)

	results := make([]RepoResult, len(repos))
	var wg sync.WaitGroup
	for i, rp := range repos {
		wg.Add(1)
		go func(idx int, repoPath string) {
			defer wg.Done()
			results[idx] = scanRepo(repoPath)
		}(i, rp)
	}
	wg.Wait()

	return ScanResult{
		Groups:   groupByParentDir(root, results),
		Duration: time.Since(start),
	}
}

// discoverRepos walks root for dirs containing a .git directory (not file).
// Skips .git files (worktrees/submodules). Once a repo is found, does not descend into it.
func discoverRepos(root string) []string {
	var repos []string
	repoSet := make(map[string]bool)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// If we're inside an already-found repo, skip
		if info.IsDir() && path != root {
			if repoSet[path] {
				return filepath.SkipDir
			}
			// Check for .git inside this dir
			gitPath := filepath.Join(path, ".git")
			gi, err := os.Stat(gitPath)
			if err == nil {
				if gi.IsDir() {
					repos = append(repos, path)
					repoSet[path] = true
				}
				// Whether .git dir (repo) or .git file (worktree) — don't descend
				return filepath.SkipDir
			}
		}
		return nil
	})
	return repos
}

func groupByParentDir(root string, results []RepoResult) map[string][]RepoResult {
	groups := make(map[string][]RepoResult)
	for _, r := range results {
		rel, err := filepath.Rel(root, r.Path)
		if err != nil {
			groups["other"] = append(groups["other"], r)
			continue
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		group := "."
		if len(parts) > 1 {
			group = parts[0]
		}
		groups[group] = append(groups[group], r)
	}
	return groups
}

func scanRepo(repoPath string) RepoResult {
	rr := RepoResult{
		Name: filepath.Base(repoPath),
		Path: repoPath,
	}
	wts, err := ScanWorktrees(repoPath)
	if err != nil {
		return rr
	}
	for _, wt := range wts {
		st, _ := CollectStatus(wt.Path)
		rr.Worktrees = append(rr.Worktrees, WorktreeResult{Worktree: wt, Status: st})
	}
	return rr
}
