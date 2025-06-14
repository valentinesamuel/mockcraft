package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/valentinesamuel/mockcraft/internal/server/handlers"
)

type Server struct {
	router  *gin.Engine
	port    int
	handler *handlers.Handler
	server  *http.Server
}

type Config struct {
	Port            int
	UploadSizeLimit int64 // in bytes
	OutputSizeLimit int64 // in bytes
}

func NewServer(cfg Config, handler *handlers.Handler) *Server {
	gin.SetMode(gin.DebugMode)
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{
		router:  router,
		port:    cfg.Port,
		handler: handler,
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.Group("/api")
	{
		// Generate single fake data
		api.GET("/generate/:generator/:type", s.handler.HandleGenerate)

		// List all available generators
		api.GET("/generators", s.handler.HandleListGenerators)

		// Upload schema and start generation
		api.POST("/seed", s.handler.HandleSeed)

		// Check job status
		api.GET("/jobs/:id", s.handler.HandleJobStatus)

		// Download generated files
		api.GET("/download/:id", s.handler.HandleDownload)
	}

	// Metrics endpoint
	s.router.GET("/metrics", s.handler.HandleMetrics)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on %s", addr)

	// Create HTTP server
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		serverErrors <- s.server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Received signal: %v", sig)
		log.Println("Starting graceful shutdown...")

		// Give outstanding requests 5 seconds to complete.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := s.server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in 5s: %v", err)
			if err := s.server.Close(); err != nil {
				return fmt.Errorf("could not stop server: %w", err)
			}
		}

		log.Println("Server stopped")
	}

	return nil
}
