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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStatsServer(t *testing.T) {
	Convey("StatsServer", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		stats := NewStats(logger)
		go stats.Listen()

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/stats", nil)
		So(err, ShouldBeNil)
		stats.ServeHTTP(w, req)

		body := w.Body.String()
		So(body, ShouldContainSubstring, `"total_served":0`)

		stats.statsChan <- &stat{StatServedPicture, "thumb"}
		time.Sleep(5 * time.Millisecond) // Give the goroutine time to process the chan

		newW := httptest.NewRecorder()
		newReq, err := http.NewRequest("GET", "/stats", nil)
		So(err, ShouldBeNil)
		stats.ServeHTTP(newW, newReq)

		newBody := newW.Body.String()
		So(newBody, ShouldContainSubstring, `"total_served":1`)
		So(newBody, ShouldContainSubstring, `"thumb":1`)
	}))
}

func TestConfigServer(t *testing.T) {
	Convey("ConfigServer", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		handler := configHandler{config, logger}

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/config", nil)
		So(err, ShouldBeNil)
		handler.ServeHTTP(w, req)

		body := w.Body.String()
		So(body, ShouldContainSubstring, `"bind_port":192048`)
	}))
}
