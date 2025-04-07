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
	GitHubRepositoryOwner = "shaharia-lab"
	GitHubRepositoryName  = "echoy-webui"
	GitHubAPIBaseURL      = "https://api.github.com"
	AssetFileName         = "dist.zip"
	DownloadTimeout       = 60 * time.Second
)

type FrontendDownloader interface {
	DownloadFrontend() error
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	Size               int    `json:"size"`
}

type FrontendGitHubReleaseDownloader struct {
	Version              string
	DestinationDirectory string
}

func NewFrontendGitHubReleaseDownloader(version string, destinationDirectory string) *FrontendGitHubReleaseDownloader {
	return &FrontendGitHubReleaseDownloader{
		Version:              version,
		DestinationDirectory: destinationDirectory,
	}
}

func (d *FrontendGitHubReleaseDownloader) getLatestReleaseURL() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest",
		GitHubAPIBaseURL,
		GitHubRepositoryOwner,
		GitHubRepositoryName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), DownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get latest release, status code: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == AssetFileName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", errors.New("dist.zip asset not found in the latest release")
}

func (d *FrontendGitHubReleaseDownloader) getSpecificReleaseURL() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s",
		GitHubAPIBaseURL,
		GitHubRepositoryOwner,
		GitHubRepositoryName,
		d.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), DownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get release %s: %w", d.Version, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get release %s, status code: %d", d.Version, resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == AssetFileName {
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
			GitHubRepositoryOwner,
			GitHubRepositoryName,
			AssetFileName,
		)
		return url, nil
	}

	// For specific versions
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		GitHubRepositoryOwner,
		GitHubRepositoryName,
		d.Version,
		AssetFileName,
	)
	return url, nil
}

func (d *FrontendGitHubReleaseDownloader) downloadAsset(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects (GitHub will redirect for latest)
			return nil
		},
	}

	resp, err := client.Do(req)
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
