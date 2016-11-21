package main

import (
	"testing"
)

func TestConnectionAndQuery(t *testing.T) {
	_, db := dbTestSetup(t)

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

func TestLoadEvent(t *testing.T) {
	_, db := dbTestSetup(t)

	ev1 := db.loadEvent(1)
	ev2 := db.loadEvent(2)

	cases := map[int]int{
		ev1.ownerID: 1,
		ev2.ownerID: 3,
	}

	for real, expected := range cases {
		if real != expected {
			t.Errorf("Expected event OwnerID to be %v, got %v", expected, real)
		}
	}
}

func TestLoadPicture(t *testing.T) {
	_, db := dbTestSetup(t)

	pic1 := db.loadPicture(1)
	pic2 := db.loadPicture(2)

	cases := map[interface{}]interface{}{
		pic1.userID: 1,
		pic2.userID: 2,

		pic1.eventID: 1,
		pic2.eventID: 1,

		pic1.attachment: "test_pic.jpg",
		pic2.attachment: "guest_test_pic.jpg",
	}

	for real, expected := range cases {
		if real != expected {
			t.Errorf("Expected %v to equal %v", real, expected)
		}
	}
}

func TestLoadPhotographerInfo(t *testing.T) {
	_, db := dbTestSetup(t)

	pi1 := db.loadPhotographerInfo(1)
	pi2 := db.loadPhotographerInfo(3)
	pi3 := db.loadPhotographerInfo(4)

	cases := map[interface{}]interface{}{
		pi1.id: 1,
		pi2.id: 2,
		pi3.id: 3,

		pi1.picture.String: "test_watermark.jpg",
		pi2.picture.String: "extra_test_watermark.jpg",
		pi3.picture.Valid:  false,
	}

	for real, expected := range cases {
		if real != expected {
			t.Errorf("Expected %v to equal %v", real, expected)
		}
	}
}
