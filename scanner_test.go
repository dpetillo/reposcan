package main

import "testing"

func TestParseWorktreeListSingle(t *testing.T) {
	data := []byte("worktree /home/user/repo\nbranch refs/heads/main\n\n")
	wts := parseWorktreeList(data)
	if len(wts) != 1 {
		t.Fatalf("got %d worktrees, want 1", len(wts))
	}
	if wts[0].Path != "/home/user/repo" {
		t.Errorf("path = %q, want /home/user/repo", wts[0].Path)
	}
	if wts[0].Branch != "main" {
		t.Errorf("branch = %q, want main", wts[0].Branch)
	}
	if !wts[0].IsMain {
		t.Error("expected IsMain = true for first worktree")
	}
}

func TestParseWorktreeListMultiple(t *testing.T) {
	data := []byte(`worktree /home/user/repo
branch refs/heads/main

worktree /home/user/repo-feature-xyz
branch refs/heads/feature-xyz

worktree /home/user/repo-hotfix
branch refs/heads/hotfix-123

`)
	wts := parseWorktreeList(data)
	if len(wts) != 3 {
		t.Fatalf("got %d worktrees, want 3", len(wts))
	}
	if !wts[0].IsMain {
		t.Error("first worktree should be IsMain")
	}
	if wts[1].IsMain || wts[2].IsMain {
		t.Error("non-first worktrees should not be IsMain")
	}
	if wts[1].Branch != "feature-xyz" {
		t.Errorf("wts[1].branch = %q, want feature-xyz", wts[1].Branch)
	}
	if wts[2].Path != "/home/user/repo-hotfix" {
		t.Errorf("wts[2].path = %q, want /home/user/repo-hotfix", wts[2].Path)
	}
}

func TestParseWorktreeListDetachedHead(t *testing.T) {
	data := []byte(`worktree /home/user/repo
branch refs/heads/main

worktree /home/user/repo-detached
detached

`)
	wts := parseWorktreeList(data)
	if len(wts) != 2 {
		t.Fatalf("got %d worktrees, want 2", len(wts))
	}
	if wts[1].Branch != "(detached)" {
		t.Errorf("branch = %q, want (detached)", wts[1].Branch)
	}
}

func TestParseWorktreeListNoTrailingNewline(t *testing.T) {
	data := []byte("worktree /home/user/repo\nbranch refs/heads/main")
	wts := parseWorktreeList(data)
	if len(wts) != 1 {
		t.Fatalf("got %d worktrees, want 1", len(wts))
	}
	if wts[0].Branch != "main" {
		t.Errorf("branch = %q, want main", wts[0].Branch)
	}
}

func TestParseWorktreeListEmpty(t *testing.T) {
	wts := parseWorktreeList([]byte(""))
	if len(wts) != 0 {
		t.Fatalf("got %d worktrees, want 0", len(wts))
	}
}
