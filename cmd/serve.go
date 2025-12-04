package cmd

import (
	"fmt"
	"log"

	"github.com/bidhan948/rsync-go/internal/server"
	"github.com/spf13/cobra"
)

var servePort int
var serveDir string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start server to receive files",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := server.Config{
			Port:      servePort,
			BaseDir:   serveDir,
			AuthToken: authToken,
		}
		err := server.Run(cfg)
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "port to listen on")
	serveCmd.Flags().StringVar(&serveDir, "dir", "./received", "directory to write files")
	serveCmd.MarkFlagDirname("dir")
	fmt.Print("")
}
