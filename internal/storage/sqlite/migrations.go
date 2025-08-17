package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Migration struct {
	Version int
	SQL     string
}

var migrations = []Migration{
	{
		Version: 1,
		SQL: `
			CREATE TABLE IF NOT EXISTS tasks (
				id TEXT PRIMARY KEY,
				jira_id TEXT NOT NULL,
				title TEXT NOT NULL,
				priority TEXT NOT NULL CHECK (priority IN ('minor', 'normal', 'high', 'critical')),
				status TEXT NOT NULL CHECK (status IN ('new', 'in_progress', 'blocked', 'done', 'archived')),
				tags TEXT NOT NULL DEFAULT '[]',
				blockers TEXT NOT NULL DEFAULT '[]',
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
			CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
			CREATE INDEX IF NOT EXISTS idx_tasks_jira_id ON tasks(jira_id);
			CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
		`,
	},
	{
		Version: 2,
		SQL: `
			CREATE TABLE IF NOT EXISTS links (
				id TEXT PRIMARY KEY,
				task_id TEXT NOT NULL,
				type TEXT NOT NULL CHECK (type IN ('pull_request', 'slack_thread', 'jira_ticket', 'documentation', 'other')),
				url TEXT NOT NULL,
				title TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT '',
				metadata TEXT NOT NULL DEFAULT '{}',
				FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_links_task_id ON links(task_id);
			CREATE INDEX IF NOT EXISTS idx_links_type ON links(type);
		`,
	},
	{
		Version: 3,
		SQL: `
			CREATE TABLE IF NOT EXISTS comments (
				id TEXT PRIMARY KEY,
				task_id TEXT NOT NULL,
				content TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_comments_task_id ON comments(task_id);
			CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);
		`,
	},
}

func runMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			// Only log if it's not because transaction was already committed
			if !errors.Is(err, sql.ErrTxDone) {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}
	}()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return err
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version); err != nil {
		return err
	}

	return tx.Commit()
}
