package cmd

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/server"
	"github.com/valentinesamuel/mockcraft/internal/server/handlers"
	"github.com/valentinesamuel/mockcraft/internal/server/jobs"
	"github.com/valentinesamuel/mockcraft/internal/server/output"
	"github.com/valentinesamuel/mockcraft/internal/server/storage"
)

// Config represents the server configuration
type Config struct {
	Database struct {
		Host            string
		Port            int
		Username        string
		Password        string
		Database        string
		SSLMode         string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
	}
}

func startCleanupTask(jobManager *jobs.Manager) {
	ticker := time.NewTicker(7 * 24 * time.Hour) // Run weekly
	go func() {
		for range ticker.C {
			if err := jobManager.CleanupCompletedJobs(); err != nil {
				log.Printf("Failed to cleanup completed jobs: %v", err)
			}
		}
	}()
}

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the REST API server",
	Long: `Start the REST API server for programmatic access to fake data generation.
Example:
mockcraft server `,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		// Create output directory
		outputDir := filepath.Join("output", "server")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		// Initialize job manager
		redisOpt := asynq.RedisClientOpt{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}

		// Load environment variables from .env file
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}

		supabaseURL := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")

		if supabaseURL == "" || supabaseKey == "" {
			log.Fatal("SUPABASE_URL and SUPABASE_KEY must be set in .env file")
		}

		jobManager, err := jobs.NewManager(
			redisOpt,
			outputDir,
			supabaseURL,
			supabaseKey,
		)
		if err != nil {
			log.Fatalf("Failed to create job manager: %v", err)
		}

		// Initialize job processor
		schemaStorage, err := storage.NewSupabaseStorage(supabaseURL, supabaseKey, "schemas")
		if err != nil {
			log.Fatalf("Failed to create schema storage: %v", err)
		}

		outputStorage, err := storage.NewSupabaseStorage(supabaseURL, supabaseKey, "output")
		if err != nil {
			log.Fatalf("Failed to create output storage: %v", err)
		}

		formatter := output.New(outputDir)

		processor, err := jobs.NewProcessor(
			schemaStorage,
			outputStorage,
			formatter,
			outputDir,
			jobManager,
			&redis.Options{
				Addr:     redisOpt.Addr,
				Password: redisOpt.Password,
				DB:       redisOpt.DB,
			},
		)
		if err != nil {
			log.Fatalf("Failed to create job processor: %v", err)
		}

		// Start the processor
		go func() {
			if err := processor.Start(); err != nil {
				log.Printf("Processor error: %v", err)
			}
		}()

		// Start cleanup task
		startCleanupTask(jobManager)

		// Initialize handlers
		handler := handlers.NewHandler(jobManager)

		// Create and start server
		srv := server.NewServer(server.Config{
			Port:            port,
			UploadSizeLimit: 30 * 1024 * 1024, // 30MB
			OutputSizeLimit: 50 * 1024 * 1024, // 50MB
		}, handler)

		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(ServerCmd)
	ServerCmd.Flags().Int("port", 8080, "Port to run the server on")
	ServerCmd.Flags().String("config", "", "Path to server configuration file")
}
