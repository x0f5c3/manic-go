package downloader

import (
	"context"
	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v33/github"
	"runtime"
	"strings"
)

func LatestRelease(repo string) (*File, error) {
	client := github.NewClient(nil)
	splitted := strings.Split(repo, "/")
	releases, _, err := client.Repositories.ListReleases(context.Background(), splitted[0], splitted[1], nil)
	if err != nil {
		return nil, err
	}
	latest := releases[0]
	os := runtime.GOOS
	arch := runtime.GOARCH
	var found string
	var size int
	for _, i := range latest.Assets {
		name := i.GetName()
		if strings.Contains(name, os) {
			if strings.Contains(name, arch) {
				found = i.GetBrowserDownloadURL()
				size = i.GetSize()
			}
		}
	}
	return New(found, "", nil, &size)

}
func AskForRelease(repo string) (*File, error) {
	client := github.NewClient(nil)
	splitted := strings.Split(repo, "/")
	releases, _, err := client.Repositories.ListReleases(context.Background(), splitted[0], splitted[1], nil)
	if err != nil {
		return nil, err
	}
	tags := make(map[string]*github.RepositoryRelease)
	var names []string
	for _, i := range releases {
		name := i.GetTagName()
		tags[name] = i
		names = append(names, name)
	}
	chosenName := ""
	prompt := &survey.Select{
		Message: "Choose a release:",
		Options: names,
	}
	err = survey.AskOne(prompt, &chosenName)
	if err != nil {
		return nil, err
	}
	chosenRelease := tags[chosenName]
	assets := make(map[string]*github.ReleaseAsset)
	var assetNames []string
	for _, i := range chosenRelease.Assets {
		name := i.GetName()
		assets[name] = i
		assetNames = append(assetNames, name)
	}
	chosenAsset := ""
	assetPrompt := &survey.Select{
		Message: "Choose an asset:",
		Options: assetNames,
	}
	err = survey.AskOne(assetPrompt, &chosenAsset)
	if err != nil {
		return nil, err
	}
	chosenAssets := assets[chosenAsset]
	size := chosenAssets.GetSize()
	url := chosenAssets.GetBrowserDownloadURL()
	return New(url, "", nil, &size)
}
