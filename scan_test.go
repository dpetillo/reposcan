package main

import "testing"

func TestGroupResults(t *testing.T) {
	results := []RepoResult{
		{Name: "repo1", Group: "core"},
		{Name: "repo2", Group: "core"},
		{Name: "repo3", Group: "test"},
		{Name: "repo4", Group: ""},
	}
	groups := groupResults(results)
	if len(groups["core"]) != 2 {
		t.Errorf("core group = %d repos, want 2", len(groups["core"]))
	}
	if len(groups["test"]) != 1 {
		t.Errorf("test group = %d repos, want 1", len(groups["test"]))
	}
	if len(groups["default"]) != 1 {
		t.Errorf("default group = %d repos, want 1", len(groups["default"]))
	}
}

func TestGroupResultsEmpty(t *testing.T) {
	groups := groupResults(nil)
	if len(groups) != 0 {
		t.Errorf("expected empty groups, got %d", len(groups))
	}
}

func TestGroupResultsSingleGroup(t *testing.T) {
	results := []RepoResult{
		{Name: "a", Group: "infra"},
		{Name: "b", Group: "infra"},
	}
	groups := groupResults(results)
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if len(groups["infra"]) != 2 {
		t.Errorf("infra group = %d repos, want 2", len(groups["infra"]))
	}
}
