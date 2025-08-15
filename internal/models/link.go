package models

type LinkType string

const (
	PullRequest   LinkType = "pull_request"
	SlackThread   LinkType = "slack_thread"
	JiraTicket    LinkType = "jira_ticket"
	Documentation LinkType = "documentation"
	Other         LinkType = "other"
)

type Link struct {
	ID       string   `json:"id" db:"id"`
	TaskID   string   `json:"task_id" db:"task_id"`
	Type     LinkType `json:"type" db:"type"`
	URL      string   `json:"url" db:"url"`
	Title    string   `json:"title" db:"title"`
	Status   string   `json:"status" db:"status"`
	Metadata string   `json:"metadata" db:"metadata"`
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