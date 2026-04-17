package main

import (
	"flag"
	"fmt"
	"os"
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

	_ = *once // will be used in later tasks for scan loop
}
