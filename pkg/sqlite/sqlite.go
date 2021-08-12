package sqlite

import (
	// Modules
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	sqLiteDriver        = &driver.SQLiteDriver{}
	sqLiteVersion, _, _ = driver.Version()
	sqLiteMemory        = ":memory:"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

func Version() string {
	return sqLiteVersion
}
