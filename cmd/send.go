package cmd

import (
	"log"

	"github.com/bidhan948/rsync-go/internal/client"
	"github.com/spf13/cobra"
)

var sendFilePath string
var sendTo string
var sendChunkSize int64

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a single file",
	Run: func(cmd *cobra.Command, args []string) {
		if sendFilePath == "" || sendTo == "" {
			log.Fatal("file and to are required")
		}
		cfg := client.UploaderConfig{
			ServerURL: sendTo,
			Token:     authToken,
			ChunkSize: sendChunkSize,
		}
		u := client.NewUploader(cfg)
		err := u.SendFile(sendFilePath, "")
		if err != nil {
			log.Fatalf("send error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringVar(&sendFilePath, "file", "", "file to send")
	sendCmd.Flags().StringVar(&sendTo, "to", "", "server url")
	sendCmd.Flags().Int64Var(&sendChunkSize, "chunk-size", 4*1024*1024, "chunk size in bytes")
}
