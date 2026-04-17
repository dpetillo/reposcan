package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiscoverAndScanGroupFindsRepos(t *testing.T) {
	tmp := t.TempDir()

	// Create two repos in the group dir
	for _, name := range []string{"repo-a", "repo-b"} {
		rp := filepath.Join(tmp, name)
		os.MkdirAll(filepath.Join(rp, ".git"), 0755)
	}
	// Create a worktree dir (has .git file, not dir) — should be skipped
	wtDir := filepath.Join(tmp, "repo-a-feature")
	os.MkdirAll(wtDir, 0755)
	os.WriteFile(filepath.Join(wtDir, ".git"), []byte("gitdir: ../repo-a/.git/worktrees/feat"), 0644)

	results := discoverAndScanGroup(Group{Name: "test", Path: tmp})
	if len(results) != 2 {
		names := make([]string, len(results))
		for i, r := range results {
			names[i] = r.Name
		}
		t.Fatalf("got %d repos %v, want 2 (repo-a, repo-b)", len(results), names)
	}
}

func TestDiscoverAndScanGroupSkipsNonRepos(t *testing.T) {
	tmp := t.TempDir()

	// Regular dir (no .git)
	os.MkdirAll(filepath.Join(tmp, "not-a-repo"), 0755)
	// File
	os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("hi"), 0644)

	results := discoverAndScanGroup(Group{Name: "test", Path: tmp})
	if len(results) != 0 {
		t.Errorf("got %d repos, want 0", len(results))
	}
}

func TestDiscoverAndScanGroupBadPath(t *testing.T) {
	results := discoverAndScanGroup(Group{Name: "test", Path: "/nonexistent"})
	if results != nil {
		t.Errorf("expected nil for bad path, got %d results", len(results))
	}
}

func TestScanRepoWithWorktrees(t *testing.T) {
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "myrepo")

	cmd := exec.Command("git", "init", repoPath)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=<email>",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=<email>")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	os.WriteFile(filepath.Join(repoPath, "f.txt"), []byte("x"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=<email>",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=<email>")
	cmd.Run()

	rr := scanRepo(repoPath)
	if rr.Name != "myrepo" {
		t.Errorf("name = %q, want myrepo", rr.Name)
	}
	if len(rr.Worktrees) < 1 {
		t.Errorf("worktrees = %d, want >= 1", len(rr.Worktrees))
	}
}
