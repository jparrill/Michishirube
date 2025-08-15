//go:build !purego

package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // CGO SQLite driver
)

// openDB opens a SQLite database using the CGO driver
func openDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open("sqlite3", dataSourceName)
}