package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"michishirube/internal/config"
	"michishirube/internal/handlers"
	"michishirube/internal/logger"
	"michishirube/internal/models"
	"michishirube/internal/storage/sqlite"
	"michishirube/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite holds the test setup for integration tests
type IntegrationTestSuite struct {
	storage     *sqlite.SQLiteStorage
	taskHandler *handlers.TaskHandler
	tempDBPath  string
}

// setupIntegrationTest creates a complete test environment
func setupIntegrationTest(t *testing.T) (*IntegrationTestSuite, func()) {
	t.Helper()
	
	// Create temporary database file
	tempDir := t.TempDir()
	tempDBPath := filepath.Join(tempDir, "integration_test.db")
	
	// Initialize storage with real SQLite database
	storage, err := sqlite.New(tempDBPath)
	require.NoError(t, err, "Failed to create test database")
	
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(storage)
	
	suite := &IntegrationTestSuite{
		storage:     storage,
		taskHandler: taskHandler,
		tempDBPath:  tempDBPath,
	}
	
	cleanup := func() {
		storage.Close()
		os.Remove(tempDBPath)
	}
	
	return suite, cleanup
}

// loadFixtureData loads test data from fixtures
func (suite *IntegrationTestSuite) loadFixtureData(t *testing.T) {
	t.Helper()
	
	// Load fixture data
	var tasks []models.Task
	testdata.LoadFixtures(t, "tasks.json", &tasks)
	
	var links []models.Link
	testdata.LoadFixtures(t, "links.json", &links)
	
	var comments []models.Comment
	testdata.LoadFixtures(t, "comments.json", &comments)
	
	// Insert fixture data
	for _, task := range tasks {
		err := suite.storage.CreateTask(&task)
		require.NoError(t, err, "Failed to insert fixture task: %s", task.ID)
	}
	
	for _, link := range links {
		err := suite.storage.CreateLink(&link)
		require.NoError(t, err, "Failed to insert fixture link: %s", link.ID)
	}
	
	for _, comment := range comments {
		err := suite.storage.CreateComment(&comment)
		require.NoError(t, err, "Failed to insert fixture comment: %s", comment.ID)
	}
}

func TestIntegration_FullTaskWorkflow(t *testing.T) {
	suite, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	// Test creating a task via HTTP API
	newTask := &models.Task{
		JiraID:   "INTEGRATION-001",
		Title:    "Integration task workflow",
		Priority: models.High,
		Status:   models.New,
		Tags:     []string{"integration", "workflow"},
		Blockers: []string{},
	}
	
	taskJSON, err := json.Marshal(newTask)
	require.NoError(t, err)
	
	// Create task via HTTP POST
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(taskJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.taskHandler.HandleTasks(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var createdTask models.Task
	err = json.Unmarshal(w.Body.Bytes(), &createdTask)
	require.NoError(t, err)
	
	taskID := createdTask.ID
	assert.NotEmpty(t, taskID)
	assert.Equal(t, newTask.Title, createdTask.Title)
	assert.Equal(t, newTask.JiraID, createdTask.JiraID)
	
	// Verify task was stored in database
	storedTask, err := suite.storage.GetTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, createdTask.Title, storedTask.Title)
	
	// Update task status via HTTP PUT
	createdTask.Status = models.InProgress
	createdTask.Title = "Updated integration task workflow"
	
	updateJSON, err := json.Marshal(createdTask)
	require.NoError(t, err)
	
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/"+taskID, bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.taskHandler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var updatedTask models.Task
	err = json.Unmarshal(w.Body.Bytes(), &updatedTask)
	require.NoError(t, err)
	assert.Equal(t, models.InProgress, updatedTask.Status)
	assert.Equal(t, "Updated integration task workflow", updatedTask.Title)
	
	// Get task with relations via HTTP GET
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+taskID, nil)
	w = httptest.NewRecorder()
	
	suite.taskHandler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var taskResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &taskResponse)
	require.NoError(t, err)
	
	assert.Equal(t, taskID, taskResponse["id"])
	assert.Equal(t, "Updated integration task workflow", taskResponse["title"])
	assert.Equal(t, "in_progress", taskResponse["status"])
	assert.NotNil(t, taskResponse["links"])
	assert.NotNil(t, taskResponse["comments"])
	
	// Delete task via HTTP DELETE
	req = httptest.NewRequest(http.MethodDelete, "/api/tasks/"+taskID, nil)
	w = httptest.NewRecorder()
	
	suite.taskHandler.HandleTask(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	
	// Verify task was deleted from database
	_, err = suite.storage.GetTask(taskID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIntegration_TaskListingWithFilters(t *testing.T) {
	suite, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	// Load fixture data for filtering tests
	suite.loadFixtureData(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedCount  int
		expectedStatus []models.Status
	}{
		{
			name:          "all tasks (excluding archived)",
			queryParams:   "",
			expectedCount: 4, // From fixtures (5 total, 1 archived)
		},
		{
			name:           "high priority tasks",
			queryParams:    "priority=high",
			expectedCount:  1,
		},
		{
			name:           "in progress tasks",
			queryParams:    "status=in_progress",
			expectedCount:  1,
			expectedStatus: []models.Status{models.InProgress},
		},
		{
			name:          "include archived tasks",
			queryParams:   "include_archived=true",
			expectedCount: 5,
		},
		{
			name:          "limit results",
			queryParams:   "limit=2",
			expectedCount: 2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/tasks"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			
			suite.taskHandler.HandleTasks(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			tasks := response["tasks"].([]interface{})
			assert.Len(t, tasks, tt.expectedCount)
			
			// Verify status filtering if specified
			if len(tt.expectedStatus) > 0 {
				for _, taskInterface := range tasks {
					task := taskInterface.(map[string]interface{})
					actualStatus := task["status"].(string)
					assert.Contains(t, tt.expectedStatus, models.Status(actualStatus))
				}
			}
		})
	}
}

func TestIntegration_SearchFunctionality(t *testing.T) {
	suite, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	// Load search scenario data
	var scenario struct {
		Name            string        `json:"name"`
		Description     string        `json:"description"`
		Tasks           []models.Task `json:"tasks"`
		SearchCases []struct {
			Query           string   `json:"query"`
			ExpectedMatches []string `json:"expected_matches"`
			Description     string   `json:"description"`
		} `json:"search_test_cases"`
	}
	
	testdata.LoadScenario(t, "search_dataset", &scenario)
	
	// Insert scenario tasks
	for _, task := range scenario.Tasks {
		err := suite.storage.CreateTask(&task)
		require.NoError(t, err)
	}
	
	// Test each search case
	for _, searchCase := range scenario.SearchCases {
		t.Run(searchCase.Description, func(t *testing.T) {
			results, err := suite.storage.SearchTasks(searchCase.Query, false, 10)
			require.NoError(t, err)
			
			var resultIDs []string
			for _, task := range results {
				resultIDs = append(resultIDs, task.ID)
			}
			
			assert.ElementsMatch(t, searchCase.ExpectedMatches, resultIDs,
				"Search for '%s' should return expected task IDs", searchCase.Query)
		})
	}
}

func TestIntegration_DatabaseMigrations(t *testing.T) {
	// Test that migrations work correctly on a fresh database
	tempDir := t.TempDir()
	tempDBPath := filepath.Join(tempDir, "migration_test.db")
	
	// Create storage - this should run migrations automatically
	storage, err := sqlite.New(tempDBPath)
	require.NoError(t, err)
	defer storage.Close()
	
	// Verify that tables were created by attempting to create a task
	task := &models.Task{
		JiraID:   "MIGRATION-001",
		Title:    "Migration verification task",
		Priority: models.Normal,
		Status:   models.New,
		Tags:     []string{"migration"},
		Blockers: []string{},
	}
	
	err = storage.CreateTask(task)
	assert.NoError(t, err, "Should be able to create task after migrations")
	
	// Verify task can be retrieved
	retrievedTask, err := storage.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.Title, retrievedTask.Title)
}

func TestIntegration_ConfigurationLoading(t *testing.T) {
	// Test configuration loading with logger integration
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "integration_config.yaml")
	
	// Create a config file
	configContent := `
port: "9999"
db_path: "integration_test.db" 
log_level: "debug"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Set environment to point to our config
	originalConfigPath := os.Getenv("CONFIG_PATH")
	os.Setenv("CONFIG_PATH", configPath)
	defer func() {
		if originalConfigPath == "" {
			os.Unsetenv("CONFIG_PATH")
		} else {
			os.Setenv("CONFIG_PATH", originalConfigPath)
		}
	}()
	
	// Create logger context
	ctx := context.Background()
	testLogger := logger.NewLogger(slog.LevelDebug)
	ctx = logger.WithLogger(ctx, testLogger)
	
	// Load configuration
	cfg, err := config.Load(ctx)
	require.NoError(t, err)
	
	assert.Equal(t, "9999", cfg.Port)
	assert.Equal(t, "integration_test.db", cfg.DBPath)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	suite, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid JSON in POST",
			method:         http.MethodPost,
			url:            "/api/tasks",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON",
		},
		{
			name:           "task not found",
			method:         http.MethodGet,
			url:            "/api/tasks/nonexistent-id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Task not found",
		},
		{
			name:           "method not allowed",
			method:         http.MethodPatch,
			url:            "/api/tasks",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "validation error",
			method:         http.MethodPost,
			url:            "/api/tasks",
			body:           `{"jira_id": "", "title": "", "priority": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}
			
			w := httptest.NewRecorder()
			
			if tt.url == "/api/tasks" {
				suite.taskHandler.HandleTasks(w, req)
			} else {
				suite.taskHandler.HandleTask(w, req)
			}
			
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestIntegration_PerformanceWithLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	suite, cleanup := setupIntegrationTest(t)
	defer cleanup()
	
	// Create a large dataset
	const numTasks = 1000
	tasks := make([]*models.Task, numTasks)
	
	start := time.Now()
	
	// Batch insert tasks
	for i := 0; i < numTasks; i++ {
		task := &models.Task{
			JiraID:   fmt.Sprintf("PERF-%04d", i),
			Title:    fmt.Sprintf("Performance verification task %d", i),
			Priority: models.Priority([]models.Priority{models.Minor, models.Normal, models.High, models.Critical}[i%4]),
			Status:   models.Status([]models.Status{models.New, models.InProgress, models.Done}[i%3]),
			Tags:     []string{fmt.Sprintf("perf-%d", i%10), "performance"},
			Blockers: []string{},
		}
		
		err := suite.storage.CreateTask(task)
		require.NoError(t, err)
		tasks[i] = task
	}
	
	insertDuration := time.Since(start)
	t.Logf("Inserted %d tasks in %v (%.2f tasks/sec)", numTasks, insertDuration, float64(numTasks)/insertDuration.Seconds())
	
	// Test listing performance
	start = time.Now()
	
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=100", nil)
	w := httptest.NewRecorder()
	
	suite.taskHandler.HandleTasks(w, req)
	
	listDuration := time.Since(start)
	t.Logf("Listed 100 tasks from %d total in %v", numTasks, listDuration)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	returnedTasks := response["tasks"].([]interface{})
	assert.Len(t, returnedTasks, 100)
	
	// Test search performance
	start = time.Now()
	
	results, err := suite.storage.SearchTasks("performance", false, 50)
	require.NoError(t, err)
	
	searchDuration := time.Since(start)
	t.Logf("Searched %d tasks in %v, found %d results", numTasks, searchDuration, len(results))
	
	assert.True(t, len(results) > 0, "Should find performance-related tasks")
	assert.True(t, searchDuration < time.Second, "Search should complete within reasonable time")
}

// setupIntegrationTestForBench creates a test environment for benchmarks
func setupIntegrationTestForBench(b *testing.B) (*IntegrationTestSuite, func()) {
	b.Helper()
	
	// Create temporary database file
	tempDir := b.TempDir()
	tempDBPath := filepath.Join(tempDir, "benchmark_test.db")
	
	// Initialize storage with real SQLite database
	storage, err := sqlite.New(tempDBPath)
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(storage)
	
	suite := &IntegrationTestSuite{
		storage:     storage,
		taskHandler: taskHandler,
		tempDBPath:  tempDBPath,
	}
	
	cleanup := func() {
		storage.Close()
		os.Remove(tempDBPath)
	}
	
	return suite, cleanup
}

// Benchmark for integration performance monitoring
func BenchmarkIntegration_TaskCRUD(b *testing.B) {
	suite, cleanup := setupIntegrationTestForBench(b)
	defer cleanup()
	
	task := &models.Task{
		JiraID:   "BENCH-001",
		Title:    "Benchmark task",
		Priority: models.Normal,
		Status:   models.New,
		Tags:     []string{"benchmark"},
		Blockers: []string{},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create
		task.ID = "" // Reset ID for new creation
		err := suite.storage.CreateTask(task)
		if err != nil {
			b.Fatal(err)
		}
		
		// Read
		_, err = suite.storage.GetTask(task.ID)
		if err != nil {
			b.Fatal(err)
		}
		
		// Update
		task.Status = models.InProgress
		err = suite.storage.UpdateTask(task)
		if err != nil {
			b.Fatal(err)
		}
		
		// Delete
		err = suite.storage.DeleteTask(task.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}