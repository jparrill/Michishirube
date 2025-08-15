//go:build purego

package sqlite

import (
	"database/sql"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// openDB opens a SQLite database using the pure Go driver
func openDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open("sqlite", dataSourceName)
}