package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var proxies []string
var timeout int
var progress bool

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download file over HTTP",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("download called")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().IntVarP(&workers, "workers", "w", 3, "amount of concurrent workers")
	downloadCmd.Flags().IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Maximum amount of threads")
	downloadCmd.Flags().StringVarP(&check, "check", "c", "", "Compare to a sha256sum")
	downloadCmd.Flags().StringVarP(&path, "output", "o", "", "Save to file")
	downloadCmd.Flags().IntVarP(&timeout, "timeout", "T", 30, "Set I/O and connection timeout")
	downloadCmd.Flags().BoolVarP(&progress, "progress", "p", false, "Progress bar")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
