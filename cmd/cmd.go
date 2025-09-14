package cmd

import (
	"flag"
	"os"
)

func ParseFlags() string {
	configPath := flag.String("config", "./config.json", "path to config file")
	help := flag.Bool("help", false, "print help and exit")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	return *configPath
}
