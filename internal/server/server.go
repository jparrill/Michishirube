package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"michishirube/internal/config"
	"michishirube/internal/handlers"
	"michishirube/internal/logger"
	"michishirube/internal/storage"
	
	_ "michishirube/docs" // Import generated docs
)

type Server struct {
	config     *config.Config
	storage    storage.Storage
	httpServer *http.Server
	logger     *slog.Logger
}

func New(config *config.Config, storage storage.Storage, logger *slog.Logger) *Server {
	return &Server{
		config:  config,
		storage: storage,
		logger:  logger,
	}
}

func (s *Server) Start() error {
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(s.storage)
	webHandler := handlers.NewWebHandler(s.storage)

	// Setup routes with middleware
	mux := http.NewServeMux()

	// Web routes (frontend)
	mux.HandleFunc("/", webHandler.Dashboard)
	mux.HandleFunc("/task/", webHandler.TaskDetail)
	mux.HandleFunc("/new", webHandler.NewTask)
	mux.HandleFunc("/health", webHandler.HealthCheck)
	
	// API Documentation routes
	mux.HandleFunc("/docs", webHandler.SwaggerUI)
	mux.HandleFunc("/api-docs/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
	mux.HandleFunc("/swagger/doc.json", webHandler.SwaggerJSON)
	mux.HandleFunc("/openapi.yaml", webHandler.OpenAPISpec)

	// API routes (for AJAX calls from frontend)
	mux.HandleFunc("/api/tasks", taskHandler.HandleTasks)
	mux.HandleFunc("/api/tasks/", taskHandler.HandleTask)
	mux.HandleFunc("/api/links", taskHandler.HandleLinks)
	mux.HandleFunc("/api/links/", taskHandler.HandleLink)
	mux.HandleFunc("/api/comments", taskHandler.HandleComments)
	mux.HandleFunc("/api/comments/", taskHandler.HandleComment)
	mux.HandleFunc("/api/report", taskHandler.HandleReport)

	// Static files
	mux.Handle("/static/", webHandler.StaticFileHandler())

	// Apply middleware
	handler := s.loggingMiddleware(mux)

	// Configure HTTP server
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("Starting HTTP server", "port", s.config.Port, "addr", s.httpServer.Addr)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		return err
	}

	slog.Info("Server stopped")
	return nil
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Add configured logger to request context
		ctx := logger.WithLogger(r.Context(), s.logger)
		r = r.WithContext(ctx)
		
		// Create a custom ResponseWriter to capture status code
		ww := &responseWriter{ResponseWriter: w}
		
		// Call the next handler
		next.ServeHTTP(ww, r)
		
		// Log the request using the configured logger
		duration := time.Since(start)
		s.logger.InfoContext(ctx, "HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}
	return rw.ResponseWriter.Write(b)
}
