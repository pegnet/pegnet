// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database_test

import (
	"testing"

	. "github.com/pegnet/pegnet/database"
)

// Helper function to check for errors, particularly where we don't ever expect them.
func Chk(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
	return
}

func TestDatabaseOpen(t *testing.T) {
	DB := Ldb{Pathname: "/tmp/testdb"}
	Chk(t, DB.Open(DB.Pathname))
}

type junk struct {
	A int
	B float64
	C string
	D []byte
}

/*  This test does not actually run under much other than by running by hand.
 *  Revisit when we actually start using the database.

func TestDatabaseValues(t *testing.T) {
	var err error

	DB := Ldb{Pathname: "/tmp/testdb"}
	err = DB.Open(DB.Pathname)
	Chk(t, err)

	stuff := junk{A: 3, B: 2.7, C: "Hello World", D: []byte{1, 2, 3, 4, 5, 6, 7}}
	js, err := json.Marshal(stuff)
	Chk(t, err)

	println("writing ", string(js))

	err = DB.Put(BUCKET_OPR, []byte{7, 8, 9}, js)
	Chk(t, err)
	newstuff := new(junk)
	data, err := DB.Get(BUCKET_OPR, []byte{7, 8, 9})
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(data, newstuff)
	Chk(t, err)

	println("got back", string(data))
	println("Delete this thing")
	err = DB.Delete(BUCKET_OPR, []byte{7, 8, 9})
	Chk(t, err)
	data, err = DB.Get(BUCKET_OPR, []byte{7, 8, 9})

	if data != nil || err == nil {
		t.Error("Should not get anything back")
	}
}
*/
