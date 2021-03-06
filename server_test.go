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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func withImagizerTestServer(handler http.HandlerFunc, f func(*httptest.Server)) func() {
	return func() {
		server := httptest.NewServer(handler)
		defer server.Close()
		f(server)
	}
}

func TestPathMatching(t *testing.T) {
	Convey("Server process", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, r.URL.Path)
		})

		Convey("Handling path recognition", withImagizerTestServer(hf, func(server *httptest.Server) {
			imagizerHost, _ := url.Parse(server.URL)
			handler := imagizerHandler{imagizerHost, config, db, logger, NewBlackHole(), 1 * time.Second}

			badReqs := []*http.Request{
				httptest.NewRequest("GET", "/foo", nil),
				httptest.NewRequest("GET", "/uploads/staging/jlindsey/picture/attachment/1/thumb", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/4/thumb", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/1/thumb/3/foo+", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/2/large", nil),
				httptest.NewRequest("DELETE", "/uploads/staging/picture/attachment/1/thumb", nil),
				httptest.NewRequest("POST", "/uploads/staging/picture/attachment/1/thumb", nil),
				httptest.NewRequest("PUT", "/uploads/staging/picture/attachment/1/thumb", nil),
				httptest.NewRequest("PATCH", "/uploads/staging/picture/attachment/1/thumb", nil),
				httptest.NewRequest("OPTIONS", "/uploads/staging/picture/attachment/1/thumb", nil),
				httptest.NewRequest("HEAD", "/uploads/staging/picture/attachment/1/thumb", nil),
			}

			goodReqs := []*http.Request{
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/1/thumb/3", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/2/gallery_thumb/abc123", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/2/gallery_thumb/abc-123", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/2/gallery_thumb/abc_123-", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/2/thumb", nil),
				httptest.NewRequest("GET", "/uploads/staging/picture/attachment/2/gallery_thumb", nil),
			}

			for _, req := range badReqs {
				Convey(fmt.Sprintf("%s %s should return 404", req.Method, req.URL.String()), func() {
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)
					So(w.Code, ShouldEqual, 404)
				})
			}

			for _, req := range goodReqs {
				Convey(fmt.Sprintf("%s %s should return 200", req.Method, req.URL.String()), func() {
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)
					So(w.Code, ShouldEqual, 200)
				})
			}
		}))
	}))
}

func TestConnectionTimeouts(t *testing.T) {
	Convey("Server Timeouts", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(70 * time.Millisecond)
			fmt.Fprintf(w, "Hello")
		})

		Convey("Do not hang forever", withImagizerTestServer(hf, func(server *httptest.Server) {
			imagizerHost, _ := url.Parse(server.URL)

			handler := imagizerHandler{imagizerHost, config, db, logger, NewBlackHole(), 50 * time.Millisecond}
			req := httptest.NewRequest("GET", "/uploads/staging/picture/attachment/1/thumb/3", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			expectedCode := (w.Code == 504 || w.Code == 500)
			So(expectedCode, ShouldBeTrue)
		}))
	}))
}
