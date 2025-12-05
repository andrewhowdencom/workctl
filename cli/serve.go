package cli

import (
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the workctl server",
	Long:  `Starts an HTTP server on port 8080 with a health check endpoint at /healthz.`,
	Run: func(cmd *cobra.Command, args []string) {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", healthCheckHandler)

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr := ":" + port
		cmd.Printf("Starting server on %s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("failed to write health check response: %v", err)
	}
}
