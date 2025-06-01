package cmd

import "github.com/spf13/cobra"

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the REST API server",
	Long: `Start the REST API server for programmatic access to fake data generation.
Example:
mockcraft server --port 8080 --config server.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement server logic
	},
}

func init() {
	rootCmd.AddCommand(ServerCmd)
	ServerCmd.Flags().Int("port", 8080, "Port to run the server on")
	ServerCmd.Flags().String("config", "", "Path to server configuration file")
}
