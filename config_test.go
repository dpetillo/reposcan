package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseValidConfig(t *testing.T) {
	yaml := []byte(`
interval: 5
groups:
  - name: core
    path: /home/user/dev/core
  - name: test
    path: /home/user/dev/test
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 5 {
		t.Errorf("interval = %d, want 5", cfg.Interval)
	}
	if len(cfg.Groups) != 2 {
		t.Fatalf("groups count = %d, want 2", len(cfg.Groups))
	}
	if cfg.Groups[0].Name != "core" {
		t.Errorf("groups[0].name = %q, want core", cfg.Groups[0].Name)
	}
	if cfg.Groups[1].Path != "/home/user/dev/test" {
		t.Errorf("groups[1].path = %q, want /home/user/dev/test", cfg.Groups[1].Path)
	}
}

func TestParseConfigDefaultInterval(t *testing.T) {
	yaml := []byte(`
groups:
  - name: default
    path: /tmp/repos
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
groups:
  - name: default
    path: /tmp/repos
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != defaultInterval {
		t.Errorf("interval = %d, want %d", cfg.Interval, defaultInterval)
	}
}

func TestParseConfigTildeExpansion(t *testing.T) {
	yaml := []byte(`
groups:
  - name: core
    path: ~/Dev/core
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "Dev/core")
	if cfg.Groups[0].Path != want {
		t.Errorf("path = %q, want %q", cfg.Groups[0].Path, want)
	}
}

func TestParseConfigGitlabGroup(t *testing.T) {
	yaml := []byte(`
gitlab_group: directbook1
groups:
  - name: core
    path: /tmp/core
`)
	cfg, err := parseConfig(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GitlabGroup != "directbook1" {
		t.Errorf("gitlab_group = %q, want directbook1", cfg.GitlabGroup)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
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
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
groups:
  - name: g1
    path: /tmp/g1
  - name: g2
    path: /tmp/g2
`), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Groups) != 2 {
		t.Errorf("groups count = %d, want 2", len(cfg.Groups))
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
