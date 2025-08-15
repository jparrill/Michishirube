package sqlite

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestMigrationDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test_migrations_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	
	dbPath := tmpFile.Name()
	
	// Open database
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	require.NoError(t, err)
	
	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}
	
	return db, cleanup
}

func TestRunMigrations_EmptyDatabase(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Run migrations on empty database
	err := runMigrations(db)
	require.NoError(t, err)
	
	// Check that all tables exist
	tables := []string{"tasks", "links", "comments", "schema_migrations"}
	for _, table := range tables {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", table).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", table)
	}
}

func TestRunMigrations_SchemaVersioning(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Run migrations
	err := runMigrations(db)
	require.NoError(t, err)
	
	// Check that all migration versions were recorded
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	require.NoError(t, err)
	defer rows.Close()
	
	var versions []int
	for rows.Next() {
		var version int
		err := rows.Scan(&version)
		require.NoError(t, err)
		versions = append(versions, version)
	}
	
	// Should have all migration versions
	expectedVersions := []int{1, 2, 3}
	assert.Equal(t, expectedVersions, versions)
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Run migrations twice
	err := runMigrations(db)
	require.NoError(t, err)
	
	err = runMigrations(db)
	require.NoError(t, err)
	
	// Check that versions are still correct (no duplicates)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count) // Should still only have 3 versions
}

func TestRunMigrations_ForeignKeys(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Run migrations
	err := runMigrations(db)
	require.NoError(t, err)
	
	// Test foreign key constraints are working
	// Insert a task first
	taskID := "test-task-id"
	_, err = db.Exec("INSERT INTO tasks (id, jira_id, title, priority, status, tags, blockers, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", taskID, "TEST-123", "Test Task", "normal", "new", "[]", "[]")
	require.NoError(t, err)
	
	// Insert link with valid task_id - should succeed
	_, err = db.Exec("INSERT INTO links (id, task_id, type, url, title, status, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)", "link-1", taskID, "pull_request", "http://test.com", "Test Link", "open", "{}")
	require.NoError(t, err)
	
	// Try to insert link with invalid task_id - should fail due to foreign key constraint
	_, err = db.Exec("INSERT INTO links (id, task_id, type, url, title, status, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)", "link-2", "nonexistent-task", "pull_request", "http://test.com", "Test Link", "open", "{}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FOREIGN KEY constraint failed")
}

func TestRunMigrations_CascadeDelete(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Run migrations
	err := runMigrations(db)
	require.NoError(t, err)
	
	// Insert task, link, and comment
	taskID := "test-task-id"
	_, err = db.Exec("INSERT INTO tasks (id, jira_id, title, priority, status, tags, blockers, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", taskID, "TEST-123", "Test Task", "normal", "new", "[]", "[]")
	require.NoError(t, err)
	
	_, err = db.Exec("INSERT INTO links (id, task_id, type, url, title, status, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)", "link-1", taskID, "pull_request", "http://test.com", "Test Link", "open", "{}")
	require.NoError(t, err)
	
	_, err = db.Exec("INSERT INTO comments (id, task_id, content, created_at) VALUES (?, ?, ?, datetime('now'))", "comment-1", taskID, "Test Comment")
	require.NoError(t, err)
	
	// Verify they exist
	var linkCount, commentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM links WHERE task_id = ?", taskID).Scan(&linkCount)
	require.NoError(t, err)
	assert.Equal(t, 1, linkCount)
	
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE task_id = ?", taskID).Scan(&commentCount)
	require.NoError(t, err)
	assert.Equal(t, 1, commentCount)
	
	// Delete task
	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	
	// Verify links and comments were cascade deleted
	err = db.QueryRow("SELECT COUNT(*) FROM links WHERE task_id = ?", taskID).Scan(&linkCount)
	require.NoError(t, err)
	assert.Equal(t, 0, linkCount)
	
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE task_id = ?", taskID).Scan(&commentCount)
	require.NoError(t, err)
	assert.Equal(t, 0, commentCount)
}

func TestRunMigrations_Indexes(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Run migrations
	err := runMigrations(db)
	require.NoError(t, err)
	
	// Check that indexes were created
	expectedIndexes := []string{
		"idx_tasks_status",
		"idx_tasks_priority", 
		"idx_tasks_jira_id",
		"idx_tasks_created_at",
		"idx_links_task_id",
		"idx_links_type",
		"idx_comments_task_id",
		"idx_comments_created_at",
	}
	
	for _, indexName := range expectedIndexes {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='index' AND name=?)", indexName).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Index %s should exist", indexName)
	}
}

func TestGetCurrentVersion(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Create migrations table first
	err := createMigrationsTable(db)
	require.NoError(t, err)
	
	// Initially should be 0 (no migrations applied)
	version, err := getCurrentVersion(db)
	require.NoError(t, err)
	assert.Equal(t, 0, version)
	
	// Add some versions
	_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", 1)
	require.NoError(t, err)
	
	_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", 3)
	require.NoError(t, err)
	
	// Should return max version
	version, err = getCurrentVersion(db)
	require.NoError(t, err)
	assert.Equal(t, 3, version)
}

func TestApplyMigration(t *testing.T) {
	db, cleanup := setupTestMigrationDB(t)
	defer cleanup()
	
	// Create migrations table
	err := createMigrationsTable(db)
	require.NoError(t, err)
	
	// Apply a simple migration
	migration := Migration{
		Version: 1,
		SQL:     "CREATE TABLE test_table (id INTEGER PRIMARY KEY)",
	}
	
	err = applyMigration(db, migration)
	require.NoError(t, err)
	
	// Check table was created
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='test_table')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists)
	
	// Check version was recorded
	var recordedVersion int
	err = db.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", 1).Scan(&recordedVersion)
	require.NoError(t, err)
	assert.Equal(t, 1, recordedVersion)
}