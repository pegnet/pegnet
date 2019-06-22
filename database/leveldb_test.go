// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package database_test

import (
	. "github.com/pegnet/pegnet/database"
	"testing"
)

func TestDatabaseOpen(t *testing.T) {
	DB := Ldb{Pathname: "/tmp/testdb"}
	DB.Open()
}
