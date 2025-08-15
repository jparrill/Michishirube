package models

import (
	"time"
)

// Comment represents a comment associated with a task
type Comment struct {
	ID        string    `json:"id" db:"id" example:"550e8400-e29b-41d4-a716-446655440002"`                  // Unique identifier
	TaskID    string    `json:"task_id" db:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`       // Associated task ID
	Content   string    `json:"content" db:"content" example:"Found the root cause in the controller"`     // Comment content
	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2024-01-15T11:00:00Z"`                // Creation timestamp
}

func (c *Comment) Validate() error {
	if c.TaskID == "" {
		return &ValidationError{Field: "task_id", Message: "task_id is required"}
	}
	if c.Content == "" {
		return &ValidationError{Field: "content", Message: "content is required"}
	}
	return nil
}