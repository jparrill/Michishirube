//go:generate mockgen -source=interface.go -destination=../handlers/mocks/storage_mock.go -package=mocks

package storage

import (
	"michishirube/internal/models"
)

type Storage interface {
	// Tasks
	// CreateTask creates a new task
	CreateTask(task *models.Task) error
	// GetTask retrieves a task by its ID
	GetTask(id string) (*models.Task, error)
	// UpdateTask updates an existing task
	UpdateTask(task *models.Task) error
	// DeleteTask deletes a task by its ID
	DeleteTask(id string) error
	// ListTasks retrieves a list of tasks based on the provided filters
	ListTasks(filters TaskFilters) ([]*models.Task, error)
	SearchTasks(query string, includeArchived bool, limit int) ([]*models.Task, error)

	// Links
	// CreateLink creates a new link
	CreateLink(link *models.Link) error
	// GetLink retrieves a link by its ID
	GetLink(id string) (*models.Link, error)
	// UpdateLink updates an existing link
	UpdateLink(link *models.Link) error
	// DeleteLink deletes a link by its ID
	DeleteLink(id string) error
	// GetTaskLinks retrieves all links for a specific task
	GetTaskLinks(taskID string) ([]*models.Link, error)

	// Comments
	// CreateComment creates a new comment
	CreateComment(comment *models.Comment) error
	// GetComment retrieves a comment by its ID
	GetComment(id string) (*models.Comment, error)
	// DeleteComment deletes a comment by its ID
	DeleteComment(id string) error
	GetTaskComments(taskID string) ([]*models.Comment, error)

	// Migrations
	// RunMigrations runs the database migrations
	RunMigrations() error
	// Close closes the database connection
	Close() error
}

// TaskFilters is a struct that contains the filters for the tasks
type TaskFilters struct {
	Status          []models.Status
	Priority        []models.Priority
	Tags            []string
	IncludeArchived bool
	Limit           int
	Offset          int
}
