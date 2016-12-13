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
)

const (
	maxIdleConns    = 0
	maxOpenConns    = 30
	maxConnLifetime = 10 * time.Second

	querySQL = `
SELECT pictures.user_id, pictures.attachment, events.owner_id, photographer_infos.id,
  photographer_infos.picture, watermarks.id, watermarks.disabled, watermarks.logo,
  watermarks.alpha, watermarks.scale, watermarks.offset, watermarks.position
FROM pictures
LEFT JOIN watermarks ON watermarks.id = pictures.watermark_id
LEFT JOIN photographer_infos ON photographer_infos.user_id = pictures.user_id
JOIN events ON events.id = pictures.event_id
WHERE pictures.id = $1;`
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

func newNullString(str string) sql.NullString {
	ns := sql.NullString{Valid: true, String: str}

	if len(str) == 0 {
		ns.Valid = false
	}

	return ns
}

func newNullBool(b bool) sql.NullBool {
	return sql.NullBool{Valid: true, Bool: b}
}

func newNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Valid: true, Int64: i}
}

type watermark struct {
	id       sql.NullInt64
	disabled sql.NullBool
	logo     sql.NullString
	alpha    sql.NullInt64
	scale    sql.NullInt64
	offset   sql.NullInt64
	position sql.NullString
}

func (wm *watermark) mungePosition() {
	if wm.position.Valid {
		matched := regexp.MustCompile(`[- ]`).ReplaceAllString(wm.position.String, "")
		matched = regexp.MustCompile(`\n`).ReplaceAllString(strings.TrimSpace(matched), ",")
		wm.position.String = matched
	}
}

// PictureInfo is the result of the picture query
type pictureInfo struct {
	userID             int
	attachment         string
	ownerID            int
	photographerInfoID sql.NullInt64
	oldMark            sql.NullString
	mark               watermark
}

// DB encapsulates a DB connection + queries
type DB struct {
	conn *sql.DB
}

// NewDB connects to the database and loads the queries from YeSQL.
func NewDB(c *Config) (*DB, error) {
	conn, err := sql.Open("postgres", c.DatabaseURL)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(maxConnLifetime)

	db := DB{
		conn: conn,
	}

	return &db, nil
}

func (db *DB) loadPictureInfo(ctx context.Context, id int) (pictureInfo, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger := ctxTimeout.Value("logger").(ILogger)

	outChan := make(chan pictureInfo)
	errChan := make(chan error)

	go func() {
		info := pictureInfo{}

		err := db.conn.QueryRow(querySQL, id).Scan(
			&info.userID, &info.attachment, &info.ownerID, &info.photographerInfoID,
			&info.oldMark, &info.mark.id, &info.mark.disabled, &info.mark.logo,
			&info.mark.alpha, &info.mark.scale, &info.mark.offset, &info.mark.position,
		)

		switch {
		case err == sql.ErrNoRows:
			err = newNoRowsError("No picture found with id %d", id)
			fallthrough
		case err != nil:
			errChan <- err
		default:
			info.mark.mungePosition()
			logger.Debug("Picture Info for %d: %v", id, info)
			outChan <- info
		}
	}()

	select {
	case info := <-outChan:
		return info, nil
	case err := <-errChan:
		return pictureInfo{}, err
	case <-ctxTimeout.Done():
		return pictureInfo{}, fmt.Errorf("context timeout: %v", ctxTimeout.Err())
	}
}
