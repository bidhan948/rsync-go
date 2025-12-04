package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/bidhan948/rsync-go/internal/client"
	"github.com/bidhan948/rsync-go/internal/ui"
)

func newSendCmd() *cobra.Command {
	var filePath string
	var serverURL string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a file to a remote rsync-go server",
		Long: `Send a file to a server running 'rsync-go serve'.

Example:
  rsync-go send --file ./movie.mkv --to http://192.168.1.20:8080/upload
`,
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" || serverURL == "" {
				log.Fatal("both --file and --to are required")
			}

			ui.Info("Preparing to send file...")
			ui.Info("File: " + filePath)
			ui.Info("To:   " + serverURL)

			cfg := client.SendConfig{
				FilePath: filePath,
				Server:   serverURL,
			}

			if err := client.SendFile(cfg); err != nil {
				ui.Error("send failed: " + err.Error())
				log.Fatalf("send file: %v", err)
			}

			ui.Success("File sent successfully ✅")
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "path to the file to send")
	cmd.Flags().StringVar(&serverURL, "to", "", "server upload URL (e.g. http://192.168.1.20:8080/upload)")

	return cmd
}
