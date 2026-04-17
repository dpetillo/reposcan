package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type glabSubgroup struct {
	FullPath string `json:"full_path"`
	Name     string `json:"name"`
}

type glabProject struct {
	Path      string `json:"path"`
	Namespace struct {
		FullPath string `json:"full_path"`
	} `json:"namespace"`
}

func runInit(targetDir string, gitlabGroup string) error {
	if targetDir == "" {
		targetDir = "."
	}
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	if gitlabGroup == "" {
		return runInitLocal(targetDir)
	}
	return runInitGitlab(targetDir, gitlabGroup)
}

// runInitGitlab queries GitLab for the group hierarchy and maps to local dirs
func runInitGitlab(targetDir string, gitlabGroup string) error {
	fmt.Printf("Querying GitLab group %q...\n", gitlabGroup)

	// Get subgroups
	subgroups, err := glabGetSubgroups(gitlabGroup)
	if err != nil {
		return fmt.Errorf("fetching subgroups: %w", err)
	}

	// Get all projects
	projects, err := glabGetProjects(gitlabGroup)
	if err != nil {
		return fmt.Errorf("fetching projects: %w", err)
	}

	// Build group→projects map using top-level subgroup
	type groupInfo struct {
		name     string
		fullPath string
	}
	topGroups := make(map[string]groupInfo) // path → info
	for _, sg := range subgroups {
		// Only direct children of the root group
		parts := strings.Split(sg.FullPath, "/")
		if len(parts) == 2 {
			topGroups[strings.ToLower(parts[1])] = groupInfo{name: sg.Name, fullPath: sg.FullPath}
		}
	}

	// Map projects to their top-level group
	groupProjects := make(map[string]int) // group path → project count
	for _, p := range projects {
		parts := strings.Split(p.Namespace.FullPath, "/")
		if len(parts) >= 2 {
			groupProjects[strings.ToLower(parts[1])]++
		} else {
			groupProjects["(root)"]++
		}
	}

	// Build config groups — only include groups that have a matching local directory
	var groups []Group
	entries, _ := os.ReadDir(targetDir)
	localDirs := make(map[string]string) // lowercase → actual name
	for _, e := range entries {
		if e.IsDir() {
			localDirs[strings.ToLower(e.Name())] = e.Name()
		}
	}

	// Match GitLab groups to local directories
	for key, info := range topGroups {
		if dirName, ok := localDirs[key]; ok {
			groups = append(groups, Group{
				Name: info.name,
				Path: filepath.Join(targetDir, dirName),
			})
		}
	}

	// Check for repos at the root level (not in any subgroup)
	hasRootRepos := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gitPath := filepath.Join(targetDir, e.Name(), ".git")
		info, err := os.Stat(gitPath)
		if err != nil || !info.IsDir() {
			continue
		}
		// Check if this dir is already covered by a group
		if _, ok := topGroups[strings.ToLower(e.Name())]; ok {
			continue
		}
		hasRootRepos = true
		break
	}
	if hasRootRepos {
		groups = append(groups, Group{Name: "(root)", Path: targetDir})
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })

	cfg := Config{
		GitlabGroup: gitlabGroup,
		Groups:      groups,
		Interval:    defaultInterval,
	}
	return writeConfig(cfg)
}

// runInitLocal falls back to filesystem-only discovery
func runInitLocal(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	var groups []Group
	hasRootRepos := false

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(targetDir, e.Name())
		gitPath := filepath.Join(dirPath, ".git")

		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			// This is a repo at root level
			hasRootRepos = true
			continue
		}

		// Check if this dir contains repos (is a group dir)
		subEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		hasRepos := false
		for _, se := range subEntries {
			if !se.IsDir() {
				continue
			}
			subGit := filepath.Join(dirPath, se.Name(), ".git")
			si, err := os.Stat(subGit)
			if err == nil && si.IsDir() {
				hasRepos = true
				break
			}
		}
		if hasRepos {
			groups = append(groups, Group{Name: e.Name(), Path: dirPath})
		}
	}

	if hasRootRepos {
		groups = append(groups, Group{Name: "(root)", Path: targetDir})
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })

	cfg := Config{Groups: groups, Interval: defaultInterval}
	return writeConfig(cfg)
}

func writeConfig(cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	outPath := "config.yaml"
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Wrote %s with %d groups\n", outPath, len(cfg.Groups))
	return nil
}

func glabGetSubgroups(group string) ([]glabSubgroup, error) {
	out, err := exec.Command("glab", "api", fmt.Sprintf("groups/%s/descendant_groups?per_page=100", group)).Output()
	if err != nil {
		return nil, err
	}
	var sgs []glabSubgroup
	return sgs, json.Unmarshal(out, &sgs)
}

func glabGetProjects(group string) ([]glabProject, error) {
	var all []glabProject
	for page := 1; page <= 10; page++ {
		out, err := exec.Command("glab", "api",
			fmt.Sprintf("groups/%s/projects?per_page=100&include_subgroups=true&page=%d", group, page)).Output()
		if err != nil {
			return nil, err
		}
		var batch []glabProject
		if err := json.Unmarshal(out, &batch); err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
	}
	return all, nil
}
