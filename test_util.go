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
	"bytes"
	"fmt"
	"io"
	"path"
)

type testLogger struct {
	output bool
	log    bytes.Buffer
}

func (t testLogger) SetPrefix(string) {}

func (t testLogger) GetPrefix(bool) string {
	return "test logger"
}

func (t testLogger) Sub() ILogger {
	return t
}

func (t testLogger) Info(format string, args ...interface{}) {
	toWrite := fmt.Sprintf("[INFO] "+format+"\n", args)
	if t.output {
		fmt.Println(toWrite)
	}
	i, _ := t.log.WriteString(toWrite)
	if i != len(toWrite) {
		panic("Bad write")
	}
}

func (t testLogger) Debug(format string, args ...interface{}) {
	toWrite := fmt.Sprintf("[DEBUG] "+format+"\n", args)
	if t.output {
		fmt.Println(toWrite)
	}
	i, _ := t.log.WriteString(toWrite)
	if i != len(toWrite) {
		panic("Bad write")
	}
}

func (t testLogger) Warn(format string, args ...interface{}) {
	toWrite := fmt.Sprintf("[WARN] "+format+"\n", args)
	if t.output {
		fmt.Println(toWrite)
	}
	i, _ := t.log.WriteString(toWrite)
	if i != len(toWrite) {
		panic("Bad write")
	}
}

func (t testLogger) Fatal(format string, args ...interface{}) {
	toWrite := fmt.Sprintf("[FATAL] "+format+"\n", args)
	if t.output {
		fmt.Println(toWrite)
	}
	i, _ := t.log.WriteString(toWrite)
	if i != len(toWrite) {
		panic("Bad write")
	}
}

func (t testLogger) HandleErr(msg interface{}) {
	var toWrite string

	switch msg := msg.(type) {
	case string:
		toWrite = fmt.Sprintln(msg)
	case error:
		toWrite = fmt.Sprintln(msg.Error())
	}

	i, _ := t.log.WriteString(toWrite)
	if i != len(toWrite) {
		panic("Bad Write")
	}
}

func (t testLogger) CloseQuietly(c io.Closer) {
	_ = c.Close()
}

func (t testLogger) String() string {
	str := t.log.String()
	t.log.Reset()
	return str
}

func load() *Config {
	p := path.Join("test_resources", "config.json")
	config, err := LoadConfig(p)

	if err != nil {
		panic(fmt.Sprintf("Error when loading config json: %s", err))
	}

	return config

}

func connect(c *Config) *DB {
	conn, err := NewDB(c)

	if err != nil {
		panic(fmt.Sprintf("Error when connecting to DB: %v", err))
	}

	return conn
}

func dbTestSetup() (*Config, *DB) {
	config := load()
	db := connect(config)

	return config, db
}

func withTestFixtures(f func(*Config, *DB, testLogger)) func() {
	return func() {
		config, db := dbTestSetup()
		logger := testLogger{}

		f(config, db, logger)
	}
}
