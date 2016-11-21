package main

import (
	"path"
	"reflect"
	"sort"
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

func TestLoading(t *testing.T) {
	config := load(t)

	if config.BindPort != 192048 {
		t.Errorf("Expected config.BindPort to be 192048, was %v", config.BindPort)
	}

	if len(config.Versions) != 3 {
		t.Errorf("Expected there to be 3 Versions, was %v", len(config.Versions))
	}

	if config.BucketName != "test-bucket" {
		t.Errorf("Expected config.BucketName to be test-bucket, was %v", config.BucketName)
	}
}

func TestBindAddr(t *testing.T) {
	config := load(t)

	addr := config.BindAddr()
	if addr != ":192048" {
		t.Errorf("Expected config.BindAddr() to be \":192048\", was %v", addr)
	}
}

func TestVersionNames(t *testing.T) {
	config := load(t)

	expected := []string{"thumb", "thumb_watermarked", "gallery_thumb"}
	names := config.VersionNames()

	contains := func(name string) bool {
		for _, n := range names {
			if name == n {
				return true
			}
		}

		return false
	}

	for _, str := range expected {
		if !contains(str) {
			t.Errorf("Expected %v to contain %s", names, str)
		}
	}
}

func TestGetVersionsByName(t *testing.T) {
	config := load(t)

	byName := config.getVersionsByName()

	keys := make([]string, len(byName))
	i := 0
	for k := range byName {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	names := config.VersionNames()
	sort.Strings(names)

	if !reflect.DeepEqual(keys, names) {
		t.Errorf("Expected %v to equal %v", keys, config.VersionNames())
	}
}
