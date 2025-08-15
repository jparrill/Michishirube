package server

import (
	"log"
	"net/http"

	"michishirube/internal/config"
	"michishirube/internal/handlers"
	"michishirube/internal/storage"
)

type Server struct {
	config  *config.Config
	storage storage.Storage
}

func New(config *config.Config, storage storage.Storage) *Server {
	return &Server{
		config:  config,
		storage: storage,
	}
}

func (s *Server) Start() error {
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(s.storage)

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/tasks", taskHandler.HandleTasks)
	mux.HandleFunc("/api/tasks/", taskHandler.HandleTask)

	// Web routes (basic for now)
	mux.HandleFunc("/", s.handleIndex)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	log.Printf("Starting server on port %s", s.config.Port)
	return http.ListenAndServe(":"+s.config.Port, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Michishirube</title>
</head>
<body>
    <h1>🗯️ Michishirube</h1>
    <p>Task management for developers</p>
    <p><a href="/api/tasks">View API</a></p>
</body>
</html>
	`))
}
