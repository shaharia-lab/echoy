// Package webui provides functionality to download and extract the frontend assets from a GitHub release.
package webui

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	webUIRepoOwner   = "shaharia-lab"
	webUIRepoName    = "echoy-webui"
	githubAPIBaseURL = "https://api.github.com"
	assetFileName    = "dist.zip"
	downloadTimeout  = 60 * time.Second
)

// FrontendDownloader is an interface for downloading the frontend assets.
type FrontendDownloader interface {
	DownloadFrontend() error
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	Size               int    `json:"size"`
}

// FrontendGitHubReleaseDownloader is a struct that implements the FrontendDownloader interface.
type FrontendGitHubReleaseDownloader struct {
	Version              string
	DestinationDirectory string
	httpClient           *http.Client
}

// NewFrontendGitHubReleaseDownloader creates a new instance of FrontendGitHubReleaseDownloader.
func NewFrontendGitHubReleaseDownloader(version string, destinationDirectory string, httpClient *http.Client) *FrontendGitHubReleaseDownloader {
	return &FrontendGitHubReleaseDownloader{
		Version:              version,
		DestinationDirectory: destinationDirectory,
		httpClient:           httpClient,
	}
}

func (d *FrontendGitHubReleaseDownloader) DownloadFrontend() error {
	downloadURL, err := d.getDownloadURL()
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	zipPath, err := d.downloadAsset(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download frontend asset: %w", err)
	}
	defer os.Remove(zipPath)

	if err := d.extractZip(zipPath); err != nil {
		return fmt.Errorf("failed to extract frontend: %w", err)
	}

	return nil
}

func (d *FrontendGitHubReleaseDownloader) getLatestReleaseURL() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest",
		githubAPIBaseURL,
		webUIRepoOwner,
		webUIRepoName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get latest release, status code: %d", resp.StatusCode)
	}

	var release release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == assetFileName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", errors.New("dist.zip asset not found in the latest release")
}

func (d *FrontendGitHubReleaseDownloader) getSpecificReleaseURL() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s",
		githubAPIBaseURL,
		webUIRepoOwner,
		webUIRepoName,
		d.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get release %s: %w", d.Version, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get release %s, status code: %d", d.Version, resp.StatusCode)
	}

	var release release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == assetFileName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("dist.zip asset not found in release %s", d.Version)
}

func (d *FrontendGitHubReleaseDownloader) getDownloadURL() (string, error) {
	version := "latest"
	if d.Version != "" {
		version = d.Version
	}

	// For GitHub public repositories, we can directly form the download URL
	// The format for assets is: https://github.com/{owner}/{repo}/releases/download/{tag}/{filename}
	// For latest, we need to make a redirect request or use the GitHub API properly

	if version == "latest" {
		// For the latest release, we'll try a different approach
		url := fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/%s",
			webUIRepoOwner,
			webUIRepoName,
			assetFileName,
		)
		return url, nil
	}

	// For specific versions
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		webUIRepoOwner,
		webUIRepoName,
		d.Version,
		assetFileName,
	)
	return url, nil
}

func (d *FrontendGitHubReleaseDownloader) downloadAsset(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download asset, status code: %d", resp.StatusCode)
	}

	// Create temporary file to store the zip
	tempFile, err := os.CreateTemp("", "echoy-webui-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFile.Close()

	// Copy the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to save downloaded asset: %w", err)
	}

	return tempFile.Name(), nil
}

func (d *FrontendGitHubReleaseDownloader) cleanDestinationDirectory() error {
	if _, err := os.Stat(d.DestinationDirectory); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(d.DestinationDirectory)
	if err != nil {
		return fmt.Errorf("failed to read destination directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(d.DestinationDirectory, entry.Name())
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove item %s: %w", path, err)
		}
	}

	return nil
}

func (d *FrontendGitHubReleaseDownloader) extractZip(zipPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	if err := d.cleanDestinationDirectory(); err != nil {
		return err
	}

	if err := os.MkdirAll(d.DestinationDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, file := range reader.File {
		path := filepath.Join(d.DestinationDirectory, file.Name)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(d.DestinationDirectory)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", path, err)
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}

		zipFile, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip file entry %s: %w", file.Name, err)
		}

		_, err = io.Copy(outFile, zipFile)
		zipFile.Close()
		outFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	return nil
}
