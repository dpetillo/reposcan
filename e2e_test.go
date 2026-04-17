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

func TestE2EFullScan(t *testing.T) {
	tmp := t.TempDir()

	// Group dirs with repos
	coreDir := filepath.Join(tmp, "core")
	os.MkdirAll(coreDir, 0755)
	repo1 := setupTestRepo(t, coreDir, "repo1")
	setupTestRepo(t, coreDir, "repo2")

	// Root-level repo
	setupTestRepo(t, tmp, "standalone")

	// Add worktree to repo1
	gitInDir(repo1, "worktree", "add", filepath.Join(coreDir, "repo1-feat"), "-b", "feat-branch")

	// Dirty repo2
	os.WriteFile(filepath.Join(coreDir, "repo2", "dirty.txt"), []byte("x"), 0644)

	result := RunScan(tmp)

	// repo1 should have worktrees
	var found bool
	for _, repos := range result.Groups {
		for _, r := range repos {
			if r.Name == "repo1" {
				found = true
				if len(r.Worktrees) < 2 {
					t.Errorf("repo1 worktrees = %d, want >= 2", len(r.Worktrees))
				}
			}
		}
	}
	if !found {
		t.Fatal("repo1 not found")
	}

	// standalone should be clean
	for _, repos := range result.Groups {
		for _, r := range repos {
			if r.Name == "standalone" && !isFullyClean(r) {
				t.Error("standalone should be clean")
			}
		}
	}

	// Render should show group labels
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "[core]") {
		t.Errorf("missing [core] group label in:\n%s", out)
	}
}
