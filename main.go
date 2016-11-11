package main

import (
	"database/sql"
	"io"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")

	if len(connStr) == 0 {
		log.Fatal("No database URL found in environment")
	}

	db, err := sql.Open("postgres", connStr)
	handleErr(err)

	rows, err := db.Query("SELECT id, attachment FROM pictures WHERE id = $1", 457)
	handleErr(err)
	defer closeQuietly(rows)

	for rows.Next() {
		var id int
		var attachment string

		err = rows.Scan(&id, &attachment)
		handleErr(err)

		log.Printf("%d: %s", id, attachment)
	}
}

func closeQuietly(handle io.Closer) {
	err := handle.Close()
	if err != nil {
		log.Printf("Warning: %s", err)
	}
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}

}
