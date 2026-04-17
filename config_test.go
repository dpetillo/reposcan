package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseValidConfig(t *testing.T) {
	yaml := []byte(`
interval: 5
repos:
  - path: /home/user/repo1
    group: core
  - path: /home/user/repo2
    group: test
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 5 {
		t.Errorf("interval = %d, want 5", cfg.Interval)
	}
	if len(cfg.Repos) != 2 {
		t.Fatalf("repos count = %d, want 2", len(cfg.Repos))
	}
	if cfg.Repos[0].Path != "/home/user/repo1" {
		t.Errorf("repos[0].path = %q, want /home/user/repo1", cfg.Repos[0].Path)
	}
	if cfg.Repos[1].Group != "test" {
		t.Errorf("repos[1].group = %q, want test", cfg.Repos[1].Group)
	}
}

func TestParseConfigDefaultInterval(t *testing.T) {
	yaml := []byte(`
repos:
  - path: /tmp/repo
    group: default
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != defaultInterval {
		t.Errorf("interval = %d, want %d", cfg.Interval, defaultInterval)
	}
}

func TestParseConfigZeroInterval(t *testing.T) {
	yaml := []byte(`
interval: 0
repos:
  - path: /tmp/repo
    group: default
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != defaultInterval {
		t.Errorf("interval = %d, want %d (should default when 0)", cfg.Interval, defaultInterval)
	}
}

func TestParseConfigTildeExpansion(t *testing.T) {
	yaml := []byte(`
repos:
  - path: ~/Dev/myrepo
    group: core
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "Dev/myrepo")
	if cfg.Repos[0].Path != want {
		t.Errorf("path = %q, want %q", cfg.Repos[0].Path, want)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Use a path that doesn't exist
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseConfigInvalidYAML(t *testing.T) {
	yaml := []byte(`{{{invalid`)
	_, err := parseConfig(yaml)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfigFromCwd(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(configFile, []byte(`
repos:
  - path: /tmp/r1
    group: g1
  - path: /tmp/r2
    group: g2
`), 0644)

	// Change to temp dir so findConfig picks up config.yaml
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Repos) != 2 {
		t.Errorf("repos count = %d, want 2", len(cfg.Repos))
	}
}

func TestExpandHomeNoTilde(t *testing.T) {
	path := "/absolute/path"
	if got := expandHome(path); got != path {
		t.Errorf("expandHome(%q) = %q, want %q", path, got, path)
	}
}

func TestExpandHomeEmpty(t *testing.T) {
	if got := expandHome(""); got != "" {
		t.Errorf("expandHome(\"\") = %q, want \"\"", got)
	}
}
