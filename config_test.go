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
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoading(t *testing.T) {
	Convey("Loads config from disk", t, func() {
		config := load()

		So(config.BindPort, ShouldEqual, 192048)
		So(len(config.Versions), ShouldEqual, 3)
		So(config.BucketName, ShouldEqual, "test-bucket")
	})
}

func TestBindAddr(t *testing.T) {
	Convey("BindAddr() output", t, func() {
		config := load()

		addr := config.BindAddr()
		So(addr, ShouldEqual, ":192048")
	})
}

func TestVersionNames(t *testing.T) {
	Convey("Version name extraction", t, func() {
		config := load()

		expected := []string{"thumb", "thumb_watermarked", "gallery_thumb"}
		names := config.VersionNames()

		for _, str := range expected {
			So(names, ShouldContain, str)
		}
	})
}

func TestGetVersionsByName(t *testing.T) {
	Convey("Extacting versions into a map", t, func() {
		config := load()

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

		So(keys, ShouldResemble, names)
	})
}
