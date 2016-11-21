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
	"path"
	"testing"
)

func load(t *testing.T) *Config {
	p := path.Join("test_resources", "config.json")
	config, err := LoadConfig(p)

	if err != nil {
		t.Errorf("Error when loading config json: %s", err)
	}

	return config

}

func connect(c *Config, t *testing.T) *DB {
	conn, err := NewDB(c)

	if err != nil {
		t.Errorf("Error when connecting to DB: %v", err)
	}

	return conn
}

func dbTestSetup(t *testing.T) (*Config, *DB) {
	config := load(t)
	db := connect(config, t)

	return config, db
}
