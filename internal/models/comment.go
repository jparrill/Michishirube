package models

import (
	"time"
)

type Comment struct {
	ID        string    `json:"id" db:"id"`
	TaskID    string    `json:"task_id" db:"task_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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