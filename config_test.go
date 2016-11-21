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
	"reflect"
	"sort"
	"testing"
)

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
