package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"michishirube/internal/logger"
	"michishirube/internal/models"
	"michishirube/internal/storage"
)

type WebHandler struct {
	storage   storage.Storage
	templates *template.Template
}

type PageData struct {
	PageTitle string
	CustomJS  string

	// Dashboard data
	Tasks           []*TaskWithRelations
	SearchQuery     string
	StatusFilter    string
	IncludeArchived bool
	TaskCount       int

	// Pagination
	Offset       int
	CurrentCount int
	TotalCount   int
	HasMore      bool

	// Task detail data
	Task     *models.Task
	Links    []*models.Link
	Comments []*models.Comment

	// Form data
	JiraID   string
	Title    string
	Priority string
	Tags     []string
	Notes    string
}

type TaskWithRelations struct {
	*models.Task
	Links    []*models.Link
	Comments []*models.Comment
}

func NewWebHandler(storage storage.Storage) *WebHandler {
	tmpl := template.New("").Funcs(template.FuncMap{
		"join": strings.Join,
		"add":  func(a, b int) int { return a + b },
		"eq":   func(a, b interface{}) bool { return a == b },
		"ne":   func(a, b interface{}) bool { return a != b },
		"len": func(slice interface{}) int {
			switch s := slice.(type) {
			case []*models.Link:
				return len(s)
			case []*models.Comment:
				return len(s)
			case []string:
				return len(s)
			default:
				return 0
			}
		},
		"string": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},
		"slice": func() []*models.Link {
			return make([]*models.Link, 0)
		},
		"append": func(slice []*models.Link, item *models.Link) []*models.Link {
			return append(slice, item)
		},
		"index": func(slice []*models.Link, index int) *models.Link {
			if index >= 0 && index < len(slice) {
				return slice[index]
			}
			return nil
		},
	})

	templates, err := tmpl.ParseGlob("web/templates/*.html")
	if err != nil {
		panic("Failed to parse templates: " + err.Error())
	}

	return &WebHandler{
		storage:   storage,
		templates: templates,
	}
}

// Dashboard - Main page
func (h *WebHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Dashboard endpoint called",
		"search", r.URL.Query().Get("search"),
		"status_filter", r.URL.Query().Get("status"))

	// Parse query parameters
	query := r.URL.Query()
	searchQuery := query.Get("search")
	statusFilter := query.Get("status")
	includeArchived := query.Get("include_archived") == "true" || query.Get("include_archived") == "1"

	limit := 20
	offset := 0
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filters
	filters := storage.TaskFilters{
		IncludeArchived: includeArchived,
		Limit:           limit + 1, // Get one extra to check if there are more
		Offset:          offset,
	}

	if statusFilter != "" {
		filters.Status = []models.Status{models.Status(statusFilter)}
	}

	var tasks []*models.Task
	var err error

	// Search or list tasks
	if searchQuery != "" {
		tasks, err = h.storage.SearchTasks(searchQuery, includeArchived, limit+1)
	} else {
		tasks, err = h.storage.ListTasks(filters)
	}

	if err != nil {
		http.Error(w, "Failed to load tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check for pagination
	hasMore := len(tasks) > limit
	if hasMore {
		tasks = tasks[:limit] // Remove the extra task
	}

	// Load relations for each task
	tasksWithRelations := make([]*TaskWithRelations, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}

		links, _ := h.storage.GetTaskLinks(task.ID)
		comments, _ := h.storage.GetTaskComments(task.ID)

		// Ensure we have empty slices instead of nil
		if links == nil {
			links = []*models.Link{}
		}
		if comments == nil {
			comments = []*models.Comment{}
		}

		tasksWithRelations = append(tasksWithRelations, &TaskWithRelations{
			Task:     task,
			Links:    links,
			Comments: comments,
		})
	}

	// Get total count (for display)
	allTasks, _ := h.storage.ListTasks(storage.TaskFilters{IncludeArchived: includeArchived})
	totalCount := len(allTasks)

	data := &PageData{
		PageTitle:       "Dashboard",
		Tasks:           tasksWithRelations,
		SearchQuery:     searchQuery,
		StatusFilter:    statusFilter,
		IncludeArchived: includeArchived,
		TaskCount:       totalCount,
		Offset:          offset + 1, // 1-based for display
		CurrentCount:    offset + len(tasks),
		TotalCount:      totalCount,
		HasMore:         hasMore,
	}

	h.renderTemplate(w, "dashboard.html", data)
}

// TaskDetail - Show individual task
func (h *WebHandler) TaskDetail(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	// Extract task ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/task/")
	taskID := strings.Split(path, "/")[0]

	log.Debug("TaskDetail endpoint called", "task_id", taskID)

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Get task
	task, err := h.storage.GetTask(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to load task: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get related data
	links, _ := h.storage.GetTaskLinks(taskID)
	comments, _ := h.storage.GetTaskComments(taskID)

	// Ensure we have empty slices instead of nil
	if links == nil {
		links = []*models.Link{}
	}
	if comments == nil {
		comments = []*models.Comment{}
	}

	data := &PageData{
		PageTitle: task.Title,
		CustomJS:  "task.js",
		Task:      task,
		Links:     links,
		Comments:  comments,
	}

	h.renderTemplate(w, "task.html", data)
}

// NewTask - Show new task form
func (h *WebHandler) NewTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showNewTaskForm(w, r)
	case http.MethodPost:
		h.createNewTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *WebHandler) showNewTaskForm(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("New task form requested")

	data := &PageData{
		PageTitle: "New Task",
		CustomJS:  "new_task.js",
		Priority:  "normal", // Default priority
	}

	h.renderTemplate(w, "new_task.html", data)
}

func (h *WebHandler) createNewTask(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Creating new task")

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Error("Failed to parse form data", "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Extract form values
	jiraID := strings.TrimSpace(r.FormValue("jira_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	priority := r.FormValue("priority")
	tagsStr := r.FormValue("tags")
	notes := strings.TrimSpace(r.FormValue("notes"))

	// Validate required fields
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Set default Jira ID if empty
	if jiraID == "" {
		jiraID = models.DefaultNoJira
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Create task
	task := &models.Task{
		JiraID:   jiraID,
		Title:    title,
		Priority: models.Priority(priority),
		Status:   models.New,
		Tags:     tags,
		Blockers: []string{}, // Empty initially
	}

	err = h.storage.CreateTask(task)
	if err != nil {
		log.Error("Failed to create task", "error", err, "title", task.Title)
		// If validation error, show form with error
		var validationErr *models.ValidationError
		if errors.As(err, &validationErr) {
			h.showNewTaskFormWithError(w, task)
			return
		}
		http.Error(w, "Failed to create task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("Task created successfully", "task_id", task.ID, "title", task.Title, "priority", task.Priority)

	// Create initial comment if notes provided
	if notes != "" {
		comment := &models.Comment{
			TaskID:  task.ID,
			Content: notes,
		}
		if err := h.storage.CreateComment(comment); err != nil {
			log.Warn("Failed to create initial comment", "error", err)
		}
	}

	// Process initial links if provided
	linkTypes := r.Form["link_types[]"]
	linkURLs := r.Form["link_urls[]"]
	linkTitles := r.Form["link_titles[]"]

	for i, linkType := range linkTypes {
		if i < len(linkURLs) && linkURLs[i] != "" {
			title := ""
			if i < len(linkTitles) {
				title = linkTitles[i]
			}
			if title == "" {
				title = linkURLs[i] // Use URL as title if not provided
			}

			link := &models.Link{
				TaskID: task.ID,
				Type:   models.LinkType(linkType),
				URL:    linkURLs[i],
				Title:  title,
				Status: "active",
			}
			if err := h.storage.CreateLink(link); err != nil {
				log.Warn("Failed to create initial link", "error", err, "url", linkURLs[i])
			}
		}
	}

	// Redirect to the new task
	http.Redirect(w, r, "/task/"+task.ID, http.StatusSeeOther)
}

func (h *WebHandler) showNewTaskFormWithError(w http.ResponseWriter, task *models.Task) {
	data := &PageData{
		PageTitle: "New Task",
		CustomJS:  "new_task.js",
		JiraID:    task.JiraID,
		Title:     task.Title,
		Priority:  string(task.Priority),
		Tags:      task.Tags,
	}

	// Set error status but show the form
	w.WriteHeader(http.StatusBadRequest)
	h.renderTemplate(w, "new_task.html", data)
}

// Helper method to render templates
func (h *WebHandler) renderTemplate(w http.ResponseWriter, templateName string, data *PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Parse the specific templates for this page
	tmpl := template.New("").Funcs(template.FuncMap{
		"join": strings.Join,
		"add":  func(a, b int) int { return a + b },
		"eq":   func(a, b interface{}) bool { return a == b },
		"ne":   func(a, b interface{}) bool { return a != b },
		"len": func(slice interface{}) int {
			switch s := slice.(type) {
			case []*models.Link:
				return len(s)
			case []*models.Comment:
				return len(s)
			case []string:
				return len(s)
			default:
				return 0
			}
		},
		"string": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},
		"slice": func() []*models.Link {
			return make([]*models.Link, 0)
		},
		"append": func(slice []*models.Link, item *models.Link) []*models.Link {
			return append(slice, item)
		},
		"index": func(slice []*models.Link, index int) *models.Link {
			if index >= 0 && index < len(slice) {
				return slice[index]
			}
			return nil
		},
	})

	// Parse base template and the specific page template
	pageTemplate, err := tmpl.ParseFiles("web/templates/base.html", "web/templates/"+templateName)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = pageTemplate.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// StaticFileHandler - Serve static files
func (h *WebHandler) StaticFileHandler() http.Handler {
	fileServer := http.FileServer(http.Dir("web/static/"))
	return http.StripPrefix("/static/", fileServer)
}

// HealthCheck - Simple health check endpoint
func (h *WebHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)); err != nil {
		log := logger.FromContext(r.Context())
		log.Error("Failed to write health check response", "error", err)
	}
}

// OpenAPISpec - Serve the OpenAPI specification
func (h *WebHandler) OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	// Read the generated OpenAPI spec file from Swaggo
	specPath := filepath.Join("docs", "swagger.yaml")
	content, err := os.ReadFile(specPath)
	if err != nil {
		log.Error("Failed to read OpenAPI spec", "error", err, "path", specPath)
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		log.Error("Failed to write OpenAPI spec", "error", err)
		return
	}
}

// SwaggerUI - Serve Swagger UI for API documentation
func (h *WebHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Serving Swagger UI")

	// Get the current request URL to build the spec URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	specURL := fmt.Sprintf("%s://%s/openapi.yaml", scheme, r.Host)

	swaggerHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Michishirube API Documentation</title>
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
	<style>
		html {
			box-sizing: border-box;
			overflow: -moz-scrollbars-vertical;
			overflow-y: scroll;
		}

		*, *:before, *:after {
			box-sizing: inherit;
		}

		body {
			margin:0;
			background: #fafafa;
		}
	</style>
</head>
<body>
	<div id="swagger-ui"></div>

	<script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			const ui = SwaggerUIBundle({
				url: '` + specURL + `',
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIStandalonePreset
				],
				plugins: [
					SwaggerUIBundle.plugins.DownloadUrl
				],
				layout: "StandaloneLayout",
				validatorUrl: null,
				tryItOutEnabled: true,
				displayRequestDuration: true,
				docExpansion: "list",
				filter: true,
				showExtensions: true,
				showCommonExtensions: true
			});
		};
	</script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(swaggerHTML)); err != nil {
		log := logger.FromContext(r.Context())
		log.Error("Failed to write Swagger UI", "error", err)
	}
}

// SwaggerJSON - Serve the OpenAPI specification in JSON format
func (h *WebHandler) SwaggerJSON(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	// Read the generated OpenAPI spec file in JSON format from Swaggo
	specPath := filepath.Join("docs", "swagger.json")
	content, err := os.ReadFile(specPath)
	if err != nil {
		log.Error("Failed to read OpenAPI JSON spec", "error", err, "path", specPath)
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		log := logger.FromContext(r.Context())
		log.Error("Failed to write JSON spec", "error", err)
		return
	}
}
