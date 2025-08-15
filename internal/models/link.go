package models

type LinkType string

const (
	PullRequest   LinkType = "pull_request"
	SlackThread   LinkType = "slack_thread"
	JiraTicket    LinkType = "jira_ticket"
	Documentation LinkType = "documentation"
	Other         LinkType = "other"
)

// Link represents an external link associated with a task
type Link struct {
	ID       string   `json:"id" db:"id" example:"550e8400-e29b-41d4-a716-446655440001"`                     // Unique identifier
	TaskID   string   `json:"task_id" db:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`          // Associated task ID
	Type     LinkType `json:"type" db:"type" example:"pull_request"`                                         // Type of link
	URL      string   `json:"url" db:"url" example:"https://github.com/org/repo/pull/456"`                  // Link URL
	Title    string   `json:"title" db:"title" example:"Fix memory leak"`                                    // Display title
	Status   string   `json:"status" db:"status" example:"merged"`                                           // Link status
	Metadata string   `json:"metadata" db:"metadata" example:"{\"pr_number\": 456, \"author\": \"user\"}"`  // Additional metadata
}

func (lt LinkType) IsValid() bool {
	switch lt {
	case PullRequest, SlackThread, JiraTicket, Documentation, Other:
		return true
	}
	return false
}

func (l *Link) Validate() error {
	if l.TaskID == "" {
		return &ValidationError{Field: "task_id", Message: "task_id is required"}
	}
	if l.URL == "" {
		return &ValidationError{Field: "url", Message: "url is required"}
	}
	if !l.Type.IsValid() {
		return &ValidationError{Field: "type", Message: "invalid link type"}
	}
	if l.Title == "" {
		l.Title = l.URL
	}
	return nil
}