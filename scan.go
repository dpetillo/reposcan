package main

import (
	"os"
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
	Worktrees []WorktreeResult
}

type ScanResult struct {
	Groups   map[string][]RepoResult
	Duration time.Duration
}

func RunScan(cfg Config) ScanResult {
	start := time.Now()
	groups := make(map[string][]RepoResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, g := range cfg.Groups {
		wg.Add(1)
		go func(grp Group) {
			defer wg.Done()
			repos := discoverAndScanGroup(grp)
			mu.Lock()
			groups[grp.Name] = repos
			mu.Unlock()
		}(g)
	}
	wg.Wait()

	return ScanResult{Groups: groups, Duration: time.Since(start)}
}

func discoverAndScanGroup(g Group) []RepoResult {
	entries, err := os.ReadDir(g.Path)
	if err != nil {
		return nil
	}

	// Find repos (dirs with .git directory, not .git file)
	var repoPaths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gitPath := filepath.Join(g.Path, e.Name(), ".git")
		info, err := os.Stat(gitPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			repoPaths = append(repoPaths, filepath.Join(g.Path, e.Name()))
		}
		// .git file = worktree, skip — it belongs to a parent repo
	}

	// Scan repos in parallel
	results := make([]RepoResult, len(repoPaths))
	var wg sync.WaitGroup
	for i, rp := range repoPaths {
		wg.Add(1)
		go func(idx int, repoPath string) {
			defer wg.Done()
			results[idx] = scanRepo(repoPath)
		}(i, rp)
	}
	wg.Wait()
	return results
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
