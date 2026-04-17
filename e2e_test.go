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

	// Create group dirs with repos
	coreDir := filepath.Join(tmp, "core")
	testDir := filepath.Join(tmp, "test")
	os.MkdirAll(coreDir, 0755)
	os.MkdirAll(testDir, 0755)

	repo1 := setupTestRepo(t, coreDir, "repo1")
	setupTestRepo(t, coreDir, "repo2")
	setupTestRepo(t, testDir, "test-repo")

	// Add a worktree to repo1 (sibling in the group dir)
	gitInDir(repo1, "worktree", "add", filepath.Join(coreDir, "repo1-feat"), "-b", "feat-branch")

	// Dirty a file in repo2
	os.WriteFile(filepath.Join(coreDir, "repo2", "dirty.txt"), []byte("changed"), 0644)

	cfg := Config{
		Groups: []Group{
			{Name: "core", Path: coreDir},
			{Name: "test", Path: testDir},
		},
		Interval: 10,
	}

	result := RunScan(cfg)

	// Verify core group has repos
	coreRepos := result.Groups["core"]
	if len(coreRepos) == 0 {
		t.Fatal("core group has no repos")
	}

	// Find repo1 — should have worktrees
	var repo1Result *RepoResult
	for i := range coreRepos {
		if coreRepos[i].Name == "repo1" {
			repo1Result = &coreRepos[i]
		}
	}
	if repo1Result == nil {
		t.Fatal("repo1 not found in core group")
	}
	if len(repo1Result.Worktrees) < 2 {
		t.Errorf("repo1 worktrees = %d, want >= 2", len(repo1Result.Worktrees))
	}

	// Find repo2 — should show dirty status
	var repo2Result *RepoResult
	for i := range coreRepos {
		if coreRepos[i].Name == "repo2" {
			repo2Result = &coreRepos[i]
		}
	}
	if repo2Result == nil {
		t.Fatal("repo2 not found in core group")
	}
	if len(repo2Result.Worktrees) == 0 || repo2Result.Worktrees[0].Status.Untracked < 1 {
		t.Error("repo2 should have untracked files")
	}

	// test-repo should be clean with no extra worktrees
	testRepos := result.Groups["test"]
	if len(testRepos) != 1 {
		t.Fatalf("test group repos = %d, want 1", len(testRepos))
	}
	if !isCleanNoWorktrees(testRepos[0]) {
		t.Error("test-repo should be clean with no extra worktrees")
	}

	// Verify render
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "[core]") {
		t.Error("missing core group label")
	}
	if !strings.Contains(out, "test:1") {
		t.Errorf("should show test:1 in clean summary, got:\n%s", out)
	}
}

func TestE2EInitLocal(t *testing.T) {
	tmp := t.TempDir()

	// Create group dirs with repos
	infraDir := filepath.Join(tmp, "infra")
	os.MkdirAll(infraDir, 0755)
	setupTestRepo(t, infraDir, "terraform")
	setupTestRepo(t, infraDir, "ansible")

	// Root-level repo
	setupTestRepo(t, tmp, "toplevel")

	// Worktree at root (should not be treated as a group)
	gitInDir(filepath.Join(tmp, "toplevel"), "worktree", "add", filepath.Join(tmp, "toplevel-wt"), "-b", "wt-branch")

	// Save cwd and change to a temp output dir
	outDir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(orig)

	err := runInit(tmp, "")
	if err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Read generated config
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("reading config.yaml: %v", err)
	}
	cfg, err := parseConfig(data)
	if err != nil {
		t.Fatalf("parsing generated config: %v", err)
	}

	// Should have 2 groups: infra + (root)
	if len(cfg.Groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(cfg.Groups))
	}

	groupNames := make(map[string]bool)
	for _, g := range cfg.Groups {
		groupNames[g.Name] = true
	}
	if !groupNames["infra"] {
		t.Error("missing infra group")
	}
	if !groupNames["(root)"] {
		t.Error("missing (root) group")
	}
}
