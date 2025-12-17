package main

import (
	"flag"
	"log"

	"linkedin-automation/cmd"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config yaml")
	flag.Parse()
	command := "start"
	if len(flag.Args()) > 0 {
		command = flag.Args()[0]
	}

	if err := cmd.Run(*cfgPath, command); err != nil {
		log.Fatalf("bot failed: %v", err)
	}
}

