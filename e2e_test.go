//go:build e2e

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=<email>",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=<email>",
	)
	return cmd.Run()
}

func setupTestRepo(t *testing.T, dir, name string) string {
	t.Helper()
	repoPath := filepath.Join(dir, name)
	os.MkdirAll(repoPath, 0755)
	gitInDir(repoPath, "init")
	gitInDir(repoPath, "checkout", "-b", "main")
	os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# "+name), 0644)
	gitInDir(repoPath, "add", ".")
	gitInDir(repoPath, "commit", "-m", "init")
	return repoPath
}

func TestE2EFullScanPipeline(t *testing.T) {
	tmp := t.TempDir()

	// Create 3 repos: 2 in a subdir (for grouping), 1 at top level
	subdir := filepath.Join(tmp, "mygroup")
	os.MkdirAll(subdir, 0755)

	repo1 := setupTestRepo(t, subdir, "repo1")
	repo2 := setupTestRepo(t, subdir, "repo2")
	repo3 := setupTestRepo(t, tmp, "standalone")

	// Add a worktree to repo1
	gitInDir(repo1, "worktree", "add", filepath.Join(tmp, "mygroup", "repo1-feat"), "-b", "feat-branch")

	// Dirty a file in repo2
	os.WriteFile(filepath.Join(repo2, "dirty.txt"), []byte("changed"), 0644)

	// Build config
	cfg := Config{
		Repos: []Repo{
			{Path: repo1, Group: "mygroup"},
			{Path: repo2, Group: "mygroup"},
			{Path: repo3, Group: "default"},
		},
		Interval: 10,
	}

	// Run scan
	result := RunScan(cfg)

	// Verify repo1 has worktrees (main + feat-branch)
	var repo1Result *RepoResult
	for _, repos := range result.Groups {
		for i := range repos {
			if repos[i].Name == "repo1" {
				repo1Result = &repos[i]
			}
		}
	}
	if repo1Result == nil {
		t.Fatal("repo1 not found in results")
	}
	if len(repo1Result.Worktrees) < 2 {
		t.Errorf("repo1 worktrees = %d, want >= 2", len(repo1Result.Worktrees))
	}

	// Verify repo2 shows modified count
	var repo2Result *RepoResult
	for _, repos := range result.Groups {
		for i := range repos {
			if repos[i].Name == "repo2" {
				repo2Result = &repos[i]
			}
		}
	}
	if repo2Result == nil {
		t.Fatal("repo2 not found in results")
	}
	if len(repo2Result.Worktrees) == 0 {
		t.Fatal("repo2 should have at least 1 worktree (main)")
	}
	if repo2Result.Worktrees[0].Status.Untracked < 1 {
		t.Errorf("repo2 untracked = %d, want >= 1", repo2Result.Worktrees[0].Status.Untracked)
	}

	// Verify standalone (clean, no extra worktrees) is collapsible
	var standaloneResult *RepoResult
	for _, repos := range result.Groups {
		for i := range repos {
			if repos[i].Name == "standalone" {
				standaloneResult = &repos[i]
			}
		}
	}
	if standaloneResult == nil {
		t.Fatal("standalone not found in results")
	}
	if !isCleanNoWorktrees(*standaloneResult) {
		t.Error("standalone should be clean with no extra worktrees")
	}

	// Verify render output
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()

	if !strings.Contains(out, "mygroup") {
		t.Error("render should contain group header 'mygroup'")
	}
	if !strings.Contains(out, "1 repos clean") {
		t.Errorf("render should show 1 clean repo, got:\n%s", out)
	}
	if strings.Contains(out, "standalone") {
		t.Error("standalone should be collapsed, not shown individually")
	}
}

func TestE2EInitCommand(t *testing.T) {
	tmp := t.TempDir()

	// Create repos in subdirs
	subdir := filepath.Join(tmp, "infra")
	os.MkdirAll(subdir, 0755)
	setupTestRepo(t, subdir, "terraform")
	setupTestRepo(t, subdir, "ansible")
	setupTestRepo(t, tmp, "toplevel")

	// Also create a worktree (should be skipped by init)
	topRepo := filepath.Join(tmp, "toplevel")
	gitInDir(topRepo, "worktree", "add", filepath.Join(tmp, "toplevel-wt"), "-b", "wt-branch")

	// Run discover
	repos, err := discoverRepos(tmp)
	if err != nil {
		t.Fatalf("discoverRepos: %v", err)
	}

	// Should find 3 repos, not the worktree
	if len(repos) != 3 {
		names := make([]string, len(repos))
		for i, r := range repos {
			names[i] = r.Path
		}
		t.Fatalf("found %d repos (want 3): %v", len(repos), names)
	}

	// Verify grouping
	groupMap := make(map[string]int)
	for _, r := range repos {
		groupMap[r.Group]++
	}
	if groupMap["infra"] != 2 {
		t.Errorf("infra group = %d repos, want 2", groupMap["infra"])
	}
	if groupMap["default"] != 1 {
		t.Errorf("default group = %d repos, want 1", groupMap["default"])
	}
}
