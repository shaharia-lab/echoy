// Package webserver provides a simple HTTP server
package webserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// WebServer represents a simple HTTP server
type WebServer struct {
	APIPort string
	server  *http.Server
	router  *chi.Mux
}

// NewWebServer creates a new WebServer instance with the specified API port
func NewWebServer(apiPort string) *WebServer {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return &WebServer{
		APIPort: apiPort,
		router:  r,
	}
}

// Router returns the chi router to allow adding routes from outside
func (ws *WebServer) Router() *chi.Mux {
	return ws.router
}

// setupRoutes configures the default routes
func (ws *WebServer) setupRoutes() {
	ws.router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
}

// Start initializes and starts the HTTP server
func (ws *WebServer) Start() error {
	ws.setupRoutes()

	ws.server = &http.Server{
		Addr:    ":" + ws.APIPort,
		Handler: ws.router,
	}

	go func() {
		if err := ws.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			println("Error starting server:", err.Error())
		}
	}()

	return nil
}

// Stop gracefully shuts down the server with a timeout
func (ws *WebServer) Stop() error {
	if ws.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ws.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
