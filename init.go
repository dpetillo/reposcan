package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func runInit(targetDir string) error {
	if targetDir == "" {
		targetDir = "."
	}
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	repos, err := discoverRepos(targetDir)
	if err != nil {
		return err
	}

	cfg := Config{Repos: repos, Interval: defaultInterval}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	outPath := filepath.Join(".", "config.yaml")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Wrote %s with %d repos\n", outPath, len(repos))
	return nil
}

func discoverRepos(root string) ([]Repo, error) {
	var repos []Repo
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}
		if info.Name() == ".git" && info.IsDir() {
			repoDir := filepath.Dir(path)
			group := inferGroup(root, repoDir)
			repos = append(repos, Repo{Path: repoDir, Group: group})
			return filepath.SkipDir
		}
		// .git file = worktree or submodule — skip, don't treat as repo
		if info.Name() == ".git" && !info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	sort.Slice(repos, func(i, j int) bool { return repos[i].Path < repos[j].Path })
	return repos, err
}

func inferGroup(root, repoDir string) string {
	rel, err := filepath.Rel(root, repoDir)
	if err != nil {
		return "default"
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) <= 1 {
		return "default"
	}
	return parts[0]
}
