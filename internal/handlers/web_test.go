package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"michishirube/internal/models"
	"michishirube/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage implements storage.Storage for testing
type MockWebStorage struct {
	tasks    map[string]*models.Task
	links    map[string][]*models.Link
	comments map[string][]*models.Comment
}

func NewMockWebStorage() *MockWebStorage {
	return &MockWebStorage{
		tasks:    make(map[string]*models.Task),
		links:    make(map[string][]*models.Link),
		comments: make(map[string][]*models.Comment),
	}
}

func (m *MockWebStorage) CreateTask(task *models.Task) error {
	if task.Title == "" {
		return &models.ValidationError{Message: "Title is required"}
	}
	task.ID = "task-" + fmt.Sprintf("%d", len(m.tasks)+1)
	m.tasks[task.ID] = task
	return nil
}

func (m *MockWebStorage) GetTask(id string) (*models.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (m *MockWebStorage) ListTasks(filters storage.TaskFilters) ([]*models.Task, error) {
	var tasks []*models.Task
	for _, task := range m.tasks {
		if filters.IncludeArchived || task.Status != models.Archived {
			tasks = append(tasks, task)
		}
	}

	// Apply offset and limit
	start := filters.Offset
	if start >= len(tasks) {
		return []*models.Task{}, nil
	}

	end := start + filters.Limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[start:end], nil
}

func (m *MockWebStorage) SearchTasks(query string, includeArchived bool, limit int) ([]*models.Task, error) {
	var results []*models.Task
	for _, task := range m.tasks {
		if !includeArchived && task.Status == models.Archived {
			continue
		}
		if strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.JiraID), strings.ToLower(query)) {
			results = append(results, task)
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (m *MockWebStorage) GetTaskLinks(taskID string) ([]*models.Link, error) {
	links, exists := m.links[taskID]
	if !exists {
		return []*models.Link{}, nil
	}
	return links, nil
}

func (m *MockWebStorage) GetTaskComments(taskID string) ([]*models.Comment, error) {
	comments, exists := m.comments[taskID]
	if !exists {
		return []*models.Comment{}, nil
	}
	return comments, nil
}

func (m *MockWebStorage) GetComment(id string) (*models.Comment, error) {
	return nil, nil
}

func (m *MockWebStorage) RunMigrations() error {
	return nil
}

// Implement other required methods with minimal functionality
func (m *MockWebStorage) UpdateTask(task *models.Task) error { return nil }
func (m *MockWebStorage) DeleteTask(id string) error         { return nil }
func (m *MockWebStorage) CreateLink(link *models.Link) error {
	if link.TaskID == "" {
		return &models.ValidationError{Message: "TaskID is required"}
	}
	m.links[link.TaskID] = append(m.links[link.TaskID], link)
	return nil
}
func (m *MockWebStorage) GetLink(id string) (*models.Link, error) { return nil, nil }
func (m *MockWebStorage) UpdateLink(link *models.Link) error      { return nil }
func (m *MockWebStorage) DeleteLink(id string) error              { return nil }
func (m *MockWebStorage) CreateComment(comment *models.Comment) error {
	if comment.TaskID == "" {
		return &models.ValidationError{Message: "TaskID is required"}
	}
	m.comments[comment.TaskID] = append(m.comments[comment.TaskID], comment)
	return nil
}
func (m *MockWebStorage) DeleteComment(id string) error { return nil }
func (m *MockWebStorage) Close() error                  { return nil }

// Helper function to create test context
func createTestContext() context.Context {
	return context.Background()
}

// Helper function to create test request
func createTestRequest(method, url string, body string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req = req.WithContext(createTestContext())
	return req
}

// Helper function to create handler using a symlink to real templates when templates are not present
func createHandlerWithSymlinkTemplates(t *testing.T) *WebHandler {
	// Create a temporary working directory
	tempDir, err := os.MkdirTemp("", "test_symlink_templates")
	require.NoError(t, err)

	// Ensure temp web directory exists
	webDir := filepath.Join(tempDir, "web")
	err = os.MkdirAll(webDir, 0755)
	require.NoError(t, err)

	// Resolve source templates directory by walking up from current working dir
	findTemplates := func(start string) (string, bool) {
		d := start
		for i := 0; i < 6; i++ {
			candidate := filepath.Join(d, "web", "templates")
			if st, err := os.Stat(candidate); err == nil && st.IsDir() {
				return candidate, true
			}
			parent := filepath.Dir(d)
			if parent == d {
				break
			}
			d = parent
		}
		return "", false
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)
	srcTemplates, ok := findTemplates(cwd)
	if !ok {
		require.FailNowf(t, "templates not found", "could not locate web/templates from %s", cwd)
	}

	// Create symlink at tempDir/web/templates pointing to repo templates
	dstTemplates := filepath.Join(webDir, "templates")
	// In case it already exists from prior runs
	_ = os.RemoveAll(dstTemplates)
	err = os.Symlink(srcTemplates, dstTemplates)
	require.NoError(t, err)

	// Change to temp directory so NewWebHandler parses tempDir/web/templates
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Cleanup and restore after test
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
		_ = os.RemoveAll(tempDir)
	})

	mockStorage := NewMockWebStorage()
	handler := NewWebHandler(mockStorage)
	return handler
}

// Helper function to create test handler with templates
func createTestHandler(t *testing.T) *WebHandler {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "test_templates")
	require.NoError(t, err)

	// Create web/templates directory
	templatesDir := filepath.Join(tempDir, "web", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create test templates
	templates := map[string]string{
		"base.html": `<!DOCTYPE html>
<html>
<head>
	<title>{{.PageTitle}}</title>
	{{if .CustomJS}}<script src="/static/js/{{.CustomJS}}"></script>{{end}}
</head>
<body>
	<header>
		<h1>Michishirube</h1>
	</header>
	<main>
		{{template "content" .}}
	</main>
</body>
</html>`,
		"test.html": `{{define "content"}}
<div>{{.PageTitle}}</div>
{{end}}`,
		"new_task.html": `{{define "content"}}
<div>
	<h1>{{.PageTitle}}</h1>
	<form method="POST">
		<input type="text" name="title" placeholder="Task title">
		<select name="priority">
			<option value="low">Low</option>
			<option value="normal" selected>Normal</option>
			<option value="high">High</option>
		</select>
		<textarea name="notes" placeholder="Notes"></textarea>
		<button type="submit">Create Task</button>
	</form>
</div>
{{end}}`,
	}

	for filename, content := range templates {
		err = os.WriteFile(filepath.Join(templatesDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Change to temp directory temporarily
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	mockStorage := NewMockWebStorage()
	handler := NewWebHandler(mockStorage)

	// Restore original directory
	err = os.Chdir(originalDir)
	require.NoError(t, err)

	// Clean up temp directory after test
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	})

	return handler
}

func TestNewWebHandler(t *testing.T) {
	// Create a temporary directory for templates
	tempDir, err := os.MkdirTemp("", "test_templates")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create web/templates directory
	templatesDir := filepath.Join(tempDir, "web", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create a simple test template
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.PageTitle}}</title></head>
<body>{{.PageTitle}}</body>
</html>`

	err = os.WriteFile(filepath.Join(templatesDir, "test.html"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Change to temp directory temporarily
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create mock storage
	mockStorage := NewMockWebStorage()

	// Test that NewWebHandler doesn't panic
	handler := NewWebHandler(mockStorage)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.storage)
	assert.NotNil(t, handler.templates)
}

func TestWebHandler_HealthCheck(t *testing.T) {
	handler := createTestHandler(t)

	req := createTestRequest(http.MethodGet, "/health", "")
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, `"status": "healthy"`)
	assert.Contains(t, body, `"timestamp"`)
}

func TestWebHandler_OpenAPISpec_Success(t *testing.T) {
	// Create a temporary directory for templates and docs
	tempDir, err := os.MkdirTemp("", "test_docs")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create web/templates directory
	templatesDir := filepath.Join(tempDir, "web", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create a simple test template
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.PageTitle}}</title></head>
<body>{{.PageTitle}}</body>
</html>`

	err = os.WriteFile(filepath.Join(templatesDir, "test.html"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Create docs directory and swagger.yaml file
	docsDir := filepath.Join(tempDir, "docs")
	err = os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	swaggerContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0`

	err = os.WriteFile(filepath.Join(docsDir, "swagger.yaml"), []byte(swaggerContent), 0644)
	require.NoError(t, err)

	// Change to temp directory temporarily
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	mockStorage := NewMockWebStorage()
	handler := NewWebHandler(mockStorage)

	req := createTestRequest(http.MethodGet, "/openapi.yaml", "")
	w := httptest.NewRecorder()

	handler.OpenAPISpec(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/yaml", w.Header().Get("Content-Type"))
	assert.Equal(t, "application/yaml", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "openapi: 3.0.0")
	assert.Contains(t, body, "Test API")
}

func TestWebHandler_OpenAPISpec_NotFound(t *testing.T) {
	// Create a temporary directory without docs
	tempDir, err := os.MkdirTemp("", "test_docs")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create web/templates directory
	templatesDir := filepath.Join(tempDir, "web", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create a simple test template
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.PageTitle}}</title></head>
<body>{{.PageTitle}}</body>
</html>`

	err = os.WriteFile(filepath.Join(templatesDir, "test.html"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Change to temp directory temporarily
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	mockStorage := NewMockWebStorage()
	handler := NewWebHandler(mockStorage)

	req := createTestRequest(http.MethodGet, "/openapi.yaml", "")
	w := httptest.NewRecorder()

	handler.OpenAPISpec(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "OpenAPI specification not found")
}

func TestWebHandler_SwaggerUI(t *testing.T) {
	handler := createTestHandler(t)

	req := createTestRequest(http.MethodGet, "/docs", "")
	w := httptest.NewRecorder()

	handler.SwaggerUI(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "<!DOCTYPE html>")
	assert.Contains(t, body, "Michishirube API Documentation")
	assert.Contains(t, body, "swagger-ui")
	assert.Contains(t, body, "openapi.yaml")
}

func TestWebHandler_NewTask_GET(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	req := createTestRequest(http.MethodGet, "/new", "")
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebHandler_NewTask_MethodNotAllowed(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	req := createTestRequest(http.MethodPut, "/new", "")
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestWebHandler_CreateNewTask_Success(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	formData := "title=Test Task&priority=high&tags=frontend,urgent&notes=Test notes"
	req := createTestRequest(http.MethodPost, "/new", formData)
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	// Should redirect or show success (depending on implementation)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestWebHandler_CreateNewTask_MissingTitle(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	formData := "priority=high&tags=frontend,urgent"
	req := createTestRequest(http.MethodPost, "/new", formData)
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	// Should return 400 for missing title
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Title is required")
}

func TestWebHandler_CreateNewTask_WithJiraID(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	formData := "jira_id=PROJ-123&title=Test Task&priority=normal"
	req := createTestRequest(http.MethodPost, "/new", formData)
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	// Should succeed
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestWebHandler_CreateNewTask_DefaultJiraID(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	formData := "title=Test Task&priority=normal"
	req := createTestRequest(http.MethodPost, "/new", formData)
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	// Should succeed with default Jira ID
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestWebHandler_CreateNewTask_InvalidFormData(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	// Send malformed form data that can't be parsed
	req := httptest.NewRequest(http.MethodPost, "/new", strings.NewReader("invalid form data"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(createTestContext())
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	// Should return 400 for invalid form data
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// The actual error message depends on how Go's ParseForm handles malformed data
	// It might return "Title is required" if it can parse it as an empty form
	assert.Contains(t, w.Body.String(), "Title is required")
}

func TestWebHandler_StaticFileHandler(t *testing.T) {
	handler := createTestHandler(t)

	// Test that StaticFileHandler returns a valid handler
	staticHandler := handler.StaticFileHandler()
	assert.NotNil(t, staticHandler)

	// Test that it's a http.Handler
	_ = staticHandler
}

func TestWebHandler_SwaggerJSON(t *testing.T) {
	// Create a temporary directory for templates and docs
	tempDir, err := os.MkdirTemp("", "test_docs")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create web/templates directory
	templatesDir := filepath.Join(tempDir, "web", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create a simple test template
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.PageTitle}}</title></head>
<body>{{.PageTitle}}</body>
</html>`

	err = os.WriteFile(filepath.Join(templatesDir, "test.html"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Create docs directory and swagger.json file
	docsDir := filepath.Join(tempDir, "docs")
	err = os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	swaggerContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`

	err = os.WriteFile(filepath.Join(docsDir, "swagger.json"), []byte(swaggerContent), 0644)
	require.NoError(t, err)

	// Change to temp directory temporarily
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	mockStorage := NewMockWebStorage()
	handler := NewWebHandler(mockStorage)

	req := createTestRequest(http.MethodGet, "/swagger.json", "")
	w := httptest.NewRecorder()

	handler.SwaggerJSON(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "openapi")
	assert.Contains(t, body, "Test API")
}

// Test helper function for creating a simple handler without complex templates (unused)
/* func createSimpleTestHandler(t *testing.T) *WebHandler {
	// Create a temporary directory for templates
	tempDir, err := os.MkdirTemp("", "test_templates")
	require.NoError(t, err)

	// Create web/templates directory
	templatesDir := filepath.Join(tempDir, "web", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create base.html and new_task.html templates
	templates := map[string]string{
		"base.html": `<!DOCTYPE html>
<html>
<head>
	<title>{{.PageTitle}}</title>
	{{if .CustomJS}}<script src="/static/js/{{.CustomJS}}"></script>{{end}}
</head>
<body>
	<header>
		<h1>Michishirube</h1>
	</header>
	<main>
		{{template "content" .}}
	</main>
</body>
</html>`,
		"new_task.html": `{{define "content"}}
<div>
	<h1>{{.PageTitle}}</h1>
	<form method="POST">
		<input type="text" name="title" placeholder="Task title">
		<select name="priority">
			<option value="low">Low</option>
			<option value="normal" selected>Normal</option>
			<option value="high">High</option>
		</select>
		<textarea name="notes" placeholder="Notes"></textarea>
		<button type="submit">Create Task</button>
	</form>
</div>
{{end}}`,
	}

	for filename, content := range templates {
		err = os.WriteFile(filepath.Join(templatesDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Change to temp directory temporarily
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	mockStorage := NewMockWebStorage()
	handler := NewWebHandler(mockStorage)

	// Keep the handler in the temp directory context
	// Don't restore original directory yet

	// Clean up temp directory after test
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	})

	return handler
}
*/

func TestWebHandler_NewTask_GET_Simple(t *testing.T) {
	handler := createHandlerWithSymlinkTemplates(t)

	// Debug: check current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Logf("Current working directory: %s", cwd)

	// Debug: check if templates directory exists
	templatesDir := filepath.Join(cwd, "web", "templates")
	t.Logf("Templates directory: %s", templatesDir)

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Logf("Templates directory does not exist")
	} else {
		files, err := os.ReadDir(templatesDir)
		if err != nil {
			t.Logf("Error reading templates directory: %v", err)
		} else {
			t.Logf("Files in templates directory:")
			for _, file := range files {
				t.Logf("  - %s", file.Name())
			}
		}
	}

	req := createTestRequest(http.MethodGet, "/new", "")
	w := httptest.NewRecorder()

	handler.NewTask(w, req)

	// Log the response for debugging
	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	// Should not be 500 (internal server error)
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}
