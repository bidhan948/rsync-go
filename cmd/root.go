package cmd

import (
	"fmt"
	"os"

	"github.com/bidhan948/rsync-go/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rsync-go",
	Short: "rsync-go - simple file sender with progress bar",
	Long: `rsync-go 🌀
A tiny cross-platform file transfer tool (Linux/Windows) 
that sends files to a server with a nice progress UI.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ui.PrintBanner()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Register subcommands here
	rootCmd.AddCommand(
		newServeCmd(),
		newSendCmd(),
	)
}
