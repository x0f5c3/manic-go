/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/x0f5c3/manic-go/pkg/downloader"
)

var workers int
var check string

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
	testsCmd.Flags().StringVarP(&check, "check", "c", "", "Compare to a sha256sum")
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
	workers := workers
	client := http.Client{}
	file := downloader.New(url, check, &client)
	if err := file.Download(workers); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if check != "" {
		if err := file.CompareSha(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

}
