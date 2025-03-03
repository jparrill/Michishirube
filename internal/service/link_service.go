package service

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/jparrill/michishirube/internal/data"
)

// LinkService handles the business logic for links and topics
type LinkService struct {
	db *sql.DB
}

// NewLinkService creates a new LinkService
func NewLinkService(db *sql.DB) *LinkService {
	return &LinkService{
		db: db,
	}
}

// CreateTopic creates a new topic
func (s *LinkService) CreateTopic(name string, parentID *int64) (int64, error) {
	var nullParentID sql.NullInt64
	if parentID != nil {
		nullParentID.Int64 = *parentID
		nullParentID.Valid = true
	}

	return data.CreateTopic(s.db, name, nullParentID)
}

// GetTopics retrieves all topics, optionally filtered by parent ID
func (s *LinkService) GetTopics(parentID *int64) ([]data.Topic, error) {
	var nullParentID sql.NullInt64
	if parentID != nil {
		nullParentID.Int64 = *parentID
		nullParentID.Valid = true
	}

	return data.GetTopics(s.db, nullParentID)
}

// CreateLink creates a new link
func (s *LinkService) CreateLink(name, urlStr string, topicID int64) (int64, error) {
	// Validate URL
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return 0, fmt.Errorf("invalid URL: %w", err)
	}

	// Generate thumbnail (placeholder for now)
	thumbnail := s.generateThumbnail(urlStr)

	return data.CreateLink(s.db, name, urlStr, thumbnail, topicID)
}

// GetLinksByTopic retrieves all links for a specific topic
func (s *LinkService) GetLinksByTopic(topicID int64) ([]data.Link, error) {
	return data.GetLinksByTopic(s.db, topicID)
}

// SearchLinks searches for links by name
func (s *LinkService) SearchLinks(searchTerm string) ([]data.Link, error) {
	return data.SearchLinks(s.db, searchTerm)
}

// MoveLink moves a link to a different topic
func (s *LinkService) MoveLink(linkID, newTopicID int64) error {
	return data.MoveLink(s.db, linkID, newTopicID)
}

// DeleteLink removes a link
func (s *LinkService) DeleteLink(linkID int64) error {
	return data.DeleteLink(s.db, linkID)
}

// DeleteTopic removes a topic and all its links
func (s *LinkService) DeleteTopic(topicID int64) error {
	return data.DeleteTopic(s.db, topicID)
}

// generateThumbnail creates a thumbnail for a URL
// This is a placeholder implementation - in a real app, you might:
// 1. Use a service like PageShot API
// 2. Render the page in a headless browser and take a screenshot
// 3. Extract the favicon from the website
func (s *LinkService) generateThumbnail(urlStr string) string {
	// For now, we'll just use a placeholder based on the domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL for thumbnail: %v", err)
		return "https://via.placeholder.com/32/3498db/ffffff?text=?"
	}

	// Extract domain (e.g., "github.com" from "https://github.com/user/repo")
	domain := parsedURL.Hostname()
	if domain == "" {
		log.Printf("Could not extract domain from URL: %s", urlStr)
		return "https://via.placeholder.com/32/3498db/ffffff?text=?"
	}

	// Use a generic icon based on the first letter of the domain
	// This avoids network requests that could freeze the UI
	firstLetter := "?"
	if len(domain) > 0 {
		firstLetter = strings.ToUpper(domain[0:1])
	}

	// Return a placeholder image without making network requests
	return fmt.Sprintf("https://via.placeholder.com/32/3498db/ffffff?text=%s", firstLetter)
}
