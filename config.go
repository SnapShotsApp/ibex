package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type versionProperties map[string]map[string]interface{}

// BucketNames maps rails environment names to their S3 buckets
var BucketNames = map[string]string{
	"development": "snapshots-photos-dev",
	"staging":     "snapshots-photos-staging",
	"production":  "heysnapshots-photos",
}

// Version contains a single picture version from config
type Version struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

// Config loads and contains configs from the json file
type Config struct {
	DatabaseURL    string    `json:"database_url"`
	BindPort       int       `json:"bind_port"`
	Versions       []Version `json:"versions"`
	ImagizerHost   string    `json:"imagizer_host"`
	CDNHost        string    `json:"cdn_host"`
	versionsByName versionProperties
}

// LoadConfig loads the config file from the given path
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	config.versionsByName = config.getVersionsByName()

	return &config, nil
}

// BindAddr returns a http.Server compatible addr
func (c *Config) BindAddr() string {
	return fmt.Sprintf(":%d", c.BindPort)
}

// VersionNames maps the contained versions' names
func (c *Config) VersionNames() []string {
	names := make([]string, len(c.Versions))

	for i, v := range c.Versions {
		names[i] = v.Name
	}

	return names
}

func (c Config) getVersionsByName() versionProperties {
	mp := make(versionProperties)

	for _, v := range c.Versions {
		mp[v.Name] = v.Properties
	}

	return mp
}
