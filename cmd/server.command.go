package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/server"
	"github.com/valentinesamuel/mockcraft/internal/server/handlers"
	"github.com/valentinesamuel/mockcraft/internal/server/jobs"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the REST API server",
	Long: `Start the REST API server for programmatic access to fake data generation.
Example:
mockcraft server --port 8080 --config server.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		// Create output directory
		outputDir := filepath.Join("output", "server")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		// Initialize job manager
		jobManager, err := jobs.NewManager(outputDir)
		if err != nil {
			log.Fatalf("Failed to create job manager: %v", err)
		}

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
