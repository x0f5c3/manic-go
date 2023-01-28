package cmd

import (
	"net/url"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/x0f5c3/manic-go/pkg/downloader"
)

var proxy string
var timeout int
var progress bool

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download file over HTTP",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runDownload,
}

func runDownload(_ *cobra.Command, args []string) error {
	_, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	// httpClient, err := func() (*http.Client, error) {
	// 	if proxy != "" {
	// 		u, err := url.Parse(proxy)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		proxyFunc := http.ProxyURL(u)
	// 		trans := &http.Transport{Proxy: proxyFunc}
	// 		t := time.Second * time.Duration(timeout)
	// 		cl := &http.Client{Transport: trans, Timeout: t}
	// 		return cl, nil
	// 	} else {
	// 		t := time.Second * time.Duration(timeout)
	// 		cl := http.DefaultClient
	// 		cl.Timeout = t
	// 		return cl, nil
	// 	}
	// }()
	// if err != nil {
	// 	return err
	// }
	dl, err := downloader.New(args[0], check, nil, nil)
	if err != nil {
		return err
	}
	f, err := dl.Download(workers, threads, progress)
	if err != nil {
		return err
	}
	outputPath := func() string {
		if path != "" {
			return path
		}
		return "."
	}()
	return f.Save(outputPath)
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().IntVarP(&workers, "workers", "w", 3, "amount of concurrent workers")
	downloadCmd.Flags().IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Maximum amount of threads")
	downloadCmd.Flags().StringVarP(&check, "check", "c", "", "Compare to a sha256sum")
	downloadCmd.Flags().StringVarP(&path, "output", "o", "", "Save to file")
	downloadCmd.Flags().IntVarP(&timeout, "timeout", "T", 30, "Set I/O and connection timeout")
	downloadCmd.Flags().BoolVarP(&progress, "progress", "p", false, "Progress bar")
	downloadCmd.Flags().StringVar(&proxy, "proxy", "", "Proxy servers to use")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
