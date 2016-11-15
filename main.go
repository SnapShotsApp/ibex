package main

import (
	"log"
	"os"
)

func main() {
	var configFile string

	if len(os.Args) < 2 {
		configFile = "/etc/ibex.json"
	} else {
		configFile = os.Args[1]
	}

	config, err := LoadConfig(configFile)
	handleErr(err)
	log.Printf("Loaded config from %s", configFile)
	log.Printf("Found %d versions: %s", len(config.Versions), config.VersionNames())

	Start(config)
}
