package sqlite

import (
	"errors"
	"os"
	"testing"
	"time"

	"michishirube/internal/models"
	"michishirube/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*SQLiteStorage, func()) {
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test_michishirube_*.db")
	require.NoError(t, err)
	if err := tmpFile.Close(); err != nil {
		t.Logf("failed to close temp file: %v", err)
	}

	dbPath := tmpFile.Name()

	// Initialize storage
	store, err := New(dbPath)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		if err := store.Close(); err != nil {
			t.Logf("failed to close store: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			t.Logf("failed to remove temp DB file: %v", err)
		}
	}

	return store, cleanup
}

func createTestTask(_ *testing.T) *models.Task {
	return &models.Task{
		Title:    "Test Task",
		JiraID:   "TEST-123",
		Priority: models.High,
		Status:   models.InProgress,
		Tags:     []string{"test", "unit"},
		Blockers: []string{"waiting for review"},
	}
}

func TestSQLiteStorage_CreateTask(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task := createTestTask(t)

	err := store.CreateTask(task)
	require.NoError(t, err)

	// Verify task was created with ID and timestamps
	assert.NotEmpty(t, task.ID)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())
	assert.Equal(t, task.CreatedAt, task.UpdatedAt)
}

func TestSQLiteStorage_CreateTask_WithDefaults(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task := &models.Task{
		Title: "Minimal Task",
	}

	err := store.CreateTask(task)
	require.NoError(t, err)

	// Verify defaults were applied
	assert.Equal(t, models.DefaultNoJira, task.JiraID)
	assert.Equal(t, models.DefaultPriority, task.Priority)
	assert.Equal(t, models.DefaultStatus, task.Status)
}

func TestSQLiteStorage_CreateTask_ValidationError(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task := &models.Task{
		// Missing title - should cause validation error
		Priority: models.High,
	}

	err := store.CreateTask(task)
	require.Error(t, err)

	// Should be validation error
	var validationErr *models.ValidationError
	ok := errors.As(err, &validationErr)
	assert.True(t, ok)
}

func TestSQLiteStorage_GetTask(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create task
	original := createTestTask(t)
	err := store.CreateTask(original)
	require.NoError(t, err)

	// Get task
	retrieved, err := store.GetTask(original.ID)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Title, retrieved.Title)
	assert.Equal(t, original.JiraID, retrieved.JiraID)
	assert.Equal(t, original.Priority, retrieved.Priority)
	assert.Equal(t, original.Status, retrieved.Status)
	assert.Equal(t, original.Tags, retrieved.Tags)
	assert.Equal(t, original.Blockers, retrieved.Blockers)
	assert.True(t, original.CreatedAt.Equal(retrieved.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(retrieved.UpdatedAt))
}

func TestSQLiteStorage_GetTask_NotFound(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task, err := store.GetTask("nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "not found")
}

func TestSQLiteStorage_UpdateTask(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create task
	task := createTestTask(t)
	err := store.CreateTask(task)
	require.NoError(t, err)

	originalUpdatedAt := task.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure timestamp difference

	// Update task
	task.Title = "Updated Title"
	task.Status = models.Done
	task.Tags = []string{"updated", "test"}

	err = store.UpdateTask(task)
	require.NoError(t, err)

	// Verify updated_at changed
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))

	// Verify changes were persisted
	retrieved, err := store.GetTask(task.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, models.Done, retrieved.Status)
	assert.Equal(t, []string{"updated", "test"}, retrieved.Tags)
}

func TestSQLiteStorage_DeleteTask(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create task
	task := createTestTask(t)
	err := store.CreateTask(task)
	require.NoError(t, err)

	// Delete task
	err = store.DeleteTask(task.ID)
	require.NoError(t, err)

	// Verify task is gone
	_, err = store.GetTask(task.ID)
	assert.Error(t, err)
}

func TestSQLiteStorage_ListTasks(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple tasks
	tasks := []*models.Task{
		{Title: "Task 1", Priority: models.High, Status: models.New, Tags: []string{"tag1"}},
		{Title: "Task 2", Priority: models.Normal, Status: models.InProgress, Tags: []string{"tag2"}},
		{Title: "Task 3", Priority: models.High, Status: models.Done, Tags: []string{"tag1", "tag3"}},
		{Title: "Task 4", Priority: models.Minor, Status: models.Archived, Tags: []string{"tag4"}},
	}

	for _, task := range tasks {
		err := store.CreateTask(task)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		filters  storage.TaskFilters
		expected int
	}{
		{
			name:     "all tasks",
			filters:  storage.TaskFilters{},
			expected: 3, // archived excluded by default
		},
		{
			name:     "include archived",
			filters:  storage.TaskFilters{IncludeArchived: true},
			expected: 4,
		},
		{
			name:     "filter by status",
			filters:  storage.TaskFilters{Status: []models.Status{models.New, models.InProgress}},
			expected: 2,
		},
		{
			name:     "filter by priority",
			filters:  storage.TaskFilters{Priority: []models.Priority{models.High}},
			expected: 2,
		},
		{
			name:     "with limit",
			filters:  storage.TaskFilters{Limit: 2},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := store.ListTasks(tt.filters)
			require.NoError(t, err)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestSQLiteStorage_SearchTasks(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create tasks with searchable content
	tasks := []*models.Task{
		{Title: "Memory leak investigation", JiraID: "BUG-123", Tags: []string{"memory", "performance"}},
		{Title: "Add new feature", JiraID: "FEAT-456", Tags: []string{"feature", "api"}},
		{Title: "Fix memory allocation", JiraID: "BUG-789", Tags: []string{"memory", "bug"}},
	}

	for _, task := range tasks {
		err := store.CreateTask(task)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "search by title",
			query:    "memory",
			expected: 2,
		},
		{
			name:     "search by jira id",
			query:    "BUG-123",
			expected: 1,
		},
		{
			name:     "search by partial match",
			query:    "feat",
			expected: 1,
		},
		{
			name:     "no matches",
			query:    "nonexistent",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := store.SearchTasks(tt.query, false, 10)
			require.NoError(t, err)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestSQLiteStorage_LinkOperations(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a task first
	task := createTestTask(t)
	err := store.CreateTask(task)
	require.NoError(t, err)

	// Create link
	link := &models.Link{
		TaskID: task.ID,
		Type:   models.PullRequest,
		URL:    "https://github.com/org/repo/pull/123",
		Title:  "Fix memory leak",
		Status: "open",
	}

	// Test CreateLink
	err = store.CreateLink(link)
	require.NoError(t, err)
	assert.NotEmpty(t, link.ID)

	// Test GetLink
	retrieved, err := store.GetLink(link.ID)
	require.NoError(t, err)
	assert.Equal(t, link.TaskID, retrieved.TaskID)
	assert.Equal(t, link.Type, retrieved.Type)
	assert.Equal(t, link.URL, retrieved.URL)
	assert.Equal(t, link.Title, retrieved.Title)
	assert.Equal(t, link.Status, retrieved.Status)

	// Test UpdateLink
	link.Status = "merged"
	link.Title = "Fix memory leak - merged"
	err = store.UpdateLink(link)
	require.NoError(t, err)

	updated, err := store.GetLink(link.ID)
	require.NoError(t, err)
	assert.Equal(t, "merged", updated.Status)
	assert.Equal(t, "Fix memory leak - merged", updated.Title)

	// Test GetTaskLinks
	links, err := store.GetTaskLinks(task.ID)
	require.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, link.ID, links[0].ID)

	// Test DeleteLink
	err = store.DeleteLink(link.ID)
	require.NoError(t, err)

	_, err = store.GetLink(link.ID)
	assert.Error(t, err)
}

func TestSQLiteStorage_LinkValidation(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create task first
	task := createTestTask(t)
	err := store.CreateTask(task)
	require.NoError(t, err)

	// Test link with missing URL
	link := &models.Link{
		TaskID: task.ID,
		Type:   models.PullRequest,
		Title:  "Test Link",
	}

	err = store.CreateLink(link)
	require.Error(t, err)

	var validationErr *models.ValidationError
	ok := errors.As(err, &validationErr)
	assert.True(t, ok)
}

func TestSQLiteStorage_CommentOperations(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a task first
	task := createTestTask(t)
	err := store.CreateTask(task)
	require.NoError(t, err)

	// Create comment
	comment := &models.Comment{
		TaskID:  task.ID,
		Content: "This is a test comment",
	}

	// Test CreateComment
	err = store.CreateComment(comment)
	require.NoError(t, err)
	assert.NotEmpty(t, comment.ID)
	assert.False(t, comment.CreatedAt.IsZero())

	// Test GetComment
	retrieved, err := store.GetComment(comment.ID)
	require.NoError(t, err)
	assert.Equal(t, comment.TaskID, retrieved.TaskID)
	assert.Equal(t, comment.Content, retrieved.Content)
	assert.True(t, comment.CreatedAt.Equal(retrieved.CreatedAt))

	// Create another comment for ordering test
	comment2 := &models.Comment{
		TaskID:  task.ID,
		Content: "Second comment",
	}
	err = store.CreateComment(comment2)
	require.NoError(t, err)

	// Test GetTaskComments (should be ordered by created_at)
	comments, err := store.GetTaskComments(task.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, comment.ID, comments[0].ID) // First comment should be first
	assert.Equal(t, comment2.ID, comments[1].ID)

	// Test DeleteComment
	err = store.DeleteComment(comment.ID)
	require.NoError(t, err)

	_, err = store.GetComment(comment.ID)
	assert.Error(t, err)

	// Verify only one comment remains
	comments, err = store.GetTaskComments(task.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 1)
}

func TestSQLiteStorage_CommentValidation(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Test comment with missing content
	comment := &models.Comment{
		TaskID: "some-task-id",
	}

	err := store.CreateComment(comment)
	require.Error(t, err)

	var validationErr *models.ValidationError
	ok := errors.As(err, &validationErr)
	assert.True(t, ok)
}

func TestSQLiteStorage_CascadeDelete(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create task with links and comments
	task := createTestTask(t)
	err := store.CreateTask(task)
	require.NoError(t, err)

	// Create link
	link := &models.Link{
		TaskID: task.ID,
		Type:   models.PullRequest,
		URL:    "https://github.com/test/repo/pull/1",
		Title:  "Test PR",
	}
	err = store.CreateLink(link)
	require.NoError(t, err)

	// Create comment
	comment := &models.Comment{
		TaskID:  task.ID,
		Content: "Test comment",
	}
	err = store.CreateComment(comment)
	require.NoError(t, err)

	// Delete task
	err = store.DeleteTask(task.ID)
	require.NoError(t, err)

	// Verify links and comments are also deleted (CASCADE)
	_, err = store.GetLink(link.ID)
	assert.Error(t, err)

	_, err = store.GetComment(comment.ID)
	assert.Error(t, err)

	links, err := store.GetTaskLinks(task.ID)
	require.NoError(t, err)
	assert.Len(t, links, 0)

	comments, err := store.GetTaskComments(task.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 0)
}
