package models

// API Request/Response DTOs for Swagger documentation

// TaskListResponse represents the response for listing tasks
type TaskListResponse struct {
	Tasks  []*Task `json:"tasks"`  // List of tasks
	Total  int     `json:"total" example:"42"`   // Total number of tasks
	Limit  int     `json:"limit" example:"20"`   // Request limit
	Offset int     `json:"offset" example:"0"`   // Request offset
}

// TaskWithDetails represents a task with all related data
type TaskWithDetails struct {
	*Task
	Links    []*Link    `json:"links"`    // Associated links
	Comments []*Comment `json:"comments"` // Associated comments
}

// CreateTaskRequest represents request to create a new task
type CreateTaskRequest struct {
	JiraID   string   `json:"jira_id" example:"OCPBUGS-5678"`                         // Jira ticket ID
	Title    string   `json:"title" example:"Implement new feature"`                  // Task title
	Priority Priority `json:"priority" example:"high"`                                // Task priority
	Tags     []string `json:"tags"`      // Task tags
	Blockers []string `json:"blockers"` // Blocking issues
}

// UpdateTaskRequest represents request to update a task
type UpdateTaskRequest struct {
	JiraID   string   `json:"jira_id" example:"OCPBUGS-5678"`                         // Jira ticket ID
	Title    string   `json:"title" example:"Updated task title"`                     // Task title
	Priority Priority `json:"priority" example:"high"`                                // Task priority
	Status   Status   `json:"status" example:"in_progress"`                           // Task status
	Tags     []string `json:"tags"`      // Task tags
	Blockers []string `json:"blockers"` // Blocking issues
}

// PatchTaskRequest represents request to partially update a task
type PatchTaskRequest struct {
	Status   *Status   `json:"status,omitempty" example:"in_progress"`                 // Task status
	Priority *Priority `json:"priority,omitempty" example:"high"`                      // Task priority
	Title    *string   `json:"title,omitempty" example:"Updated title"`                // Task title
	Tags     []string  `json:"tags,omitempty"`     // Task tags
	Blockers []string  `json:"blockers,omitempty"` // Blocking issues
}

// CreateLinkRequest represents request to create a new link
type CreateLinkRequest struct {
	TaskID   string   `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`             // Associated task ID
	Type     LinkType `json:"type" example:"pull_request"`                                         // Link type
	URL      string   `json:"url" example:"https://github.com/org/repo/pull/456"`                 // Link URL
	Title    string   `json:"title,omitempty" example:"Fix memory leak"`                          // Display title
	Status   string   `json:"status,omitempty" example:"merged"`                                  // Link status
	Metadata string   `json:"metadata,omitempty"`  // Additional metadata
}

// UpdateLinkRequest represents request to update a link
type UpdateLinkRequest struct {
	Type     LinkType `json:"type" example:"pull_request"`                                         // Link type
	URL      string   `json:"url" example:"https://github.com/org/repo/pull/456"`                 // Link URL
	Title    string   `json:"title,omitempty" example:"Updated link title"`                       // Display title
	Status   string   `json:"status,omitempty" example:"merged"`                                  // Link status
	Metadata string   `json:"metadata,omitempty"`  // Additional metadata
}

// CreateCommentRequest represents request to create a new comment
type CreateCommentRequest struct {
	TaskID  string `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`               // Associated task ID
	Content string `json:"content" example:"Found the root cause in the controller"`             // Comment content
}

// CreateCommentResponse represents response when creating a comment
type CreateCommentResponse struct {
	ID      string `json:"id" example:"550e8400-e29b-41d4-a716-446655440002"`                    // Comment ID
	Message string `json:"message" example:"Comment created successfully"`                       // Success message
}

// DeleteCommentResponse represents response when deleting a comment
type DeleteCommentResponse struct {
	Message string `json:"message" example:"Comment deleted successfully"`                       // Success message
}

// ReportResponse represents the status report response
type ReportResponse struct {
	WorkingOn []*TaskWithDetails `json:"working_on"` // Tasks in progress or completed
	NextUp    []*TaskWithDetails `json:"next_up"`    // Tasks to work on next
	Blockers  []*TaskWithDetails `json:"blockers"`   // Blocked tasks
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Task not found"`           // Error message
	Code  string `json:"code,omitempty" example:"TASK_NOT_FOUND"`  // Error code
}