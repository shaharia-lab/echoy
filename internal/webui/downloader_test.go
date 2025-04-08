package webui

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockHTTPClient is a mock implementation of the http.Client
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// MockZipCreator creates a simple valid zip file for testing
func createTestZip(t *testing.T) []byte {
	t.Helper()

	// Create a buffer to write our zip to
	buf := new(bytes.Buffer)

	// Create a new zip writer
	zipWriter := zip.NewWriter(buf)

	// Add a simple file to the zip
	fileWriter, err := zipWriter.Create("index.html")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}

	// Write some content to the file
	_, err = fileWriter.Write([]byte("<html><body>Test</body></html>"))
	if err != nil {
		t.Fatalf("Failed to write to file in zip: %v", err)
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func TestDownloadFrontend(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "frontend-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test zip file
	testZipData := createTestZip(t)

	tests := []struct {
		name          string
		version       string
		mockResponses []MockResponse
		wantErr       bool
		errContains   string
	}{
		{
			name:    "successful download latest version",
			version: "latest",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/latest",
					StatusCode: http.StatusOK,
					Body:       `{"tag_name":"v1.0.0","assets":[{"name":"dist.zip","browser_download_url":"https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip","content_type":"application/zip","size":1024}]}`,
				},
				{
					URL:        "https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip",
					StatusCode: http.StatusOK,
					Body:       string(testZipData),
				},
			},
			wantErr: false,
		},
		{
			name:    "successful download specific version",
			version: "v1.0.0",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/tags/v1.0.0",
					StatusCode: http.StatusOK,
					Body:       `{"tag_name":"v1.0.0","assets":[{"name":"dist.zip","browser_download_url":"https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip","content_type":"application/zip","size":1024}]}`,
				},
				{
					URL:        "https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip",
					StatusCode: http.StatusOK,
					Body:       string(testZipData),
				},
			},
			wantErr: false,
		},
		{
			name:    "error getting release info - HTTP error",
			version: "latest",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/latest",
					StatusCode: http.StatusNotFound,
					Body:       `{"message": "Not Found"}`,
				},
			},
			wantErr:     true,
			errContains: "failed to get download URL",
		},
		{
			name:    "error getting release info - network error",
			version: "latest",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/latest",
					StatusCode: 0,
					Err:        errors.New("network error"),
				},
			},
			wantErr:     true,
			errContains: "failed to get download URL",
		},
		{
			name:    "error downloading asset - HTTP error",
			version: "latest",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/latest",
					StatusCode: http.StatusOK,
					Body:       `{"tag_name":"v1.0.0","assets":[{"name":"dist.zip","browser_download_url":"https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip","content_type":"application/zip","size":1024}]}`,
				},
				{
					URL:        "https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip",
					StatusCode: http.StatusNotFound,
					Body:       `{"message": "Not Found"}`,
				},
			},
			wantErr:     true,
			errContains: "failed to download frontend asset",
		},
		{
			name:    "error downloading asset - network error",
			version: "latest",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/latest",
					StatusCode: http.StatusOK,
					Body:       `{"tag_name":"v1.0.0","assets":[{"name":"dist.zip","browser_download_url":"https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip","content_type":"application/zip","size":1024}]}`,
				},
				{
					URL:        "https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/dist.zip",
					StatusCode: 0,
					Err:        errors.New("network error"),
				},
			},
			wantErr:     true,
			errContains: "failed to download frontend asset",
		},
		{
			name:    "asset not found in release",
			version: "latest",
			mockResponses: []MockResponse{
				{
					URL:        "https://api.github.com/repos/shaharia-lab/echoy-webui/releases/latest",
					StatusCode: http.StatusOK,
					Body:       `{"tag_name":"v1.0.0","assets":[{"name":"other-file.zip","browser_download_url":"https://github.com/shaharia-lab/echoy-webui/releases/download/v1.0.0/other-file.zip","content_type":"application/zip","size":1024}]}`,
				},
			},
			wantErr:     true,
			errContains: "dist.zip asset not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP client
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Find the matching mock response for the request URL
					for _, mock := range tt.mockResponses {
						if req.URL.String() == mock.URL {
							if mock.Err != nil {
								return nil, mock.Err
							}
							return &http.Response{
								StatusCode: mock.StatusCode,
								Body:       io.NopCloser(bytes.NewBufferString(mock.Body)),
							}, nil
						}
					}
					t.Fatalf("Unexpected request to URL: %s", req.URL.String())
					return nil, nil
				},
			}

			// Create a destination directory for this test
			testDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			// Create the downloader with our mock client
			downloader := NewFrontendGitHubReleaseDownloader(testDir, mockClient)

			// Call the method we're testing
			err := downloader.DownloadFrontend(tt.version)

			// Check the result
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFrontend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && (err == nil || !errors.Is(err, err) && !strings.Contains(err.Error(), tt.errContains)) {
				t.Errorf("DownloadFrontend() error = %v, should contain: %v", err, tt.errContains)
			}

			// For successful tests, we could add additional checks here
			// For example, verify that files were created in the destination directory
		})
	}
}

// MockResponse represents a mock HTTP response for testing
type MockResponse struct {
	URL        string
	StatusCode int
	Body       string
	Err        error
}
