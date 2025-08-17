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
	var taskJSON []byte
	var err error
	taskJSON, err = json.Marshal(task)
	require.NoError(t, err)

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
	err = json.Unmarshal(w.Body.Bytes(), &createdTask)
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
	var taskJSON []byte
	var err error
	taskJSON, err = json.Marshal(task)
	require.NoError(t, err)

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
	var taskJSON []byte
	var err error
	taskJSON, err = json.Marshal(task)
	require.NoError(t, err)

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
	err = json.Unmarshal(w.Body.Bytes(), &updatedTask)
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

	req := httptest.NewRequest(http.MethodConnect, "/api/tasks/task-123", nil)
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
	var taskJSON []byte
	var err error
	taskJSON, err = json.Marshal(task)
	require.NoError(b, err)

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

func TestTaskHandler_HandleTask_PATCH_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Create existing task
	existingTask := createValidTask()
	existingTask.ID = "task-123"
	existingTask.Status = models.New
	existingTask.Priority = models.Normal

	// Patch data
	patchData := map[string]interface{}{
		"status":   "in_progress",
		"priority": "high",
		"title":    "Updated title",
		"tags":     []string{"updated", "patch"},
	}

	patchJSON, err := json.Marshal(patchData)
	require.NoError(t, err)

	// Mock expectations
	mockStorage.EXPECT().
		GetTask("task-123").
		Return(existingTask, nil).
		Times(1)

	mockStorage.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(taskArg *models.Task) error {
			// Verify the task was updated correctly
			assert.Equal(t, models.InProgress, taskArg.Status)
			assert.Equal(t, models.High, taskArg.Priority)
			assert.Equal(t, "Updated title", taskArg.Title)
			assert.Equal(t, []string{"updated", "patch"}, taskArg.Tags)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/task-123", bytes.NewBuffer(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var updatedTask models.Task
	err = json.Unmarshal(w.Body.Bytes(), &updatedTask)
	require.NoError(t, err)
	assert.Equal(t, "task-123", updatedTask.ID)
	assert.Equal(t, models.InProgress, updatedTask.Status)
	assert.Equal(t, models.High, updatedTask.Priority)
}

func TestTaskHandler_HandleTask_PATCH_TaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	patchData := map[string]interface{}{
		"status": "in_progress",
	}
	patchJSON, err := json.Marshal(patchData)
	require.NoError(t, err)

	// Mock task not found
	mockStorage.EXPECT().
		GetTask("nonexistent").
		Return(nil, fmt.Errorf("task not found")).
		Times(1)

	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/nonexistent", bytes.NewBuffer(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Task not found")
}

func TestTaskHandler_HandleTask_PATCH_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Create existing task
	existingTask := createValidTask()
	existingTask.ID = "task-123"

	mockStorage.EXPECT().
		GetTask("task-123").
		Return(existingTask, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/task-123", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestTaskHandler_HandleTask_PATCH_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Create existing task
	existingTask := createValidTask()
	existingTask.ID = "task-123"

	patchData := map[string]interface{}{
		"status": "invalid_status",
	}
	patchJSON, err := json.Marshal(patchData)
	require.NoError(t, err)

	mockStorage.EXPECT().
		GetTask("task-123").
		Return(existingTask, nil).
		Times(1)

	validationErr := &models.ValidationError{Message: "Invalid status"}
	mockStorage.EXPECT().
		UpdateTask(gomock.Any()).
		Return(validationErr).
		Times(1)

	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/task-123", bytes.NewBuffer(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid status")
}

func TestTaskHandler_HandleReport_GET_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Create test tasks
	tasks := []*models.Task{
		{
			ID:       "task-1",
			Title:    "In Progress Task",
			Status:   models.InProgress,
			Priority: models.High,
			Tags:     []string{"frontend"},
		},
		{
			ID:       "task-2",
			Title:    "New Task",
			Status:   models.New,
			Priority: models.Critical,
			Tags:     []string{"backend"},
		},
		{
			ID:       "task-3",
			Title:    "Blocked Task",
			Status:   models.Blocked,
			Priority: models.Normal,
			Tags:     []string{"database"},
		},
	}

	// Mock expectations
	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(tasks, nil).
		Times(1)

	// Mock GetTaskLinks for each task
	for _, task := range tasks {
		mockStorage.EXPECT().
			GetTaskLinks(task.ID).
			Return([]*models.Link{}, nil).
			Times(1)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/report", nil)
	w := httptest.NewRecorder()

	handler.HandleReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var report map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	// Verify report structure
	assert.Contains(t, report, "working_on")
	assert.Contains(t, report, "next_up")
	assert.Contains(t, report, "blockers")

	workingOn := report["working_on"].([]interface{})
	nextUp := report["next_up"].([]interface{})
	blockers := report["blockers"].([]interface{})

	// Verify task distribution
	assert.Len(t, workingOn, 1) // Only in_progress tasks
	assert.Len(t, nextUp, 2)    // in_progress + new tasks
	assert.Len(t, blockers, 1)  // blocked tasks
}

func TestTaskHandler_HandleReport_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodPost, "/api/report", nil)
	w := httptest.NewRecorder()

	handler.HandleReport(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestTaskHandler_HandleReport_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Mock storage error
	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(nil, fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/report", nil)
	w := httptest.NewRecorder()

	handler.HandleReport(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to generate report")
}

func TestTaskHandler_HandleLinks_POST_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	link := &models.Link{
		TaskID: "task-123",
		Type:   models.PullRequest,
		URL:    "https://github.com/test/repo/pull/1",
		Title:  "Test PR",
	}

	linkJSON, err := json.Marshal(link)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateLink(gomock.Any()).
		DoAndReturn(func(linkArg *models.Link) error {
			// Verify the link data was parsed correctly
			assert.Equal(t, link.TaskID, linkArg.TaskID)
			assert.Equal(t, link.Type, linkArg.Type)
			assert.Equal(t, link.URL, linkArg.URL)
			assert.Equal(t, link.Title, linkArg.Title)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(linkJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLinks(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTaskHandler_HandleLinks_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()

	handler.HandleLinks(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestTaskHandler_HandleLink_GET_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	link := &models.Link{
		ID:       "link-123",
		TaskID:   "task-123",
		Type:     models.PullRequest,
		URL:      "https://github.com/test/repo/pull/1",
		Title:    "Test PR",
		Status:   "open",
		Metadata: "{}",
	}

	mockStorage.EXPECT().
		GetLink("link-123").
		Return(link, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/links/link-123", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.Link
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, link.ID, response.ID)
	assert.Equal(t, link.Title, response.Title)
}

func TestTaskHandler_HandleLink_NoLinkID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/links/", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Link ID required")
}

func TestTaskHandler_HandleLink_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodConnect, "/api/links/link-123", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestTaskHandler_CreateLink_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Link missing required fields
	link := &models.Link{
		Type: models.PullRequest,
		// Missing TaskID and URL
	}

	linkJSON, err := json.Marshal(link)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(linkJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLinks(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "task_id is required")
}

func TestTaskHandler_CreateLink_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLinks(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestTaskHandler_HandleComments_POST_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	comment := &models.Comment{
		TaskID:  "task-123",
		Content: "This is a test comment",
	}

	commentJSON, err := json.Marshal(comment)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateComment(gomock.Any()).
		DoAndReturn(func(commentArg *models.Comment) error {
			// Verify the comment data was parsed correctly
			assert.Equal(t, comment.TaskID, commentArg.TaskID)
			assert.Equal(t, comment.Content, commentArg.Content)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBuffer(commentJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleComments(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTaskHandler_HandleComments_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	w := httptest.NewRecorder()

	handler.HandleComments(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestTaskHandler_HandleComment_DELETE_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		DeleteComment("comment-123").
		Return(nil).
		Times(1)

	req := httptest.NewRequest(http.MethodDelete, "/api/comments/comment-123", nil)
	w := httptest.NewRecorder()

	handler.HandleComment(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Comment deleted successfully", response["message"])
}

func TestTaskHandler_HandleComment_NoCommentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodDelete, "/api/comments/", nil)
	w := httptest.NewRecorder()

	handler.HandleComment(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Comment ID required")
}

func TestTaskHandler_HandleComment_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodConnect, "/api/comments/comment-123", nil)
	w := httptest.NewRecorder()

	handler.HandleComment(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestTaskHandler_CreateComment_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	// Comment missing required fields
	comment := &models.Comment{
		// Missing TaskID and Content
	}

	commentJSON, err := json.Marshal(comment)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBuffer(commentJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleComments(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "task_id is required")
}

func TestTaskHandler_CreateComment_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodPost, "/api/comments", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleComments(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestTaskHandler_DeleteComment_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		DeleteComment("comment-123").
		Return(fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodDelete, "/api/comments/comment-123", nil)
	w := httptest.NewRecorder()

	handler.HandleComment(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to delete comment")
}

func TestTaskHandler_UpdateTask_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := createValidTask()
	task.Title = "" // Invalid: empty title
	taskJSON, err := json.Marshal(task)
	require.NoError(t, err)

	validationErr := &models.ValidationError{Message: "Title is required"}
	mockStorage.EXPECT().
		UpdateTask(gomock.Any()).
		Return(validationErr).
		Times(1)

	req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-123", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Title is required")
}

func TestTaskHandler_UpdateTask_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := createValidTask()
	taskJSON, err := json.Marshal(task)
	require.NoError(t, err)

	mockStorage.EXPECT().
		UpdateTask(gomock.Any()).
		Return(fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-123", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "database connection failed")
}

func TestTaskHandler_DeleteTask_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		DeleteTask("task-123").
		Return(fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/task-123", nil)
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "database connection failed")
}

func TestTaskHandler_ListTasks_WithComplexFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	expectedTasks := []*models.Task{createValidTask()}

	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		DoAndReturn(func(filters storage.TaskFilters) ([]*models.Task, error) {
			// Verify that complex filters are passed correctly
			assert.True(t, filters.IncludeArchived)
			assert.Equal(t, 50, filters.Limit)
			assert.Equal(t, 10, filters.Offset)
			assert.Equal(t, []models.Status{models.New, models.InProgress}, filters.Status)
			assert.Equal(t, []models.Priority{models.High, models.Critical}, filters.Priority)
			assert.Equal(t, []string{"frontend", "backend"}, filters.Tags)
			return expectedTasks, nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?status=new,in_progress&priority=high,critical&tags=frontend,backend&include_archived=true&limit=50&offset=10", nil)
	w := httptest.NewRecorder()

	handler.HandleTasks(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_ListTasks_InvalidLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	expectedTasks := []*models.Task{createValidTask()}

	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(expectedTasks, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=invalid", nil)
	w := httptest.NewRecorder()

	handler.HandleTasks(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should use default limit when invalid
}

func TestTaskHandler_ListTasks_InvalidOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	expectedTasks := []*models.Task{createValidTask()}

	mockStorage.EXPECT().
		ListTasks(gomock.Any()).
		Return(expectedTasks, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?offset=invalid", nil)
	w := httptest.NewRecorder()

	handler.HandleTasks(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should use default offset when invalid
}

func TestTaskHandler_UpdateLink_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	link := &models.Link{
		ID:       "link-123",
		TaskID:   "task-123",
		Type:     models.PullRequest,
		URL:      "https://github.com/test/repo/pull/2",
		Title:    "Updated PR",
		Status:   "merged",
		Metadata: "{}",
	}

	linkJSON, err := json.Marshal(link)
	require.NoError(t, err)

	mockStorage.EXPECT().
		UpdateLink(gomock.Any()).
		DoAndReturn(func(linkArg *models.Link) error {
			// Verify the link data was parsed correctly
			assert.Equal(t, link.ID, linkArg.ID)
			assert.Equal(t, link.Title, linkArg.Title)
			assert.Equal(t, link.Status, linkArg.Status)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPut, "/api/links/link-123", bytes.NewBuffer(linkJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTaskHandler_UpdateLink_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	req := httptest.NewRequest(http.MethodPut, "/api/links/link-123", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestTaskHandler_UpdateLink_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	link := &models.Link{
		ID:     "link-123",
		TaskID: "task-123",
		Type:   models.PullRequest,
		URL:    "", // Invalid: empty URL
		Title:  "Updated PR",
	}

	linkJSON, err := json.Marshal(link)
	require.NoError(t, err)

	validationErr := &models.ValidationError{Message: "URL is required"}
	mockStorage.EXPECT().
		UpdateLink(gomock.Any()).
		Return(validationErr).
		Times(1)

	req := httptest.NewRequest(http.MethodPut, "/api/links/link-123", bytes.NewBuffer(linkJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "URL is required")
}

func TestTaskHandler_DeleteLink_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		DeleteLink("link-123").
		Return(nil).
		Times(1)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/link-123", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "", w.Header().Get("Content-Type"))
	assert.Empty(t, w.Body.String())
}

func TestTaskHandler_DeleteLink_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		DeleteLink("link-123").
		Return(fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/link-123", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to delete link")
}

func TestTaskHandler_GetLink_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		GetLink("nonexistent").
		Return(nil, fmt.Errorf("link not found")).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/links/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Link not found")
}

func TestTaskHandler_GetLink_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	mockStorage.EXPECT().
		GetLink("link-123").
		Return(nil, fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/links/link-123", nil)
	w := httptest.NewRecorder()

	handler.HandleLink(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get link")
}

func TestTaskHandler_CreateLink_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	link := &models.Link{
		TaskID: "task-123",
		Type:   models.PullRequest,
		URL:    "https://github.com/test/repo/pull/1",
		Title:  "Test PR",
	}

	linkJSON, err := json.Marshal(link)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateLink(gomock.Any()).
		Return(fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(linkJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLinks(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create link")
}

func TestTaskHandler_CreateComment_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	comment := &models.Comment{
		TaskID:  "task-123",
		Content: "This is a test comment",
	}

	commentJSON, err := json.Marshal(comment)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateComment(gomock.Any()).
		Return(fmt.Errorf("database connection failed")).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBuffer(commentJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleComments(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create comment")
}

func TestTaskHandler_GetTask_WithLinksAndComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := createValidTask()
	links := []*models.Link{
		{
			ID:     "link-1",
			TaskID: "task-123",
			Type:   models.PullRequest,
			URL:    "https://github.com/test/repo/pull/1",
			Title:  "Test PR 1",
			Status: "open",
		},
		{
			ID:     "link-2",
			TaskID: "task-123",
			Type:   models.JiraTicket,
			URL:    "https://github.com/test/repo/issues/1",
			Title:  "Test Issue 1",
			Status: "open",
		},
	}
	comments := []*models.Comment{
		{
			ID:      "comment-1",
			TaskID:  "task-123",
			Content: "First comment",
		},
		{
			ID:      "comment-2",
			TaskID:  "task-123",
			Content: "Second comment",
		},
	}

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

	// Verify links
	responseLinks := response["links"].([]interface{})
	assert.Len(t, responseLinks, 2)
	assert.Equal(t, "link-1", responseLinks[0].(map[string]interface{})["id"])
	assert.Equal(t, "link-2", responseLinks[1].(map[string]interface{})["id"])

	// Verify comments
	responseComments := response["comments"].([]interface{})
	assert.Len(t, responseComments, 2)
	assert.Equal(t, "First comment", responseComments[0].(map[string]interface{})["content"])
	assert.Equal(t, "Second comment", responseComments[1].(map[string]interface{})["content"])
}

func TestTaskHandler_GetTask_StorageErrorOnLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := createValidTask()

	mockStorage.EXPECT().GetTask("task-123").Return(task, nil).Times(1)
	mockStorage.EXPECT().GetTaskLinks("task-123").Return(nil, fmt.Errorf("database error")).Times(1)
	mockStorage.EXPECT().GetTaskComments("task-123").Return([]*models.Comment{}, nil).Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/task-123", nil)
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should still return the task with empty links
	assert.Equal(t, task.ID, response["id"])
	assert.Empty(t, response["links"])
}

func TestTaskHandler_GetTask_StorageErrorOnComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := createValidTask()
	links := []*models.Link{createValidLink()}

	mockStorage.EXPECT().GetTask("task-123").Return(task, nil).Times(1)
	mockStorage.EXPECT().GetTaskLinks("task-123").Return(links, nil).Times(1)
	mockStorage.EXPECT().GetTaskComments("task-123").Return(nil, fmt.Errorf("database error")).Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/task-123", nil)
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should still return the task with empty comments
	assert.Equal(t, task.ID, response["id"])
	assert.Empty(t, response["comments"])
}

func TestTaskHandler_CreateTask_WithAllFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := &models.Task{
		Title:    "Complete Task",
		Priority: models.Critical,
		Status:   models.New,
		Tags:     []string{"frontend", "urgent", "bug"},
		Blockers: []string{"dependency-issue"},
		JiraID:   "PROJ-123",
	}

	taskJSON, err := json.Marshal(task)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateTask(gomock.Any()).
		DoAndReturn(func(taskArg *models.Task) error {
			// Verify all fields were parsed correctly
			assert.Equal(t, task.Title, taskArg.Title)
			assert.Equal(t, task.Priority, taskArg.Priority)
			assert.Equal(t, task.Status, taskArg.Status)
			assert.Equal(t, task.Tags, taskArg.Tags)
			assert.Equal(t, task.Blockers, taskArg.Blockers)
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
}

func TestTaskHandler_UpdateTask_WithAllFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	task := &models.Task{
		Title:    "Updated Complete Task",
		Priority: models.High,
		Status:   models.InProgress,
		Tags:     []string{"backend", "feature", "in-progress"},
		Blockers: []string{"code-review"},
		JiraID:   "PROJ-456",
	}

	taskJSON, err := json.Marshal(task)
	require.NoError(t, err)

	mockStorage.EXPECT().
		UpdateTask(gomock.Any()).
		DoAndReturn(func(taskArg *models.Task) error {
			// Verify all fields were parsed correctly
			assert.Equal(t, "task-123", taskArg.ID) // Should be set from URL
			assert.Equal(t, task.Title, taskArg.Title)
			assert.Equal(t, task.Priority, taskArg.Priority)
			assert.Equal(t, task.Status, taskArg.Status)
			assert.Equal(t, task.Tags, taskArg.Tags)
			assert.Equal(t, task.Blockers, taskArg.Blockers)
			assert.Equal(t, task.JiraID, taskArg.JiraID)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-123", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleTask(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTaskHandler_CreateLink_WithAllFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	link := &models.Link{
		TaskID:   "task-123",
		Type:     models.JiraTicket,
		URL:      "https://github.com/test/repo/issues/123",
		Title:    "Critical Bug Issue",
		Status:   "open",
		Metadata: `{"priority": "high", "assignee": "developer"}`,
	}

	linkJSON, err := json.Marshal(link)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateLink(gomock.Any()).
		DoAndReturn(func(linkArg *models.Link) error {
			// Verify all fields were parsed correctly
			assert.Equal(t, link.TaskID, linkArg.TaskID)
			assert.Equal(t, link.Type, linkArg.Type)
			assert.Equal(t, link.URL, linkArg.URL)
			assert.Equal(t, link.Title, linkArg.Title)
			assert.Equal(t, link.Status, linkArg.Status)
			assert.Equal(t, link.Metadata, linkArg.Metadata)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(linkJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLinks(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTaskHandler_CreateComment_WithAllFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := NewTaskHandler(mockStorage)

	comment := &models.Comment{
		TaskID:  "task-123",
		Content: "This is a detailed comment with multiple lines\nand special characters: !@#$%^&*()",
	}

	commentJSON, err := json.Marshal(comment)
	require.NoError(t, err)

	mockStorage.EXPECT().
		CreateComment(gomock.Any()).
		DoAndReturn(func(commentArg *models.Comment) error {
			// Verify the comment data was parsed correctly
			assert.Equal(t, comment.TaskID, commentArg.TaskID)
			assert.Equal(t, comment.Content, commentArg.Content)
			return nil
		}).
		Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBuffer(commentJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleComments(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
