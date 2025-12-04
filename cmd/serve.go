package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/bidhan948/rsync-go/internal/server"
	"github.com/bidhan948/rsync-go/internal/ui"
)

func newServeCmd() *cobra.Command {
	var listen string
	var dir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run in server mode to receive files",
		Long: `Run an HTTP server that accepts file uploads.

Example:
  rsync-go serve --listen :8080 --dir ./received
`,
		Run: func(cmd *cobra.Command, args []string) {
			if listen == "" {
				listen = ":8080"
			}
			if dir == "" {
				dir = "./received"
			}

			ui.Info("Starting server...")
			ui.Info("Listen: " + listen)
			ui.Info("Output dir: " + dir)

			cfg := server.Config{
				ListenAddr: listen,
				OutputDir:  dir,
			}

			if err := server.Run(cfg); err != nil {
				log.Fatalf("server error: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&listen, "listen", ":8080", "address to listen on (e.g. :8080 or 0.0.0.0:8080)")
	cmd.Flags().StringVar(&dir, "dir", "./received", "directory where received files will be stored")

	return cmd
}
