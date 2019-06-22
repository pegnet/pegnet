package database_test

import (
	. "github.com/pegnet/pegnet/database"
	"testing"
)

func TestDatabaseOpen(t *testing.T) {
	DB := Ldb{Pathname: "/tmp/testdb"}
	DB.Open()
}