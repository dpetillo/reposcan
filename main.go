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
	noColor := flag.Bool("no-color", false, "disable colored output")
	flag.Parse()

	if *showVersion {
		fmt.Println("reposcan", version)
		return
	}

	if os.Getenv("NO_COLOR") != "" {
		*noColor = true
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for {
		result := RunScan(cfg)
		Render(result, os.Stdout, *noColor)
		if *once {
			break
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}
