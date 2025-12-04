package cmd

import (
	"log"

	"github.com/bidhan948/rsync-go/internal/client"
	"github.com/spf13/cobra"
)

var syncWorkers int
var syncDest string
var syncChunkSize int64

var syncCmd = &cobra.Command{
	Use:   "sync <local_dir> <server_url>",
	Short: "Sync a folder to server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		localDir := args[0]
		serverURL := args[1]
		if syncDest == "" {
			log.Fatal("dest is required")
		}
		cfg := client.SyncConfig{
			ServerURL: serverURL,
			Token:     authToken,
			ChunkSize: syncChunkSize,
			Workers:   syncWorkers,
			LocalDir:  localDir,
			RemoteDir: syncDest,
		}
		s := client.NewSyncer(cfg)
		err := s.Sync()
		if err != nil {
			log.Fatalf("sync error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().IntVar(&syncWorkers, "workers", 4, "number of workers")
	syncCmd.Flags().StringVar(&syncDest, "dest", "", "remote destination dir")
	syncCmd.Flags().Int64Var(&syncChunkSize, "chunk-size", 4*1024*1024, "chunk size in bytes")
}
