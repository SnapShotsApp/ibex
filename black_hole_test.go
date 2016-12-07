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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func shouldNotTimeOutIn(actual interface{}, expected ...interface{}) string {
	fail := make(chan bool)
	out := make(chan bool)
	duration := expected[0].(time.Duration)
	timer := time.AfterFunc(duration, func() { fail <- true })
	go func() {
		actual.(func())()
		timer.Stop()
		out <- true
	}()

	select {
	case <-fail:
		return fmt.Sprintf("Expected function to return within %v", expected)
	case <-out:
		return ""
	}
}

func TestBlackHole(t *testing.T) {
	Convey("BlackHole chan accepts any number of things without blocking", t, func() {
		bh := NewBlackHole()
		f := func() {
			for i := 0; i <= 50; i++ {
				bh <- &stat{}
			}
		}

		So(f, shouldNotTimeOutIn, 1*time.Second)
	})
}
