// Package webserver provides a simple HTTP server
package webserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ShutdownTimeout defines how long to wait for server to gracefully shutdown
const ShutdownTimeout = 10 * time.Second

// WebServer represents a simple HTTP server
type WebServer struct {
	APIPort            string
	server             *http.Server
	router             *chi.Mux
	webStaticDirectory string
}

// NewWebServer creates a new WebServer instance with the specified API port
func NewWebServer(apiPort string, webStaticDirectory string) *WebServer {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return &WebServer{
		APIPort:            apiPort,
		router:             r,
		webStaticDirectory: webStaticDirectory,
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

	ws.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web", http.StatusFound)
	})

	// Serve "/ping" endpoint
	ws.router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Serve static files from the dist directory
	fileServer := http.FileServer(http.Dir(ws.webStaticDirectory))
	ws.router.Handle("/web", http.StripPrefix("/web", fileServer))
	ws.router.Handle("/web/*", http.StripPrefix("/web", fileServer))

}

// Start initializes and starts the HTTP server
func (ws *WebServer) Start() error {
	if ws.server != nil {
		return errors.New("server already running")
	}

	ws.setupRoutes()

	ws.server = &http.Server{
		Addr:    ":" + ws.APIPort,
		Handler: ws.router,
	}

	go func() {
		if err := ws.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server ListenAndServe error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the server and blocks until shutdown is complete or timeout occurs
func (ws *WebServer) Stop(ctx context.Context) error {
	if ws.server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, ShutdownTimeout)
	defer cancel()

	err := ws.server.Shutdown(shutdownCtx)

	ws.server = nil

	return err
}
