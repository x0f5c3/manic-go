/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"github.com/spf13/cobra"
	"github.com/x0f5c3/manic-go/pkg/downloader"
)

// ghCmd represents the gh command
var ghCmd = &cobra.Command{
	Use:   "gh",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: run,
}

func init() {
	rootCmd.AddCommand(ghCmd)
	ghCmd.Flags().BoolP("interactive", "i", false, "Choose the release to download via a menu")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ghCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ghCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func run(cmd *cobra.Command, args []string) {
	repo := args[0]
	interactive, err := cmd.Flags().GetBool("interactive")
	if err != nil {
		panic(err)
	}
	var file *downloader.File
	if interactive {
		var err error
		file, err = downloader.AskForRelease(repo)
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		file, err = downloader.LatestRelease(repo)
		if err != nil {
			panic(err)
		}
	}
	err = file.Download(3, 2, true)
	if err != nil {
		panic(err)
	}

}