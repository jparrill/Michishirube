package testdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// LoadFixtures loads JSON fixture data into the target interface
// If UPDATE=true env var is set, it will update the fixture file with target data after test
func LoadFixtures(t *testing.T, filename string, target interface{}) {
	t.Helper()
	
	// Find the root directory (where testdata exists)
	rootDir := findProjectRoot(t)
	fixturePath := filepath.Join(rootDir, "testdata", "fixtures", filename)
	
	// Check if file exists
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		// If UPDATE=true and file doesn't exist, create it
		if shouldUpdate() {
			t.Logf("Creating new fixture file: %s", fixturePath)
			UpdateFixture(t, filename, target)
			return
		}
		t.Fatalf("Fixture file does not exist: %s. Run with UPDATE=true to create it.", fixturePath)
	}
	
	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "Failed to read fixture file: %s", fixturePath)
	
	err = json.Unmarshal(data, target)
	require.NoError(t, err, "Failed to unmarshal fixture data from: %s", fixturePath)
	
	// Validate fixture data for reserved words that could cause test issues
	validateFixtureData(t, filename, target)
}

// UpdateFixture writes data to a JSON fixture file if UPDATE=true is set
func UpdateFixture(t *testing.T, filename string, data interface{}) {
	t.Helper()
	
	if !shouldUpdate() {
		return
	}
	
	rootDir := findProjectRoot(t)
	fixturePath := filepath.Join(rootDir, "testdata", "fixtures", filename)
	
	// Ensure directory exists
	dir := filepath.Dir(fixturePath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "Failed to create fixture directory: %s", dir)
	
	// Marshal with pretty printing
	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err, "Failed to marshal fixture data")
	
	err = os.WriteFile(fixturePath, jsonData, 0644)
	require.NoError(t, err, "Failed to write fixture file: %s", fixturePath)
	
	t.Logf("Updated fixture file: %s", fixturePath)
}

// LoadScenario loads a complete test scenario with expected results
func LoadScenario(t *testing.T, scenarioName string, target interface{}) {
	t.Helper()
	LoadFixtures(t, filepath.Join("scenarios", scenarioName+".json"), target)
}

// UpdateScenario updates a test scenario file
func UpdateScenario(t *testing.T, scenarioName string, data interface{}) {
	t.Helper()
	UpdateFixture(t, filepath.Join("scenarios", scenarioName+".json"), data)
}

// shouldUpdate checks if the UPDATE environment variable is set to true
func shouldUpdate() bool {
	update := os.Getenv("UPDATE")
	return update == "true" || update == "1"
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()
	
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// AssertOrUpdate compares expected vs actual data, or updates fixture if UPDATE=true
func AssertOrUpdate(t *testing.T, filename string, expected, actual interface{}) {
	t.Helper()
	
	if shouldUpdate() {
		UpdateFixture(t, filename, actual)
		t.Logf("Updated fixture %s with actual data", filename)
		return
	}
	
	// Load expected data and compare  
	LoadFixtures(t, filename, expected)
	
	// Convert both to JSON for comparison (handles struct differences)
	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err)
	
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	
	require.JSONEq(t, string(expectedJSON), string(actualJSON), 
		"Fixture data mismatch. Run with UPDATE=true to update fixture.")
}

// reservedWords are words that can cause false positives in search tests
// These should not appear in fixture data unless intentionally testing for them
var reservedWords = []string{
	"search", "find", "query", "match", "filter", "sort",
	"test", "fixture", "mock", "sample", "example",
}

// validateFixtureData checks for reserved words that could interfere with search tests
func validateFixtureData(t *testing.T, filename string, data interface{}) {
	t.Helper()
	
	// Skip validation for search-specific scenarios
	if strings.Contains(filename, "search_dataset") {
		return
	}
	
	// Convert to JSON for easier string searching
	jsonData, err := json.Marshal(data)
	if err != nil {
		return // Skip validation if we can't marshal
	}
	
	dataStr := strings.ToLower(string(jsonData))
	
	for _, word := range reservedWords {
		if strings.Contains(dataStr, strings.ToLower(word)) {
			t.Errorf(`
FIXTURE VALIDATION ERROR in %s:
Found reserved word "%s" in fixture data.

Reserved words can cause false positives in search tests.
Either:
1. Replace "%s" with a synonym in your fixture data
2. Add this file to the skip list in validateFixtureData() if intentional

Reserved words: %v
`, filename, word, word, reservedWords)
		}
	}
}