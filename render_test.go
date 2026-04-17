package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderGroupLabels(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"core": {{Name: "repo1", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{Modified: 1}},
			}}},
		},
		Duration: 100 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	if !strings.Contains(buf.String(), "[core]") {
		t.Error("missing group label [core]")
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
	if strings.Contains(out, "clean1") || strings.Contains(out, "clean2") {
		t.Error("clean repos should not appear individually")
	}
	if !strings.Contains(out, "infra:2") {
		t.Errorf("missing clean summary, got:\n%s", out)
	}
}

func TestRenderDirtyRepoShowsStatus(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"core": {{Name: "dirty-repo", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{Modified: 3, Ahead: 2}},
				{Worktree: Worktree{Branch: "feature-abc"}, Status: WorktreeStatus{Staged: 1, Behind: 1}},
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
	if !strings.Contains(out, "M3") {
		t.Error("missing modified count")
	}
	if !strings.Contains(out, "↑2") {
		t.Error("missing ahead arrow")
	}
	if !strings.Contains(out, "S1") {
		t.Error("missing staged count on worktree")
	}
	if !strings.Contains(out, "↓1") {
		t.Error("missing behind arrow on worktree")
	}
}

func TestRenderCleanWorktreesHidden(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"g": {{Name: "r", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{Modified: 1}},
				{Worktree: Worktree{Branch: "feature-clean-one"}, Status: WorktreeStatus{}},
			}}},
		},
		Duration: 10 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if strings.Contains(out, "clean-one") {
		t.Error("clean worktree should be hidden")
	}
}

func TestRenderShortBranchNames(t *testing.T) {
	result := ScanResult{
		Groups: map[string][]RepoResult{
			"g": {{Name: "r", Worktrees: []WorktreeResult{
				{Worktree: Worktree{Branch: "main", IsMain: true}, Status: WorktreeStatus{}},
				{Worktree: Worktree{Branch: "feature-CORE-1216-fix"}, Status: WorktreeStatus{Modified: 1}},
			}}},
		},
		Duration: 10 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	out := buf.String()
	if strings.Contains(out, "feature-CORE") {
		t.Error("should strip feature- prefix")
	}
	if !strings.Contains(out, "CORE-1216-fix") {
		t.Error("should show shortened branch name")
	}
}

func TestRenderScanDuration(t *testing.T) {
	result := ScanResult{
		Groups:   map[string][]RepoResult{},
		Duration: 250 * time.Millisecond,
	}
	var buf bytes.Buffer
	Render(result, &buf, true)
	if !strings.Contains(buf.String(), "250ms") {
		t.Errorf("should show scan duration, got:\n%s", buf.String())
	}
}

func TestShortBranch(t *testing.T) {
	tests := []struct{ in, want string }{
		{"feature-CORE-123-fix", "CORE-123-fix"},
		{"bugfix-thing", "thing"},
		{"release-7.2.0", "7.2.0"},
		{"main", "main"},
		{"develop", "develop"},
	}
	for _, tt := range tests {
		if got := shortBranch(tt.in); got != tt.want {
			t.Errorf("shortBranch(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
