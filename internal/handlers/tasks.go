package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"michishirube/internal/logger"
	"michishirube/internal/models"
	"michishirube/internal/storage"
)

type TaskHandler struct {
	storage storage.Storage
}

func NewTaskHandler(storage storage.Storage) *TaskHandler {
	return &TaskHandler{storage: storage}
}

func (h *TaskHandler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TaskHandler) HandleTask(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if path == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	taskID := strings.Split(path, "/")[0]

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, taskID)
	case http.MethodPut:
		h.updateTask(w, r, taskID)
	case http.MethodPatch:
		h.patchTask(w, r, taskID)
	case http.MethodDelete:
		h.deleteTask(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listTasks retrieves a list of tasks with optional filtering
// @Summary List tasks
// @Description Get all tasks with optional filtering by status, priority, tags, etc.
// @Tags tasks
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (comma-separated)" example("new,in_progress")
// @Param priority query string false "Filter by priority (comma-separated)" example("high,critical")
// @Param tags query string false "Filter by tags (comma-separated)" example("k8s,memory")
// @Param include_archived query boolean false "Include archived tasks" default(false)
// @Param limit query int false "Maximum number of results" default(50) minimum(1) maximum(200)
// @Param offset query int false "Number of results to skip" default(0) minimum(0)
// @Success 200 {object} models.TaskListResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /tasks [get]
func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	filters := storage.TaskFilters{}
	query := r.URL.Query()

	// Parse query parameters using switch for cleaner logic
	for param, values := range query {
		if len(values) == 0 {
			continue
		}
		value := values[0]

		switch param {
		case "status":
			statusStrings := strings.Split(value, ",")
			for _, s := range statusStrings {
				filters.Status = append(filters.Status, models.Status(strings.TrimSpace(s)))
			}
		case "priority":
			priorityStrings := strings.Split(value, ",")
			for _, p := range priorityStrings {
				filters.Priority = append(filters.Priority, models.Priority(strings.TrimSpace(p)))
			}
		case "tags":
			filters.Tags = strings.Split(value, ",")
		case "include_archived":
			switch value {
			case "true", "1":
				filters.IncludeArchived = true
			}
		case "limit":
			if limit, err := strconv.Atoi(value); err == nil && limit > 0 {
				filters.Limit = limit
			}
		case "offset":
			if offset, err := strconv.Atoi(value); err == nil && offset >= 0 {
				filters.Offset = offset
			}
		}
	}

	tasks, err := h.storage.ListTasks(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tasks":  tasks,
		"total":  len(tasks),
		"limit":  filters.Limit,
		"offset": filters.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// createTask creates a new task
// @Summary Create a new task
// @Description Create a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.CreateTaskRequest true "Task to create"
// @Success 201 {object} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Router /tasks [post]
func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.storage.CreateTask(&task)
	switch {
	case err == nil:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	case isValidationError(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getTask retrieves a specific task by ID
// @Summary Get task by ID
// @Description Retrieve a specific task with its links and comments
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID" format(uuid)
// @Success 200 {object} models.TaskWithDetails
// @Failure 404 {object} models.ErrorResponse
// @Router /tasks/{id} [get]
func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request, taskID string) {
	task, err := h.storage.GetTask(taskID)
	switch {
	case err == nil:
		// Get related links and comments
		links, _ := h.storage.GetTaskLinks(taskID)
		if links == nil {
			links = []*models.Link{}
		}
		comments, _ := h.storage.GetTaskComments(taskID)
		if comments == nil {
			comments = []*models.Comment{}
		}

		response := map[string]interface{}{
			"id":         task.ID,
			"jira_id":    task.JiraID,
			"title":      task.Title,
			"priority":   task.Priority,
			"status":     task.Status,
			"tags":       task.Tags,
			"blockers":   task.Blockers,
			"created_at": task.CreatedAt,
			"updated_at": task.UpdatedAt,
			"links":      links,
			"comments":   comments,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case strings.Contains(err.Error(), "not found"):
		http.Error(w, "Task not found", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// updateTask fully updates a task
// @Summary Update entire task
// @Description Replace entire task with provided data (PUT)
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID" format(uuid)
// @Param task body models.Task true "Task data"
// @Success 200 {object} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /tasks/{id} [put]
func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request, taskID string) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task.ID = taskID
	err := h.storage.UpdateTask(&task)
	switch {
	case err == nil:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case isValidationError(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// patchTask partially updates a task
// @Summary Update task fields
// @Description Partially update a task with the provided fields (PATCH)
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID" format(uuid)
// @Param task body models.PatchTaskRequest true "Fields to update"
// @Success 200 {object} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /tasks/{id} [patch]
func (h *TaskHandler) patchTask(w http.ResponseWriter, r *http.Request, taskID string) {
	log := logger.FromContext(r.Context())
	log.Debug("Patching task", "task_id", taskID)

	// Get the existing task first
	existingTask, err := h.storage.GetTask(taskID)
	if err != nil {
		log.Error("Failed to get existing task for patch", "error", err, "task_id", taskID)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	// Parse the partial update data
	var patchData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patchData); err != nil {
		log.Error("Failed to decode patch JSON", "error", err, "task_id", taskID)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Debug("Patch data received", "task_id", taskID, "patch_data", patchData)

	// Apply patches to existing task
	if status, ok := patchData["status"]; ok {
		if statusStr, ok := status.(string); ok {
			existingTask.Status = models.Status(statusStr)
			log.Debug("Updated task status", "task_id", taskID, "new_status", statusStr)
		}
	}

	if priority, ok := patchData["priority"]; ok {
		if priorityStr, ok := priority.(string); ok {
			existingTask.Priority = models.Priority(priorityStr)
			log.Debug("Updated task priority", "task_id", taskID, "new_priority", priorityStr)
		}
	}

	if title, ok := patchData["title"]; ok {
		if titleStr, ok := title.(string); ok {
			existingTask.Title = titleStr
		}
	}

	if tags, ok := patchData["tags"]; ok {
		if tagsArray, ok := tags.([]interface{}); ok {
			stringTags := make([]string, len(tagsArray))
			for i, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					stringTags[i] = tagStr
				}
			}
			existingTask.Tags = stringTags
		}
	}

	if blockers, ok := patchData["blockers"]; ok {
		if blockersArray, ok := blockers.([]interface{}); ok {
			stringBlockers := make([]string, len(blockersArray))
			for i, blocker := range blockersArray {
				if blockerStr, ok := blocker.(string); ok {
					stringBlockers[i] = blockerStr
				}
			}
			existingTask.Blockers = stringBlockers
		}
	}

	// Update the task
	err = h.storage.UpdateTask(existingTask)
	if err != nil {
		log.Error("Failed to patch task", "error", err, "task_id", taskID)
		if isValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
		}
		return
	}

	log.Info("Task patched successfully", "task_id", taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTask)
}

// deleteTask removes a task
// @Summary Delete task
// @Description Delete a task by ID
// @Tags tasks
// @Param id path string true "Task ID" format(uuid)
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Router /tasks/{id} [delete]
func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	err := h.storage.DeleteTask(taskID)
	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func isValidationError(err error) bool {
	_, ok := err.(*models.ValidationError)
	return ok
}

// HandleReport generates an automatic status report
func (h *TaskHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Generating automatic report")

	switch r.Method {
	case http.MethodGet:
		h.generateReport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// generateReport creates automatic status report
// @Summary Generate status report
// @Description Generate an automatic status report with working_on, next_up, and blockers sections
// @Tags report
// @Produce json
// @Success 200 {object} models.ReportResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /report [get]
func (h *TaskHandler) generateReport(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	// Get all non-archived tasks
	allFilters := storage.TaskFilters{
		IncludeArchived: false,
	}
	
	allTasks, err := h.storage.ListTasks(allFilters)
	if err != nil {
		log.Error("Failed to get tasks for report", "error", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	report := map[string]interface{}{
		"working_on": []*models.Task{},
		"next_up":    []*models.Task{},
		"blockers":   []*models.Task{},
	}

	// Helper function to get task with links
	getTaskWithLinks := func(task *models.Task) map[string]interface{} {
		links, _ := h.storage.GetTaskLinks(task.ID)
		if links == nil {
			links = []*models.Link{}
		}
		
		return map[string]interface{}{
			"id":         task.ID,
			"jira_id":    task.JiraID,
			"title":      task.Title,
			"priority":   task.Priority,
			"status":     task.Status,
			"tags":       task.Tags,
			"blockers":   task.Blockers,
			"created_at": task.CreatedAt,
			"updated_at": task.UpdatedAt,
			"links":      links,
		}
	}

	workingOn := []map[string]interface{}{}
	nextUp := []map[string]interface{}{}
	blockers := []map[string]interface{}{}

	for _, task := range allTasks {
		if task == nil {
			continue
		}

		taskWithLinks := getTaskWithLinks(task)

		switch task.Status {
		case models.InProgress:
			// All in_progress tasks go to both working_on and next_up
			workingOn = append(workingOn, taskWithLinks)
			nextUp = append(nextUp, taskWithLinks)
			
		case models.Done:
			// All completed tasks go to working_on
			workingOn = append(workingOn, taskWithLinks)
			
		case models.New:
			// All new tasks go to next_up, ordered by priority
			nextUp = append(nextUp, taskWithLinks)
			
		case models.Blocked:
			// All blocked tasks
			blockers = append(blockers, taskWithLinks)
		}
	}

	// Sort next_up by priority (critical > high > normal > minor)
	sort.Slice(nextUp, func(i, j int) bool {
		priorityOrder := map[string]int{
			"critical": 0,
			"high":     1,
			"normal":   2,
			"minor":    3,
		}
		priority1 := nextUp[i]["priority"].(models.Priority)
		priority2 := nextUp[j]["priority"].(models.Priority)
		return priorityOrder[string(priority1)] < priorityOrder[string(priority2)]
	})

	report["working_on"] = workingOn
	report["next_up"] = nextUp
	report["blockers"] = blockers

	log.Debug("Report generated", 
		"working_on_count", len(workingOn),
		"next_up_count", len(nextUp), 
		"blockers_count", len(blockers))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// HandleLinks handles POST requests to create new links
func (h *TaskHandler) HandleLinks(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	switch r.Method {
	case http.MethodPost:
		h.createLink(w, r)
	default:
		log.Debug("Method not allowed for links", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleLink handles individual link operations (GET, PUT, DELETE)
func (h *TaskHandler) HandleLink(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	// Extract link ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/links/")
	if path == "" {
		log.Debug("Link ID required but not provided")
		http.Error(w, "Link ID required", http.StatusBadRequest)
		return
	}

	linkID := strings.Split(path, "/")[0]
	log.Debug("HandleLink called", "link_id", linkID, "method", r.Method)

	switch r.Method {
	case http.MethodGet:
		h.getLink(w, r, linkID)
	case http.MethodPut:
		h.updateLink(w, r, linkID)
	case http.MethodDelete:
		h.deleteLink(w, r, linkID)
	default:
		log.Debug("Method not allowed for link", "method", r.Method, "link_id", linkID)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createLink creates a new link
// @Summary Create a new link
// @Description Create a new link associated with a task
// @Tags links
// @Accept json
// @Produce json
// @Param link body models.CreateLinkRequest true "Link to create"
// @Success 201 {object} models.Link
// @Failure 400 {object} models.ErrorResponse
// @Router /links [post]
func (h *TaskHandler) createLink(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Creating new link")

	var link models.Link
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		log.Error("Failed to decode link JSON", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Debug("Link data received", 
		"task_id", link.TaskID,
		"type", link.Type,
		"url", link.URL,
		"title", link.Title)

	// Validate required fields
	if link.TaskID == "" {
		log.Debug("Missing task_id in link creation")
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}
	if link.URL == "" {
		log.Debug("Missing URL in link creation")
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	if link.Type == "" {
		log.Debug("Missing type in link creation")
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	// Set default values
	if link.Title == "" {
		link.Title = link.URL
	}
	if link.Status == "" {
		link.Status = "active"
	}

	err := h.storage.CreateLink(&link)
	if err != nil {
		log.Error("Failed to create link", "error", err, "task_id", link.TaskID)
		if isValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create link", http.StatusInternalServerError)
		}
		return
	}

	log.Info("Link created successfully", "link_id", link.ID, "task_id", link.TaskID, "type", link.Type)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

// getLink retrieves a specific link
// @Summary Get link by ID
// @Description Retrieve a specific link by its ID
// @Tags links
// @Produce json
// @Param id path string true "Link ID" format(uuid)
// @Success 200 {object} models.Link
// @Failure 404 {object} models.ErrorResponse
// @Router /links/{id} [get]
func (h *TaskHandler) getLink(w http.ResponseWriter, r *http.Request, linkID string) {
	log := logger.FromContext(r.Context())
	log.Debug("Getting link", "link_id", linkID)

	link, err := h.storage.GetLink(linkID)
	if err != nil {
		log.Error("Failed to get link", "error", err, "link_id", linkID)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get link", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(link)
}

// updateLink updates a link
// @Summary Update link
// @Description Update a link with new data
// @Tags links
// @Accept json
// @Produce json
// @Param id path string true "Link ID" format(uuid)
// @Param link body models.Link true "Link data"
// @Success 200 {object} models.Link
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /links/{id} [put]
func (h *TaskHandler) updateLink(w http.ResponseWriter, r *http.Request, linkID string) {
	log := logger.FromContext(r.Context())
	log.Debug("Updating link", "link_id", linkID)

	var link models.Link
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		log.Error("Failed to decode link JSON for update", "error", err, "link_id", linkID)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Ensure the ID matches the URL parameter
	link.ID = linkID

	err := h.storage.UpdateLink(&link)
	if err != nil {
		log.Error("Failed to update link", "error", err, "link_id", linkID)
		if isValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update link", http.StatusInternalServerError)
		}
		return
	}

	log.Info("Link updated successfully", "link_id", linkID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(link)
}

// deleteLink removes a link
// @Summary Delete link
// @Description Delete a link by ID
// @Tags links
// @Param id path string true "Link ID" format(uuid)
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Router /links/{id} [delete]
func (h *TaskHandler) deleteLink(w http.ResponseWriter, r *http.Request, linkID string) {
	log := logger.FromContext(r.Context())
	log.Debug("Deleting link", "link_id", linkID)

	err := h.storage.DeleteLink(linkID)
	if err != nil {
		log.Error("Failed to delete link", "error", err, "link_id", linkID)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete link", http.StatusInternalServerError)
		}
		return
	}

	log.Info("Link deleted successfully", "link_id", linkID)
	w.WriteHeader(http.StatusNoContent)
}

// HandleComments handles comment collection operations
func (h *TaskHandler) HandleComments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createComment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleComment handles individual comment operations
func (h *TaskHandler) HandleComment(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	// Extract comment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	commentID := strings.Split(path, "/")[0]
	
	log.Debug("HandleComment called", "comment_id", commentID, "method", r.Method)
	
	if commentID == "" {
		http.Error(w, "Comment ID required", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case http.MethodDelete:
		h.deleteComment(w, r, commentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createComment creates a new comment
// @Summary Create a new comment
// @Description Create a new comment for a task
// @Tags comments
// @Accept json
// @Produce json
// @Param comment body models.CreateCommentRequest true "Comment to create"
// @Success 200 {object} models.CreateCommentResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /comments [post]
func (h *TaskHandler) createComment(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	var req struct {
		TaskID  string `json:"task_id"`
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	log.Debug("Creating comment", "task_id", req.TaskID, "content_length", len(req.Content))
	
	// Validate input
	if req.TaskID == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}
	
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	
	// Create comment
	comment := &models.Comment{
		TaskID:  req.TaskID,
		Content: strings.TrimSpace(req.Content),
	}
	
	if err := h.storage.CreateComment(comment); err != nil {
		log.Error("Failed to create comment", "error", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}
	
	log.Info("Comment created successfully", "comment_id", comment.ID, "task_id", req.TaskID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      comment.ID,
		"message": "Comment created successfully",
	})
}

// deleteComment removes a comment
// @Summary Delete comment
// @Description Delete a comment by ID
// @Tags comments
// @Param id path string true "Comment ID" format(uuid)
// @Success 200 {object} models.DeleteCommentResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /comments/{id} [delete]
func (h *TaskHandler) deleteComment(w http.ResponseWriter, r *http.Request, commentID string) {
	log := logger.FromContext(r.Context())
	
	log.Debug("Deleting comment", "comment_id", commentID)
	
	if err := h.storage.DeleteComment(commentID); err != nil {
		log.Error("Failed to delete comment", "error", err, "comment_id", commentID)
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}
	
	log.Info("Comment deleted successfully", "comment_id", commentID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Comment deleted successfully",
	})
}