package cmd

import (
	"github.com/pterm/pterm"
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
	RunE: run,
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
func run(cmd *cobra.Command, args []string) error {
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
			pterm.Error.Println(err)
			return err
		}
	} else {
		var err error
		file, err = downloader.LatestRelease(repo)
		if err != nil {
			pterm.Error.Println(err)
			return err
		}
	}
	err = file.Download(3, 2, true)
	if err != nil {
		pterm.Error.Println(err)
		return err
	}
	return nil
}
