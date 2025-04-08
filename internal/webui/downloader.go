// Package webui provides functionality to download and extract the frontend assets from a GitHub release.
package webui

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/logger"
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

// HTTPClient is an interface that wraps the Do method, allowing for custom HTTP clients.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FrontendDownloader is an interface for downloading the frontend assets.
type FrontendDownloader interface {
	DownloadFrontend(version string) error
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
	DestinationDirectory string
	httpClient           HTTPClient
	logger               *logger.Logger
}

// NewFrontendGitHubReleaseDownloader creates a new instance of FrontendGitHubReleaseDownloader.
func NewFrontendGitHubReleaseDownloader(destinationDirectory string, httpClient HTTPClient, logger *logger.Logger) *FrontendGitHubReleaseDownloader {
	return &FrontendGitHubReleaseDownloader{
		DestinationDirectory: destinationDirectory,
		httpClient:           httpClient,
		logger:               logger,
	}
}

// DownloadFrontend downloads the frontend assets from a GitHub release and extracts them to the specified directory.
func (d *FrontendGitHubReleaseDownloader) DownloadFrontend(version string) error {
	d.logger.WithField("version", version).Info("Downloading frontend assets...")
	downloadURL, err := d.getDownloadURL(version)
	if err != nil {
		d.logger.WithField("error", err).Error("Failed to get download URL")
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	d.logger.WithFields(map[string]interface{}{"version": version, "download_url": downloadURL}).Info("Downloading frontend asset...")
	zipPath, err := d.downloadAsset(downloadURL)
	if err != nil {
		d.logger.WithField("error", err).Error("Failed to download frontend asset")
		return fmt.Errorf("failed to download frontend asset: %w", err)
	}
	defer os.Remove(zipPath)

	d.logger.WithField("zip_path", zipPath).Info("Extracting frontend asset...")

	if err := d.extractZip(zipPath); err != nil {
		d.logger.WithField("error", err).Error("Failed to extract frontend asset")
		return fmt.Errorf("failed to extract frontend: %w", err)
	}

	d.logger.WithFields(map[string]interface{}{
		"zip_path":              zipPath,
		"destination_directory": d.DestinationDirectory,
		"version":               version,
		"download_url":          downloadURL,
	}).Info("Frontend assets downloaded and extracted successfully")

	return nil
}

func (d *FrontendGitHubReleaseDownloader) getReleaseURL(releasePath string, releaseIdentifier string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/%s",
		githubAPIBaseURL,
		webUIRepoOwner,
		webUIRepoName,
		releasePath,
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
		return "", fmt.Errorf("failed to get release %s: %w", releaseIdentifier, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get release %s, status code: %d", releaseIdentifier, resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	for _, asset := range rel.Assets {
		if asset.Name == assetFileName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("dist.zip asset not found in release %s", releaseIdentifier)
}

func (d *FrontendGitHubReleaseDownloader) getDownloadURL(version string) (string, error) {
	if version == "latest" {
		return d.getReleaseURL("releases/latest", "latest")
	} else {
		return d.getReleaseURL(fmt.Sprintf("releases/tags/%s", version), version)
	}
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
