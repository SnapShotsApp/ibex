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
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConnectionAndQuery(t *testing.T) {
	Convey("Connection and querying", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		testStruct := struct {
			id, userID, eventID int
			attachment          string
		}{}

		rows, err := db.conn.Query("select * from pictures limit 1")
		if err != nil {
			t.Errorf("Error querying database: %v", err)
			t.FailNow()
		}
		defer logger.CloseQuietly(rows)

		for rows.Next() {
			err = rows.Scan(&testStruct.id, &testStruct.userID, &testStruct.eventID, &testStruct.attachment)
			So(err, ShouldBeNil)
		}

		cases := map[interface{}]interface{}{
			testStruct.id:         1,
			testStruct.userID:     1,
			testStruct.eventID:    1,
			testStruct.attachment: "test_pic.jpg",
		}

		for got, expected := range cases {
			So(got, ShouldEqual, expected)
		}
	}))
}

func TestLoadEvent(t *testing.T) {
	Convey("Loading events", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		ev1 := db.loadEvent(1, logger)
		ev2 := db.loadEvent(2, logger)
		So(logger.log.Len(), ShouldEqual, 0)

		cases := map[int]int{
			ev1.ownerID: 1,
			ev2.ownerID: 3,
		}

		for real, expected := range cases {
			So(real, ShouldEqual, expected)
		}
	}))
}

func TestLoadPicture(t *testing.T) {
	Convey("Loading pictures", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		pic1 := db.loadPicture(1, logger)
		pic2 := db.loadPicture(2, logger)
		So(logger.log.Len(), ShouldEqual, 0)

		cases := map[interface{}]interface{}{
			pic1.userID: 1,
			pic2.userID: 2,

			pic1.eventID: 1,
			pic2.eventID: 1,

			pic1.attachment: "test_pic.jpg",
			pic2.attachment: "guest_test_pic.jpg",
		}

		for real, expected := range cases {
			So(real, ShouldEqual, expected)
		}
	}))
}

func TestLoadPhotographerInfo(t *testing.T) {
	Convey("Loading photographerInfos", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		pi1 := db.loadPhotographerInfo(1, logger)
		pi2 := db.loadPhotographerInfo(3, logger)
		pi3 := db.loadPhotographerInfo(4, logger)
		So(logger.log.Len(), ShouldEqual, 0)

		cases := map[interface{}]interface{}{
			pi1.id: 1,
			pi2.id: 2,
			pi3.id: 3,

			pi1.picture.String: "test_watermark.jpg",
			pi2.picture.String: "extra_test_watermark.jpg",
			pi3.picture.Valid:  false,
		}

		for real, expected := range cases {
			So(real, ShouldEqual, expected)
		}
	}))
}

func TestLoadWatermark(t *testing.T) {
	Convey("Loading watermarks", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		Convey("Not found", func() {
			wm := db.loadWatermark(20, logger)
			So(wm.id, ShouldEqual, 0)
		})

		Convey("Should only return the default", func() {
			wm := db.loadWatermark(1, logger)
			So(wm.id, ShouldEqual, 1)
			So(wm.logo.String, ShouldEqual, "test_watermark.jpg")
			So(wm.alpha, ShouldEqual, 70)
			So(wm.position, ShouldEqual, "bottom,left")
		})

		Convey("Should properly parse all data", func() {
			wm := db.loadWatermark(3, logger)
			So(wm.id, ShouldEqual, 4)
			So(wm.logo.String, ShouldEqual, "test_watermark3.jpg")
			So(wm.disabled, ShouldEqual, false)
			So(wm.alpha, ShouldEqual, 20)
			So(wm.scale, ShouldEqual, 75)
			So(wm.offset, ShouldEqual, 0)
			So(wm.position, ShouldEqual, "top,right")

			wm = db.loadWatermark(4, logger)
			So(wm.id, ShouldEqual, 5)
			So(wm.logo.Valid, ShouldEqual, false)
			So(wm.disabled, ShouldEqual, true)
			So(wm.alpha, ShouldEqual, 0)
			So(wm.scale, ShouldEqual, 0)
			So(wm.offset, ShouldEqual, 0)
			So(wm.position, ShouldEqual, "")
		})
	}))
}

func TestNewNullString(t *testing.T) {
	Convey("NewNullString output", t, func() {
		cases := map[string]sql.NullString{
			"foo":     sql.NullString{String: "foo", Valid: true},
			"goobers": sql.NullString{String: "goobers", Valid: true},
			"":        sql.NullString{String: "", Valid: false},
		}

		for input, expected := range cases {
			So(newNullString(input), ShouldResemble, expected)
		}
	})
}
