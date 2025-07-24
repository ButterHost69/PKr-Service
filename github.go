package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
)

func getLatestTag(owner, repo string) (string, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}

	return release.GetTagName(), nil
}

func downloadExeFromTag(owner, repo, tag, outputPath string) error {
	ctx := context.Background()
	client := github.NewClient(nil)

	// Get release by tag
	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to get release by tag: %w", err)
	}

	// Find the .exe asset
	var exeURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.GetName(), ".exe") {
			exeURL = asset.GetBrowserDownloadURL()
			break
		}
	}

	if exeURL == "" {
		return fmt.Errorf("no .exe asset found in release %s", tag)
	}

	// Download the asset
	resp, err := http.Get(exeURL)
	if err != nil {
		return fmt.Errorf("failed to download exe: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status while downloading: %s", resp.Status)
	}

	// Save to file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save exe to file: %w", err)
	}

	return nil
}
