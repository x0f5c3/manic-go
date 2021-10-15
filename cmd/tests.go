package cmd

import (
	"fmt"
	"github.com/i582/cfmt/cmd/cfmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/x0f5c3/manic-go/pkg/downloader"
)

var workers int
var check string
var threads int
var path string

// testsCmd represents the tests command
var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Testing downloading",
	Long: `Command used for testing the program.
	To use it, pass it the url and optionally workers and a sha256sum to compare with
	By default amount of workers is 2`,
	Args: cobra.MinimumNArgs(1),
	Run:  download,
}

func init() {
	rootCmd.AddCommand(testsCmd)
	testsCmd.Flags().IntVarP(&workers, "workers", "w", 3, "amount of concurrent workers")
	testsCmd.Flags().IntVarP(&threads, "threads", "t", 2, "Maximum amount of threads")
	testsCmd.Flags().StringVarP(&check, "check", "c", "", "Compare to a sha256sum")
	testsCmd.Flags().StringVarP(&path, "output", "o", "", "Save to file")
	testsCmd.Flags().BoolP("progress", "p", false, "Progress bar")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func download(cmd *cobra.Command, args []string) {
	url := args[0]
	client := http.Client{}
	file, err := downloader.New(url, check, &client, nil)
	if err != nil {
		cfmt.Printf("%v", err)
	}
	flag, err := cmd.Flags().GetBool("progress")
	if err != nil {
		panic(err)
	}
	if err := file.Download(workers, threads, flag); err != nil {
		cfmt.Printf("Error: %v\n", err)
	}
	if path != "" {
		if err := file.Save(path); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

}
