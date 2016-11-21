package main

import (
	"testing"
)

func connect(c *Config, t *testing.T) *DB {
	conn, err := NewDB(c)

	if err != nil {
		t.Errorf("Error when connecting to DB: %v", err)
	}

	return conn
}

func TestConnectionAndQuery(t *testing.T) {
	config := load(t)
	db := connect(config, t)

	testStruct := struct {
		id, userID, eventID int
		attachment          string
	}{}

	rows, err := db.conn.Query("select * from pictures limit 1")
	if err != nil {
		t.Errorf("Error querying database: %v", err)
		t.FailNow()
	}
	defer closeQuietly(rows)

	for rows.Next() {
		err = rows.Scan(&testStruct.id, &testStruct.userID, &testStruct.eventID, &testStruct.attachment)
		if err != nil {
			t.Errorf("Error scanning database row: %v", err)
			t.FailNow()
		}
	}

	cases := map[interface{}]interface{}{
		testStruct.id:         1,
		testStruct.userID:     1,
		testStruct.eventID:    1,
		testStruct.attachment: "test_pic.jpg",
	}

	for got, expected := range cases {
		if got != expected {
			t.Errorf("Expected %v to == %v", got, expected)
		}
	}
}
