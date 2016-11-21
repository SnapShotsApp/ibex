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
