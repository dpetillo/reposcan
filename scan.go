package main

import (
	"path/filepath"
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
	Group     string
	Worktrees []WorktreeResult
}

type ScanResult struct {
	Groups   map[string][]RepoResult
	Duration time.Duration
}

func RunScan(cfg Config) ScanResult {
	start := time.Now()
	results := make([]RepoResult, len(cfg.Repos))
	var wg sync.WaitGroup

	for i, repo := range cfg.Repos {
		wg.Add(1)
		go func(idx int, r Repo) {
			defer wg.Done()
			results[idx] = scanRepo(r)
		}(i, repo)
	}
	wg.Wait()

	return ScanResult{
		Groups:   groupResults(results),
		Duration: time.Since(start),
	}
}

func scanRepo(r Repo) RepoResult {
	rr := RepoResult{
		Name:  filepath.Base(r.Path),
		Path:  r.Path,
		Group: r.Group,
	}
	wts, err := ScanWorktrees(r.Path)
	if err != nil {
		return rr
	}
	for _, wt := range wts {
		st, _ := CollectStatus(wt.Path)
		rr.Worktrees = append(rr.Worktrees, WorktreeResult{Worktree: wt, Status: st})
	}
	return rr
}

func groupResults(results []RepoResult) map[string][]RepoResult {
	groups := make(map[string][]RepoResult)
	for _, r := range results {
		g := r.Group
		if g == "" {
			g = "default"
		}
		groups[g] = append(groups[g], r)
	}
	return groups
}
