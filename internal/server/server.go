package server

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/valentinesamuel/mockcraft/internal/server/handlers"
)

type Server struct {
	router  *gin.Engine
	port    int
	handler *handlers.Handler
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
		api.GET("/generate/:type", s.handler.HandleGenerate)

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
	return s.router.Run(addr)
}

// Dummy upload function - to be implemented later
func uploadToStorage(filePath string) (string, error) {
	// TODO: Implement actual file storage
	return "dummy-storage-path", nil
}
