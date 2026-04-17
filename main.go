package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var version = "dev"

func main() {
	once := flag.Bool("once", false, "scan once and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	noColor := flag.Bool("no-color", false, "disable colored output")
	interval := flag.Int("interval", defaultInterval, "refresh interval in seconds")
	flag.Parse()

	if *showVersion {
		fmt.Println("reposcan", version)
		return
	}

	if os.Getenv("NO_COLOR") != "" {
		*noColor = true
	}

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	for {
		result := RunScan(root)
		Render(result, os.Stdout, *noColor)
		if *once {
			break
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}
