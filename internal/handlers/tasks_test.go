package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"michishirube/internal/handlers/mocks"
	"michishirube/internal/models"
	"michishirube/internal/storage"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions for creating data
func createValidTask() *models.Task {
	return &models.Task{
		ID:        "task-123",
		JiraID:    "TASK-123",
		Title:     "Implementation task",
		Priority:  models.Normal,
		Status:    models.New,
		Tags:      []string{"backend", "api"},
		Blockers:  []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createValidLink() *models.Link {
	return &models.Link{
		ID:     "link-123",
		TaskID: "task-123",
		Type:   models.PullRequest,
		URL:    "https://github.com/company/repo/pull/123",
		Title:  "Fix implementation",
		Status: "open",
	}
}

func createValidComment() *models.Comment {
	return &models.Comment{
		ID:        "comment-123",
		TaskID:    "task-123",
		Content:   "Progress update on implementation",
		CreatedAt: time.Now(),
	}
}

func TestNewTaskHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockStorage, handler.storage)
}

func TestTaskHandler_HandleTasks_GET(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	expectedTasks := []*models.Task{createValidTask()}
	
	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(expectedTasks, nil).
		Times(1)
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, float64(1), response["total"])
	tasks := response["tasks"].([]interface{})
	assert.Len(t, tasks, 1)
}

func TestTaskHandler_HandleTasks_GET_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	expectedTasks := []*models.Task{createValidTask()}
	
	// Verify that filters are parsed correctly
	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		DoAndReturn(func(filters storage.TaskFilters) ([]*models.Task, error) {
			assert.Len(t, filters.Status, 2)
			assert.Contains(t, filters.Status, models.New)
			assert.Contains(t, filters.Status, models.InProgress)
			assert.Len(t, filters.Priority, 1)
			assert.Contains(t, filters.Priority, models.High)
			assert.Equal(t, 10, filters.Limit)
			assert.Equal(t, 5, filters.Offset)
			assert.True(t, filters.IncludeArchived)
			return expectedTasks, nil
		}).
		Times(1)
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?status=new,in_progress&priority=high&limit=10&offset=5&include_archived=true", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_HandleTasks_GET_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(nil, fmt.Errorf("database connection failed")).
		Times(1)
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "database connection failed")
}

func TestTaskHandler_HandleTasks_POST_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	task := createValidTask()
	taskJSON, _ := json.Marshal(task)
	
	mockStorage.EXPECT().
		CreateTask(gomock.Any()).
		DoAndReturn(func(taskArg *models.Task) error {
			// Verify the task data was parsed correctly
			assert.Equal(t, task.Title, taskArg.Title)
			assert.Equal(t, task.JiraID, taskArg.JiraID)
			return nil
		}).
		Times(1)
	
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var createdTask models.Task
	err := json.Unmarshal(w.Body.Bytes(), &createdTask)
	require.NoError(t, err)
	assert.Equal(t, task.Title, createdTask.Title)
}

func TestTaskHandler_HandleTasks_POST_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestTaskHandler_HandleTasks_POST_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	task := createValidTask()
	taskJSON, _ := json.Marshal(task)
	
	validationErr := &models.ValidationError{Message: "Title is required"}
	mockStorage.EXPECT().
		CreateTask(gomock.Any()).
		Return(validationErr).
		Times(1)
	
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Title is required")
}

func TestTaskHandler_HandleTasks_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	req := httptest.NewRequest(http.MethodPatch, "/api/tasks", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestTaskHandler_HandleTask_GET_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	task := createValidTask()
	links := []*models.Link{createValidLink()}
	comments := []*models.Comment{createValidComment()}
	
	mockStorage.EXPECT().GetTask("task-123").Return(task, nil).Times(1)
	mockStorage.EXPECT().GetTaskLinks("task-123").Return(links, nil).Times(1)
	mockStorage.EXPECT().GetTaskComments("task-123").Return(comments, nil).Times(1)
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/task-123", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, task.ID, response["id"])
	assert.Equal(t, task.Title, response["title"])
	assert.NotNil(t, response["links"])
	assert.NotNil(t, response["comments"])
}

func TestTaskHandler_HandleTask_GET_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	mockStorage.EXPECT().
		GetTask("nonexistent").
		Return(nil, fmt.Errorf("task not found")).
		Times(1)
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Task not found")
}

func TestTaskHandler_HandleTask_PUT_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	task := createValidTask()
	task.Title = "Updated title"
	taskJSON, _ := json.Marshal(task)
	
	mockStorage.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(taskArg *models.Task) error {
			// Verify the ID was set correctly and title updated
			assert.Equal(t, "task-123", taskArg.ID)
			assert.Equal(t, "Updated title", taskArg.Title)
			return nil
		}).
		Times(1)
	
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-123", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var updatedTask models.Task
	err := json.Unmarshal(w.Body.Bytes(), &updatedTask)
	require.NoError(t, err)
	assert.Equal(t, "task-123", updatedTask.ID)
	assert.Equal(t, "Updated title", updatedTask.Title)
}

func TestTaskHandler_HandleTask_DELETE_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	mockStorage.EXPECT().
		DeleteTask("task-123").
		Return(nil).
		Times(1)
	
	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/task-123", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestTaskHandler_HandleTask_NoTaskID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Task ID required")
}

func TestTaskHandler_HandleTask_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/task-123", nil)
	w := httptest.NewRecorder()
	
	handler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "validation error",
			err:      &models.ValidationError{Message: "validation failed"},
			expected: true,
		},
		{
			name:     "regular error",
			err:      fmt.Errorf("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkTaskHandler_ListTasks(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	tasks := make([]*models.Task, 100)
	for i := 0; i < 100; i++ {
		task := createValidTask()
		task.ID = fmt.Sprintf("task-%d", i)
		tasks[i] = task
	}
	
	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(tasks, nil).
		AnyTimes()
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.HandleTasks(w, req)
	}
}

func BenchmarkTaskHandler_CreateTask(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)
	
	task := createValidTask()
	taskJSON, _ := json.Marshal(task)
	
	mockStorage.EXPECT().
		CreateTask(gomock.Any()).
		Return(nil).
		AnyTimes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandleTasks(w, req)
	}
}