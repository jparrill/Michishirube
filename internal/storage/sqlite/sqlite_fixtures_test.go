package sqlite

import (
	"testing"

	"michishirube/internal/models"
	"michishirube/internal/storage"
	"michishirube/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStorage_WithFixtures(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Load fixture data
	var tasks []models.Task
	testdata.LoadFixtures(t, "tasks.json", &tasks)
	
	var links []models.Link
	testdata.LoadFixtures(t, "links.json", &links)
	
	var comments []models.Comment
	testdata.LoadFixtures(t, "comments.json", &comments)
	
	// Insert all fixture data
	for _, task := range tasks {
		err := store.CreateTask(&task)
		require.NoError(t, err)
	}
	
	for _, link := range links {
		err := store.CreateLink(&link)
		require.NoError(t, err)
	}
	
	for _, comment := range comments {
		err := store.CreateComment(&comment)
		require.NoError(t, err)
	}
	
	// Test various scenarios with realistic data
	t.Run("list tasks with filters", func(t *testing.T) {
		// Get high priority tasks
		highPriorityTasks, err := store.ListTasks(storage.TaskFilters{
			Priority: []models.Priority{models.High, models.Critical},
		})
		require.NoError(t, err)
		
		// Should find high and critical priority tasks
		assert.Len(t, highPriorityTasks, 2) // high + critical from fixtures
		
		// Check that we got the right ones
		priorities := make(map[models.Priority]bool)
		for _, task := range highPriorityTasks {
			priorities[task.Priority] = true
		}
		assert.True(t, priorities[models.High])
		assert.True(t, priorities[models.Critical])
	})
	
	t.Run("search realistic content", func(t *testing.T) {
		// Search for memory-related tasks
		results, err := store.SearchTasks("memory", false, 10)
		require.NoError(t, err)
		
		// Should find the memory leak task
		assert.Len(t, results, 1)
		assert.Contains(t, results[0].Title, "memory leak")
		assert.Contains(t, results[0].Tags, "memory")
	})
	
	t.Run("get task with full relations", func(t *testing.T) {
		// Get the memory leak task (first in fixtures)
		taskID := tasks[0].ID
		
		// Get task details
		task, err := store.GetTask(taskID)
		require.NoError(t, err)
		assert.Equal(t, "Fix memory leak in pod controller", task.Title)
		
		// Get related links
		taskLinks, err := store.GetTaskLinks(taskID)
		require.NoError(t, err)
		assert.Len(t, taskLinks, 3) // PR + Slack + Jira from fixtures
		
		// Verify link types
		linkTypes := make(map[models.LinkType]bool)
		for _, link := range taskLinks {
			linkTypes[link.Type] = true
		}
		assert.True(t, linkTypes[models.PullRequest])
		assert.True(t, linkTypes[models.SlackThread])
		assert.True(t, linkTypes[models.JiraTicket])
		
		// Get related comments
		taskComments, err := store.GetTaskComments(taskID)
		require.NoError(t, err)
		assert.Len(t, taskComments, 3) // 3 comments from fixtures
		
		// Comments should be ordered by created_at
		for i := 1; i < len(taskComments); i++ {
			assert.True(t, taskComments[i-1].CreatedAt.Before(taskComments[i].CreatedAt) || 
				taskComments[i-1].CreatedAt.Equal(taskComments[i].CreatedAt))
		}
	})
}

func TestSQLiteStorage_SearchScenario(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Load search scenario
	var scenario struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		Tasks         []models.Task `json:"tasks"`
		SearchTestCases []struct {
			Query           string   `json:"query"`
			ExpectedMatches []string `json:"expected_matches"`
			Description     string   `json:"description"`
		} `json:"search_test_cases"`
	}
	
	testdata.LoadScenario(t, "search_dataset", &scenario)
	
	// Insert scenario tasks
	for _, task := range scenario.Tasks {
		err := store.CreateTask(&task)
		require.NoError(t, err)
	}
	
	// Run all search test cases
	for _, testCase := range scenario.SearchTestCases {
		t.Run(testCase.Description, func(t *testing.T) {
			results, err := store.SearchTasks(testCase.Query, false, 10)
			require.NoError(t, err)
			
			// Extract IDs from results
			var resultIDs []string
			for _, task := range results {
				resultIDs = append(resultIDs, task.ID)
			}
			
			// Compare with expected matches
			assert.ElementsMatch(t, testCase.ExpectedMatches, resultIDs,
				"Search for '%s' should return expected task IDs", testCase.Query)
		})
	}
}

func TestSQLiteStorage_WorkflowScenario(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Load workflow scenario
	var scenario struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Task         models.Task `json:"task"`
		WorkflowSteps []struct {
			Step             string   `json:"step"`
			ExpectedStatus   string   `json:"expected_status"`
			ExpectedBlockers []string `json:"expected_blockers"`
		} `json:"workflow_steps"`
		Links    []models.Link    `json:"links"`
		Comments []models.Comment `json:"comments"`
	}
	
	testdata.LoadScenario(t, "basic_workflow", &scenario)
	
	// Create initial task
	task := scenario.Task
	err := store.CreateTask(&task)
	require.NoError(t, err)
	
	// Test workflow transitions
	for _, step := range scenario.WorkflowSteps {
		t.Run(step.Step, func(t *testing.T) {
			// Update task based on workflow step
			task.Status = models.Status(step.ExpectedStatus)
			task.Blockers = step.ExpectedBlockers
			
			err := store.UpdateTask(&task)
			require.NoError(t, err)
			
			// Verify the update
			retrieved, err := store.GetTask(task.ID)
			require.NoError(t, err)
			
			assert.Equal(t, step.ExpectedStatus, string(retrieved.Status))
			assert.Equal(t, step.ExpectedBlockers, retrieved.Blockers)
		})
	}
	
	// Add links and comments
	for _, link := range scenario.Links {
		err := store.CreateLink(&link)
		require.NoError(t, err)
	}
	
	for _, comment := range scenario.Comments {
		err := store.CreateComment(&comment)
		require.NoError(t, err)
	}
	
	// Verify final state
	finalTask, err := store.GetTask(task.ID)
	require.NoError(t, err)
	
	finalLinks, err := store.GetTaskLinks(task.ID)
	require.NoError(t, err)
	assert.Len(t, finalLinks, len(scenario.Links))
	
	finalComments, err := store.GetTaskComments(task.ID)
	require.NoError(t, err)
	assert.Len(t, finalComments, len(scenario.Comments))
	
	// Update fixture with final results if UPDATE=true
	finalState := map[string]interface{}{
		"final_task":     finalTask,
		"final_links":    finalLinks,
		"final_comments": finalComments,
	}
	testdata.UpdateScenario(t, "basic_workflow_results", finalState)
}