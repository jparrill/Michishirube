package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"michishirube/internal/models"
	"michishirube/internal/storage"

	"github.com/google/uuid"
)

type SQLiteStorage struct {
	db *sql.DB
}

func New(dbPath string) (*SQLiteStorage, error) {
	db, err := openDB(dbPath + "?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db}

	if err := storage.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return storage, nil
}

func (s *SQLiteStorage) RunMigrations() error {
	return runMigrations(s.db)
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Task operations
func (s *SQLiteStorage) CreateTask(task *models.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	if task.Status == "" {
		task.Status = models.DefaultStatus
	}
	if task.Priority == "" {
		task.Priority = models.DefaultPriority
	}

	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	blockersJSON, err := json.Marshal(task.Blockers)
	if err != nil {
		return fmt.Errorf("failed to marshal blockers: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO tasks (id, jira_id, title, priority, status, tags, blockers, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.JiraID, task.Title, task.Priority, task.Status, string(tagsJSON), string(blockersJSON), task.CreatedAt, task.UpdatedAt)

	return err
}

func (s *SQLiteStorage) GetTask(id string) (*models.Task, error) {
	var task models.Task
	var tagsJSON, blockersJSON string

	err := s.db.QueryRow(`
		SELECT id, jira_id, title, priority, status, tags, blockers, created_at, updated_at
		FROM tasks WHERE id = ?
	`, id).Scan(
		&task.ID, &task.JiraID, &task.Title, &task.Priority, &task.Status,
		&tagsJSON, &blockersJSON, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if err := json.Unmarshal([]byte(blockersJSON), &task.Blockers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blockers: %w", err)
	}

	return &task, nil
}

func (s *SQLiteStorage) UpdateTask(task *models.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	task.UpdatedAt = time.Now()

	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	blockersJSON, err := json.Marshal(task.Blockers)
	if err != nil {
		return fmt.Errorf("failed to marshal blockers: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE tasks 
		SET jira_id = ?, title = ?, priority = ?, status = ?, tags = ?, blockers = ?, updated_at = ?
		WHERE id = ?
	`, task.JiraID, task.Title, task.Priority, task.Status, string(tagsJSON), string(blockersJSON), task.UpdatedAt, task.ID)

	return err
}

func (s *SQLiteStorage) DeleteTask(id string) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) ListTasks(filters storage.TaskFilters) ([]*models.Task, error) {
	query := "SELECT id, jira_id, title, priority, status, tags, blockers, created_at, updated_at FROM tasks WHERE 1=1"
	args := []interface{}{}

	if !filters.IncludeArchived {
		query += " AND status != 'archived'"
	}

	if len(filters.Status) > 0 {
		query += " AND status IN ("
		for i, status := range filters.Status {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, status)
		}
		query += ")"
	}

	if len(filters.Priority) > 0 {
		query += " AND priority IN ("
		for i, priority := range filters.Priority {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, priority)
		}
		query += ")"
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		var tagsJSON, blockersJSON string

		err := rows.Scan(
			&task.ID, &task.JiraID, &task.Title, &task.Priority, &task.Status,
			&tagsJSON, &blockersJSON, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if err := json.Unmarshal([]byte(blockersJSON), &task.Blockers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal blockers: %w", err)
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (s *SQLiteStorage) SearchTasks(query string, includeArchived bool, limit int) ([]*models.Task, error) {
	sqlQuery := `
		SELECT id, jira_id, title, priority, status, tags, blockers, created_at, updated_at 
		FROM tasks 
		WHERE (title LIKE ? OR jira_id LIKE ? OR tags LIKE ?)
	`
	args := []interface{}{
		"%" + query + "%",
		"%" + query + "%", 
		"%" + query + "%",
	}

	if !includeArchived {
		sqlQuery += " AND status != 'archived'"
	}

	sqlQuery += " ORDER BY created_at DESC"

	if limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		var tagsJSON, blockersJSON string

		err := rows.Scan(
			&task.ID, &task.JiraID, &task.Title, &task.Priority, &task.Status,
			&tagsJSON, &blockersJSON, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if err := json.Unmarshal([]byte(blockersJSON), &task.Blockers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal blockers: %w", err)
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// Link operations (simplified for now)
func (s *SQLiteStorage) CreateLink(link *models.Link) error {
	if err := link.Validate(); err != nil {
		return err
	}

	if link.ID == "" {
		link.ID = uuid.New().String()
	}

	_, err := s.db.Exec(`
		INSERT INTO links (id, task_id, type, url, title, status, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, link.ID, link.TaskID, link.Type, link.URL, link.Title, link.Status, link.Metadata)

	return err
}

func (s *SQLiteStorage) GetLink(id string) (*models.Link, error) {
	var link models.Link
	err := s.db.QueryRow(`
		SELECT id, task_id, type, url, title, status, metadata
		FROM links WHERE id = ?
	`, id).Scan(&link.ID, &link.TaskID, &link.Type, &link.URL, &link.Title, &link.Status, &link.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("link not found")
		}
		return nil, err
	}

	return &link, nil
}

func (s *SQLiteStorage) UpdateLink(link *models.Link) error {
	if err := link.Validate(); err != nil {
		return err
	}

	_, err := s.db.Exec(`
		UPDATE links 
		SET task_id = ?, type = ?, url = ?, title = ?, status = ?, metadata = ?
		WHERE id = ?
	`, link.TaskID, link.Type, link.URL, link.Title, link.Status, link.Metadata, link.ID)

	return err
}

func (s *SQLiteStorage) DeleteLink(id string) error {
	_, err := s.db.Exec("DELETE FROM links WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) GetTaskLinks(taskID string) ([]*models.Link, error) {
	rows, err := s.db.Query(`
		SELECT id, task_id, type, url, title, status, metadata
		FROM links WHERE task_id = ?
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(&link.ID, &link.TaskID, &link.Type, &link.URL, &link.Title, &link.Status, &link.Metadata)
		if err != nil {
			return nil, err
		}
		links = append(links, &link)
	}

	return links, nil
}

// Comment operations (simplified for now)
func (s *SQLiteStorage) CreateComment(comment *models.Comment) error {
	if err := comment.Validate(); err != nil {
		return err
	}

	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}

	comment.CreatedAt = time.Now()

	_, err := s.db.Exec(`
		INSERT INTO comments (id, task_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, comment.ID, comment.TaskID, comment.Content, comment.CreatedAt)

	return err
}

func (s *SQLiteStorage) GetComment(id string) (*models.Comment, error) {
	var comment models.Comment
	err := s.db.QueryRow(`
		SELECT id, task_id, content, created_at
		FROM comments WHERE id = ?
	`, id).Scan(&comment.ID, &comment.TaskID, &comment.Content, &comment.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, err
	}

	return &comment, nil
}

func (s *SQLiteStorage) DeleteComment(id string) error {
	_, err := s.db.Exec("DELETE FROM comments WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) GetTaskComments(taskID string) ([]*models.Comment, error) {
	rows, err := s.db.Query(`
		SELECT id, task_id, content, created_at
		FROM comments WHERE task_id = ? ORDER BY created_at ASC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.TaskID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}