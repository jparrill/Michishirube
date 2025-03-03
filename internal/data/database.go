package data

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database and creates the necessary tables
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./michishirube.db")
	if err != nil {
		return nil, err
	}

	// Create the topics table
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
		return nil, fmt.Errorf("failed to create topics table: %w", err)
	}

	// Create the links table with foreign key to topics
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
		return nil, fmt.Errorf("failed to create links table: %w", err)
	}

	// Enable foreign key support
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}

// Topic represents a category or folder for organizing links
type Topic struct {
	ID       int64
	Name     string
	ParentID sql.NullInt64
}

// Link represents a bookmark to a website
type Link struct {
	ID        int64
	Name      string
	URL       string
	Thumbnail string
	TopicID   int64
	CreatedAt string
}

// CreateTopic adds a new topic to the database
func CreateTopic(db *sql.DB, name string, parentID sql.NullInt64) (int64, error) {
	query := `INSERT INTO topics (name, parent_id) VALUES (?, ?)`
	result, err := db.Exec(query, name, parentID)
	if err != nil {
		return 0, fmt.Errorf("failed to create topic: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetTopics retrieves all topics, optionally filtered by parent ID
func GetTopics(db *sql.DB, parentID sql.NullInt64) ([]Topic, error) {
	var query string
	var args []interface{}

	if parentID.Valid {
		query = `SELECT id, name, parent_id FROM topics WHERE parent_id = ?`
		args = append(args, parentID.Int64)
		log.Printf("Executing query for topics with parent ID %d: %s", parentID.Int64, query)
	} else {
		// If no parent ID is specified, get all topics
		query = `SELECT id, name, parent_id FROM topics`
		log.Printf("Executing query for all topics: %s", query)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Name, &topic.ParentID); err != nil {
			return nil, fmt.Errorf("failed to scan topic row: %w", err)
		}
		topics = append(topics, topic)
		log.Printf("Found topic: ID=%d, Name=%s, ParentID=%v",
			topic.ID,
			topic.Name,
			topic.ParentID)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over topic rows: %w", err)
	}

	log.Printf("Retrieved %d topics", len(topics))
	return topics, nil
}

// CreateLink adds a new link to the database
func CreateLink(db *sql.DB, name, url, thumbnail string, topicID int64) (int64, error) {
	query := `INSERT INTO links (name, url, thumbnail, topic_id) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, name, url, thumbnail, topicID)
	if err != nil {
		return 0, fmt.Errorf("failed to create link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetLinksByTopic retrieves all links for a specific topic
func GetLinksByTopic(db *sql.DB, topicID int64) ([]Link, error) {
	query := `SELECT id, name, url, thumbnail, topic_id, created_at FROM links WHERE topic_id = ?`

	// Log the query and parameters for debugging
	log.Printf("Executing query: %s with topicID: %d", query, topicID)

	rows, err := db.Query(query, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.Name, &link.URL, &link.Thumbnail, &link.TopicID, &link.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan link row: %w", err)
		}
		links = append(links, link)
		log.Printf("Found link: ID=%d, Name=%s, URL=%s, TopicID=%d", link.ID, link.Name, link.URL, link.TopicID)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over link rows: %w", err)
	}

	log.Printf("Retrieved %d links for topic %d", len(links), topicID)
	return links, nil
}

// SearchLinks searches for links by name
func SearchLinks(db *sql.DB, searchTerm string) ([]Link, error) {
	query := `SELECT id, name, url, thumbnail, topic_id, created_at FROM links WHERE name LIKE ?`
	rows, err := db.Query(query, "%"+searchTerm+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search links: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.Name, &link.URL, &link.Thumbnail, &link.TopicID, &link.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan link row: %w", err)
		}
		links = append(links, link)
	}

	return links, nil
}

// MoveLink moves a link to a different topic
func MoveLink(db *sql.DB, linkID, newTopicID int64) error {
	query := `UPDATE links SET topic_id = ? WHERE id = ?`
	_, err := db.Exec(query, newTopicID, linkID)
	if err != nil {
		return fmt.Errorf("failed to move link: %w", err)
	}

	return nil
}

// DeleteLink removes a link from the database
func DeleteLink(db *sql.DB, linkID int64) error {
	query := `DELETE FROM links WHERE id = ?`
	_, err := db.Exec(query, linkID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}

// DeleteTopic removes a topic and all its links from the database
func DeleteTopic(db *sql.DB, topicID int64) error {
	query := `DELETE FROM topics WHERE id = ?`
	_, err := db.Exec(query, topicID)
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	return nil
}
