package main

import (
	"flag"
)

var (
	debugLogEnabled bool
	configFile      string
)

func init() {
	flag.BoolVar(&debugLogEnabled, "debug", false, "enable verbose debug logging")
}

func main() {
	flag.Parse()

	configFile = flag.Arg(0)
	if len(configFile) == 0 {
		configFile = "/etc/ibex.json"
	}

	Debug("Debug logging enabled")

	config, err := LoadConfig(configFile)
	handleErr(err)

	Info("Loaded config from %s", configFile)
	Info("Found %d versions: %s", len(config.Versions), config.VersionNames())

	Start(config)
}
