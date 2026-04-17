package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiscoverReposFindsGitDirs(t *testing.T) {
	tmp := t.TempDir()
	// Real repo (.git dir)
	os.MkdirAll(filepath.Join(tmp, "repo-a", ".git"), 0755)
	os.MkdirAll(filepath.Join(tmp, "repo-b", ".git"), 0755)
	// Worktree (.git file) — should be skipped
	wtDir := filepath.Join(tmp, "repo-a-feat")
	os.MkdirAll(wtDir, 0755)
	os.WriteFile(filepath.Join(wtDir, ".git"), []byte("gitdir: ../repo-a/.git/worktrees/feat"), 0644)
	// Not a repo
	os.MkdirAll(filepath.Join(tmp, "just-a-dir"), 0755)

	repos := discoverRepos(tmp)
	if len(repos) != 2 {
		t.Fatalf("got %d repos, want 2: %v", len(repos), repos)
	}
}

func TestDiscoverReposNested(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "group1", "repo1", ".git"), 0755)
	os.MkdirAll(filepath.Join(tmp, "group1", "repo2", ".git"), 0755)
	os.MkdirAll(filepath.Join(tmp, "group2", "repo3", ".git"), 0755)

	repos := discoverRepos(tmp)
	if len(repos) != 3 {
		t.Fatalf("got %d repos, want 3", len(repos))
	}
}

func TestDiscoverReposDoesNotDescendIntoRepo(t *testing.T) {
	tmp := t.TempDir()
	// Repo with a nested .git inside (submodule-like) — should not be found
	os.MkdirAll(filepath.Join(tmp, "repo", ".git"), 0755)
	os.MkdirAll(filepath.Join(tmp, "repo", "vendor", "lib", ".git"), 0755)

	repos := discoverRepos(tmp)
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1 (should not descend into repo): %v", len(repos), repos)
	}
}

func TestGroupByParentDir(t *testing.T) {
	root := "/home/user/dev"
	results := []RepoResult{
		{Name: "repo1", Path: "/home/user/dev/core/repo1"},
		{Name: "repo2", Path: "/home/user/dev/core/repo2"},
		{Name: "repo3", Path: "/home/user/dev/test/repo3"},
		{Name: "toplevel", Path: "/home/user/dev/toplevel"},
	}
	groups := groupByParentDir(root, results)
	if len(groups["core"]) != 2 {
		t.Errorf("core = %d, want 2", len(groups["core"]))
	}
	if len(groups["test"]) != 1 {
		t.Errorf("test = %d, want 1", len(groups["test"]))
	}
	if len(groups["."]) != 1 {
		t.Errorf(". = %d, want 1", len(groups["."]))
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
	cmd = exec.Command("git", "-C", repoPath, "add", ".")
	cmd.Run()
	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "init")
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
