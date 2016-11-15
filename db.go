package main

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/nleof/goyesql"
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

// DB encapsulates a DB connection + queries
type DB struct {
	conn    *sql.DB
	queries goyesql.Queries
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
