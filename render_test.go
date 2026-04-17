package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderGroupHeaders(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"core": {{Name: "repo1", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "feat-x"}, Status: WorktreeStatus{Modified: 1}},
			}}},
			"test": {{Name: "repo2", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{Modified: 2}},
			}}},
		},
		Duration: 100 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "═══ core ═══") {
		t.Error("missing core group header")
	}
	if !strings.Contains(out, "═══ test ═══") {
		t.Error("missing test group header")
	}
}

func TestRenderCleanReposCollapsed(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"infra": {
				{Name: "clean1", Worktrees: []WorktreeResult{
					{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{}},
				}},
				{Name: "clean2", Worktrees: []WorktreeResult{
					{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{}},
				}},
			},
		},
		Duration: 50 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if strings.Contains(out, "clean1") {
		t.Error("clean repo 'clean1' should not appear individually")
	}
	if !strings.Contains(out, "2 repos clean (no worktrees)") {
		t.Errorf("missing clean summary, got:\n%s", out)
	}
}

func TestRenderDirtyRepoShowsStatus(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"core": {{Name: "dirty-repo", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "feat-abc"}, Status: WorktreeStatus{Modified: 3, Staged: 1, Untracked: 2, Ahead: 5, Behind: 1}},
			}}},
		},
		Duration: 10 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "dirty-repo") {
		t.Error("dirty repo should appear")
	}
	if !strings.Contains(out, "M:3") {
		t.Error("missing modified count")
	}
	if !strings.Contains(out, "S:1") {
		t.Error("missing staged count")
	}
	if !strings.Contains(out, "U:2") {
		t.Error("missing untracked count")
	}
	if !strings.Contains(out, "+5") {
		t.Error("missing ahead count")
	}
	if !strings.Contains(out, "-1") {
		t.Error("missing behind count")
	}
}

func TestRenderWorktreesIndented(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"core": {{Name: "myrepo", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{Modified: 1}},
				{Worktree: Worktree{Branch: "feat-x"}, Status: WorktreeStatus{}},
			}}},
		},
		Duration: 10 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "myrepo") {
		t.Error("repo name should appear")
	}
	if !strings.Contains(out, "├─") {
		t.Error("worktrees should be indented with tree chars")
	}
	if !strings.Contains(out, "feat-x") {
		t.Error("worktree branch should appear")
	}
}

func TestRenderCleanWorktreeShowsCheckmark(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"g": {{Name: "r", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{}},
				{Worktree: Worktree{Branch: "feat"}, Status: WorktreeStatus{}},
			}}},
		},
		Duration: 10 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "✓ clean") {
		t.Error("clean worktree should show checkmark")
	}
}

func TestRenderScanDuration(t *testing.T) {
	result := ScanResult{
		Groups:   map[string][]RepoResult{},
		Duration: 250 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if !strings.Contains(out, "250ms") {
		t.Errorf("should show scan duration, got:\n%s", out)
	}
}
