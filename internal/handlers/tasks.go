package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
		comments, _ := h.storage.GetTaskComments(taskID)

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