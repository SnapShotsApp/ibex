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

	_ "github.com/lib/pq"
	"github.com/nleof/goyesql"
	"regexp"
	"strings"
)

type picture struct {
	userID     int
	eventID    int
	attachment string
}

type event struct {
	ownerID int
}

type photographerInfo struct {
	id      int
	picture sql.NullString
}

type watermark struct {
	id       int
	logo     sql.NullString
	disabled bool
	alpha    int
	scale    int
	offset   int
	position string
}

// DB encapsulates a DB connection + queries
type DB struct {
	conn    *sql.DB
	queries goyesql.Queries
}

func newNullString(str string) sql.NullString {
	return sql.NullString{Valid: true, String: str}
}

// NewDB connects to the database and loads the queries from YeSQL.
func NewDB(c *Config) (*DB, error) {
	conn, err := dbConnect(c)
	if err != nil {
		return nil, err
	}

	db := DB{
		conn:    conn,
		queries: loadYesql(),
	}

	return &db, nil
}

func loadYesql() goyesql.Queries {
	data := MustAsset("resources/pictures.sql")
	return goyesql.MustParseBytes(data)
}

func dbConnect(c *Config) (*sql.DB, error) {
	return sql.Open("postgres", c.DatabaseURL)
}

func (db *DB) loadPicture(id int) picture {
	var pic = picture{}

	rows, err := db.conn.Query(db.queries["picture_by_id"], id)
	handleErr(err)
	defer closeQuietly(rows)

	for rows.Next() {
		err = rows.Scan(&pic.userID, &pic.eventID, &pic.attachment)
		handleErr(err)
	}

	return pic
}

func (db *DB) loadEvent(id int) event {
	var ev = event{}

	rows, err := db.conn.Query(db.queries["event_by_id"], id)
	handleErr(err)
	defer closeQuietly(rows)

	for rows.Next() {
		err = rows.Scan(&ev.ownerID)
		handleErr(err)
	}

	return ev
}

func (db *DB) loadPhotographerInfo(userID int) photographerInfo {
	var pi = photographerInfo{}

	rows, err := db.conn.Query(db.queries["photographer_info_by_user_id"], userID)
	handleErr(err)
	defer closeQuietly(rows)

	for rows.Next() {
		err = rows.Scan(&pi.id, &pi.picture)
		handleErr(err)
	}

	return pi
}

func (db *DB) loadWatermark(photographerInfoID int) watermark {
	var wm = watermark{}

	err := db.conn.QueryRow(db.queries["watermark_by_photographer_info_id"], photographerInfoID).Scan(
		&wm.id, &wm.logo, &wm.disabled, &wm.alpha, &wm.scale, &wm.offset, &wm.position)

	switch {
	case err == sql.ErrNoRows:
		Debug("No watermark found for photographerInfo %d", photographerInfoID)
		break
	case err != nil:
		handleErr(err)
		break
	default:
		matched := regexp.MustCompile(`[- ]`).ReplaceAllString(wm.position, "")
		matched = regexp.MustCompile(`\n`).ReplaceAllString(strings.TrimSpace(matched), ",")
		wm.position = matched
	}

	return wm
}
