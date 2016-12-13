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
	"context"
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConnectionAndQuery(t *testing.T) {
	Convey("Connection and querying", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		testStruct := struct {
			id, userID, eventID int
			attachment          string
			watermarkID         sql.NullInt64
		}{}

		rows, err := db.conn.Query("select * from pictures limit 1")
		if err != nil {
			t.Errorf("Error querying database: %v", err)
			t.FailNow()
		}
		defer logger.CloseQuietly(rows)

		for rows.Next() {
			err = rows.Scan(
				&testStruct.id, &testStruct.userID, &testStruct.eventID,
				&testStruct.attachment, &testStruct.watermarkID,
			)
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

func TestLoadPictureInfo(t *testing.T) {
	Convey("LoadPictureInfo", t, withTestFixtures(func(config *Config, db *DB, logger testLogger) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "logger", logger)

		_, err := db.loadPictureInfo(ctx, 42)
		So(err, ShouldHaveSameTypeAs, noRowsErr{})

		info, err := db.loadPictureInfo(ctx, 1)
		So(err, ShouldBeNil)

		So(info.ownerID, ShouldEqual, 1)
		So(info.mark.id.Valid, ShouldBeFalse)
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
