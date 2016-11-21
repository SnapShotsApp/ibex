/* Copyright 2016 Snapshots LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
