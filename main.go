package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var version = "dev"

func main() {
	configPath := flag.String("c", "", "path to config file")
	once := flag.Bool("once", false, "scan once and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("reposcan", version)
		return
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d repos (interval: %ds)\n", len(cfg.Repos), cfg.Interval)

	for {
		result := RunScan(cfg)
		fmt.Printf("\nScan complete: %d groups, %s\n", len(result.Groups), result.Duration.Round(time.Millisecond))
		for group, repos := range result.Groups {
			fmt.Printf("  [%s] %d repos\n", group, len(repos))
			for _, repo := range repos {
				fmt.Printf("    %s (%d worktrees)\n", repo.Name, len(repo.Worktrees))
				for _, wt := range repo.Worktrees {
					fmt.Printf("      %s: M:%d S:%d U:%d +%d -%d\n",
						wt.Branch, wt.Status.Modified, wt.Status.Staged,
						wt.Status.Untracked, wt.Status.Ahead, wt.Status.Behind)
				}
			}
		}
		if *once {
			break
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}
