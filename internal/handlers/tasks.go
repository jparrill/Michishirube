package handlers

import (
	"encoding/json"
	"net/http"
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
	case http.MethodDelete:
		h.deleteTask(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

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