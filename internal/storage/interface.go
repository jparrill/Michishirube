//go:generate mockgen -source=interface.go -destination=../handlers/mocks/storage_mock.go -package=mocks

package storage

import (
	"michishirube/internal/models"
)

type Storage interface {
	// Tasks
	CreateTask(task *models.Task) error
	GetTask(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ListTasks(filters TaskFilters) ([]*models.Task, error)
	SearchTasks(query string, includeArchived bool, limit int) ([]*models.Task, error)

	// Links
	CreateLink(link *models.Link) error
	GetLink(id string) (*models.Link, error)
	UpdateLink(link *models.Link) error
	DeleteLink(id string) error
	GetTaskLinks(taskID string) ([]*models.Link, error)

	// Comments
	CreateComment(comment *models.Comment) error
	GetComment(id string) (*models.Comment, error)
	DeleteComment(id string) error
	GetTaskComments(taskID string) ([]*models.Comment, error)

	// Migrations
	RunMigrations() error
	Close() error
}

type TaskFilters struct {
	Status           []models.Status
	Priority         []models.Priority
	Tags             []string
	IncludeArchived  bool
	Limit            int
	Offset           int
}