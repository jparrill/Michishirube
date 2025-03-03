package data

import (
	"database/sql"
	"fmt"
	"log"
)

// LoadFixtures loads sample data into the database for testing and development
func LoadFixtures(db *sql.DB) error {
	log.Println("Loading fixtures into the database...")

	// First clear existing data
	if err := clearExistingData(db); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Create root topics
	rootTopics := []struct {
		name string
	}{
		{"Development"},
		{"Design"},
		{"Learning"},
		{"Tools"},
	}

	rootTopicIDs := make(map[string]int64)
	for _, topic := range rootTopics {
		id, err := CreateTopic(db, topic.name, sql.NullInt64{Valid: false})
		if err != nil {
			return fmt.Errorf("failed to create root topic '%s': %w", topic.name, err)
		}
		rootTopicIDs[topic.name] = id
		log.Printf("Created root topic: %s (ID: %d)", topic.name, id)
	}

	// Create subtopics
	subtopics := []struct {
		name     string
		parentID string
	}{
		{"Go", "Development"},
		{"Rust", "Development"},
		{"Python", "Development"},
		{"UI/UX", "Design"},
		{"Graphics", "Design"},
		{"Courses", "Learning"},
		{"Books", "Learning"},
		{"CLI", "Tools"},
		{"Productivity", "Tools"},
	}

	subtopicIDs := make(map[string]int64)
	for _, topic := range subtopics {
		parentID := rootTopicIDs[topic.parentID]
		id, err := CreateTopic(db, topic.name, sql.NullInt64{Valid: true, Int64: parentID})
		if err != nil {
			return fmt.Errorf("failed to create subtopic '%s': %w", topic.name, err)
		}
		subtopicIDs[topic.name] = id
		log.Printf("Created subtopic: %s under %s (ID: %d)", topic.name, topic.parentID, id)
	}

	// Create links
	links := []struct {
		name    string
		url     string
		topicID string
		isRoot  bool
	}{
		{"Go Documentation", "https://golang.org/doc/", "Go", false},
		{"Rust Programming Language", "https://www.rust-lang.org/", "Rust", false},
		{"Python.org", "https://www.python.org/", "Python", false},
		{"Material Design", "https://material.io/design", "UI/UX", false},
		{"Dribbble", "https://dribbble.com/", "Graphics", false},
		{"Coursera", "https://www.coursera.org/", "Courses", false},
		{"O'Reilly", "https://www.oreilly.com/", "Books", false},
		{"GitHub", "https://github.com/", "Development", true},
		{"Figma", "https://www.figma.com/", "Design", true},
		{"Udemy", "https://www.udemy.com/", "Learning", true},
		{"VS Code", "https://code.visualstudio.com/", "Tools", true},
	}

	for _, link := range links {
		var topicID int64
		if link.isRoot {
			topicID = rootTopicIDs[link.topicID]
		} else {
			topicID = subtopicIDs[link.topicID]
		}

		id, err := CreateLink(db, link.name, link.url, "", topicID)
		if err != nil {
			return fmt.Errorf("failed to create link '%s': %w", link.name, err)
		}
		log.Printf("Created link: %s (ID: %d) in topic ID: %d", link.name, id, topicID)
	}

	log.Println("Fixtures loaded successfully!")
	return nil
}

// clearExistingData removes all existing data from the database
func clearExistingData(db *sql.DB) error {
	// Disable foreign key constraints temporarily
	_, err := db.Exec("PRAGMA foreign_keys = OFF;")
	if err != nil {
		return err
	}

	// Delete all data from tables
	_, err = db.Exec("DELETE FROM links;")
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM topics;")
	if err != nil {
		return err
	}

	// Reset auto-increment counters
	_, err = db.Exec("DELETE FROM sqlite_sequence WHERE name='links' OR name='topics';")
	if err != nil {
		return err
	}

	// Re-enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	return nil
}
