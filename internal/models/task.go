package models

import (
	"time"
)

type Priority string

const (
	Minor    Priority = "minor"
	Normal   Priority = "normal"
	High     Priority = "high"
	Critical Priority = "critical"
)

const DefaultPriority = Normal

type Status string

const (
	New        Status = "new"
	InProgress Status = "in_progress"
	Blocked    Status = "blocked"
	Done       Status = "done"
	Archived   Status = "archived"
)

const (
	DefaultStatus  = New
	DefaultNoJira = "NO-JIRA"
)

type Task struct {
	ID        string    `json:"id" db:"id"`
	JiraID    string    `json:"jira_id" db:"jira_id"`
	Title     string    `json:"title" db:"title"`
	Priority  Priority  `json:"priority" db:"priority"`
	Status    Status    `json:"status" db:"status"`
	Tags      []string  `json:"tags" db:"tags"`
	Blockers  []string  `json:"blockers" db:"blockers"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (p Priority) IsValid() bool {
	switch p {
	case Minor, Normal, High, Critical:
		return true
	}
	return false
}

func (s Status) IsValid() bool {
	switch s {
	case New, InProgress, Blocked, Done, Archived:
		return true
	}
	return false
}

func (t *Task) Validate() error {
	if t.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	
	// Set defaults if empty
	if t.JiraID == "" {
		t.JiraID = DefaultNoJira
	}
	if t.Priority == "" {
		t.Priority = DefaultPriority
	}
	if t.Status == "" {
		t.Status = DefaultStatus
	}
	
	// Validate after setting defaults
	if !t.Priority.IsValid() {
		return &ValidationError{Field: "priority", Message: "invalid priority"}
	}
	if !t.Status.IsValid() {
		return &ValidationError{Field: "status", Message: "invalid status"}
	}
	
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}