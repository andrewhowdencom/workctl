package cli

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"workctl/internal/handler/github"
	"workctl/internal/handler/health"
	"workctl/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the workctl server",
	Long:  `Starts an HTTP server on port 8080 with a health check endpoint at /healthz.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize structured logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", health.Handler)
		mux.Handle("/v1/github/webhooks", github.NewHandler(github.WithLogger(logger)))

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr := ":" + port
		logger.Info("Starting server", "addr", addr)

		srv, err := server.New(addr, mux)
		if err != nil {
			logger.Error("Failed to initialize server", "error", err)
			os.Exit(1)
		}

		if err := srv.Run(); err != nil {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
