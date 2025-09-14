package cmd

import (
	"flag"
	"fmt"
	"os"
)

const Version = "0.2.0"

func ParseFlags() string {
	configPath := flag.String("config", "./config.json", "path to config file")
	help := flag.Bool("help", false, "print help and exit")
	version := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *version {
		println(Version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	return *configPath
}
