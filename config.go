package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Group struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Config struct {
	GitlabGroup string  `yaml:"gitlab_group,omitempty"`
	Groups      []Group `yaml:"groups"`
	Interval    int     `yaml:"interval"`
}

const defaultInterval = 10

func LoadConfig(path string) (Config, error) {
	if path == "" {
		path = findConfig()
	}
	if path == "" {
		return Config{}, fmt.Errorf("no config file found")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	return parseConfig(data)
}

func parseConfig(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Interval <= 0 {
		cfg.Interval = defaultInterval
	}
	for i := range cfg.Groups {
		cfg.Groups[i].Path = expandHome(cfg.Groups[i].Path)
	}
	return cfg, nil
}

func findConfig() string {
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	fallback := filepath.Join(home, ".config", "reposcan", "config.yaml")
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	return ""
}

func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
