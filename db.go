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
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/nleof/goyesql"
)

type noRowsErr struct {
	message string
}

func (n noRowsErr) Error() string {
	return n.message
}

func newNoRowsError(text string, a ...interface{}) noRowsErr {
	n := noRowsErr{}
	n.message = fmt.Sprintf(text, a)

	return n
}

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
	alpha    sql.NullInt64
	scale    sql.NullInt64
	offset   sql.NullInt64
	position sql.NullString
}

// DB encapsulates a DB connection + queries
type DB struct {
	conn    *sql.DB
	queries goyesql.Queries
}

func newNullString(str string) sql.NullString {
	ns := sql.NullString{Valid: true, String: str}

	if len(str) == 0 {
		ns.Valid = false
	}

	return ns
}

func newNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Valid: true, Int64: i}
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

func (db *DB) loadPicture(ctx context.Context, id int) (picture, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger := ctxTimeout.Value("logger").(ILogger)

	outChan := make(chan picture)
	errChan := make(chan error)

	go func() {
		pic := picture{}
		err := db.conn.QueryRow(db.queries["picture_by_id"], id).Scan(
			&pic.userID, &pic.eventID, &pic.attachment)

		switch {
		case err == sql.ErrNoRows:
			errChan <- newNoRowsError("No picture found with id %d", id)
		case err != nil:
			errChan <- err
		default:
			logger.Debug("Picture for id %d: %v", id, pic)
			outChan <- pic
		}
	}()

	select {
	case pic := <-outChan:
		return pic, nil
	case err := <-errChan:
		return picture{}, err
	case <-ctxTimeout.Done():
		return picture{}, fmt.Errorf("context timeout: %v", ctxTimeout.Err())
	}
}

func (db *DB) loadEvent(ctx context.Context, id int) (event, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger := timeoutCtx.Value("logger").(ILogger)

	outChan := make(chan event)
	errChan := make(chan error)

	go func() {
		ev := event{}

		err := db.conn.QueryRow(db.queries["event_by_id"], id).Scan(&ev.ownerID)
		switch {
		case err == sql.ErrNoRows:
			errChan <- newNoRowsError("No event found with id %d", id)
		case err != nil:
			errChan <- err
		default:
			logger.Debug("Event for id %d: %v", id, ev)
			outChan <- ev
		}
	}()

	select {
	case ev := <-outChan:
		return ev, nil
	case err := <-errChan:
		return event{}, err
	case <-timeoutCtx.Done():
		return event{}, fmt.Errorf("context timeout: %v", timeoutCtx.Err())
	}
}

func (db *DB) loadPhotographerInfo(ctx context.Context, userID int) (photographerInfo, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger := timeoutCtx.Value("logger").(ILogger)

	outChan := make(chan photographerInfo)
	errChan := make(chan error)

	go func() {
		pi := photographerInfo{}

		err := db.conn.QueryRow(db.queries["photographer_info_by_user_id"], userID).Scan(
			&pi.id, &pi.picture)
		switch {
		case err == sql.ErrNoRows:
			errChan <- newNoRowsError("No photographer info found for user id %d", userID)
		case err != nil:
			errChan <- err
		default:
			logger.Debug("PhotographerInfo for user ID %d: %v", userID, pi)
			outChan <- pi
		}
	}()

	select {
	case pi := <-outChan:
		return pi, nil
	case err := <-errChan:
		return photographerInfo{}, err
	case <-timeoutCtx.Done():
		return photographerInfo{}, fmt.Errorf("context timeout: %v", timeoutCtx.Err())
	}
}

func (db *DB) loadWatermark(ctx context.Context, photographerInfoID int) (watermark, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger := timeoutCtx.Value("logger").(ILogger)

	outChan := make(chan watermark)
	errChan := make(chan error)

	go func() {
		wm := watermark{}

		err := db.conn.QueryRow(db.queries["watermark_by_photographer_info_id"], photographerInfoID).Scan(
			&wm.id, &wm.logo, &wm.disabled, &wm.alpha, &wm.scale, &wm.offset, &wm.position)

		switch {
		case err == sql.ErrNoRows:
			errChan <- newNoRowsError("No watermark found for photographerInfo %d", photographerInfoID)
		case err != nil:
			errChan <- err
		default:
			if wm.position.Valid {
				matched := regexp.MustCompile(`[- ]`).ReplaceAllString(wm.position.String, "")
				matched = regexp.MustCompile(`\n`).ReplaceAllString(strings.TrimSpace(matched), ",")
				wm.position.String = matched
			}

			logger.Debug("Watermark for photographer ID %d: %v", photographerInfoID, wm)
			outChan <- wm
		}
	}()

	select {
	case wm := <-outChan:
		return wm, nil
	case err := <-errChan:
		return watermark{}, err
	case <-timeoutCtx.Done():
		return watermark{}, fmt.Errorf("context timeout: %v", timeoutCtx.Err())
	}
}
