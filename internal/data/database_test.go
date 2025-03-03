package data

import (
	"database/sql"
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Use an in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create tables
	topicsQuery := `
	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		parent_id INTEGER,
		FOREIGN KEY (parent_id) REFERENCES topics(id) ON DELETE CASCADE
	);
	`
	_, err = db.Exec(topicsQuery)
	if err != nil {
		t.Fatalf("Failed to create topics table: %v", err)
	}

	linksQuery := `
	CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		thumbnail TEXT,
		topic_id INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
	);
	`
	_, err = db.Exec(linksQuery)
	if err != nil {
		t.Fatalf("Failed to create links table: %v", err)
	}

	// Test creating a topic
	topicName := "Test Topic"
	var nullParentID sql.NullInt64
	topicID, err := CreateTopic(db, topicName, nullParentID)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	if topicID <= 0 {
		t.Errorf("Expected positive topic ID, got %d", topicID)
	}

	// Test getting topics
	topics, err := GetTopics(db, nullParentID)
	if err != nil {
		t.Fatalf("Failed to get topics: %v", err)
	}

	if len(topics) != 1 {
		t.Errorf("Expected 1 topic, got %d", len(topics))
	}

	if topics[0].Name != topicName {
		t.Errorf("Expected topic name '%s', got '%s'", topicName, topics[0].Name)
	}

	// Test creating a link
	linkName := "Test Link"
	linkURL := "https://example.com"
	linkThumbnail := "https://example.com/favicon.ico"

	linkID, err := CreateLink(db, linkName, linkURL, linkThumbnail, topicID)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	if linkID <= 0 {
		t.Errorf("Expected positive link ID, got %d", linkID)
	}

	// Test getting links by topic
	links, err := GetLinksByTopic(db, topicID)
	if err != nil {
		t.Fatalf("Failed to get links: %v", err)
	}

	if len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	}

	if links[0].Name != linkName {
		t.Errorf("Expected link name '%s', got '%s'", linkName, links[0].Name)
	}

	if links[0].URL != linkURL {
		t.Errorf("Expected link URL '%s', got '%s'", linkURL, links[0].URL)
	}
}

func TestMain(m *testing.M) {
	// Setup

	// Run tests
	code := m.Run()

	// Teardown
	os.Remove("./michishirube.db") // Clean up any test database files

	os.Exit(code)
}
