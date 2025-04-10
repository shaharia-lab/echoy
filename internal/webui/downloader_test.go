package webui

import (
	"archive/zip"
	"bytes"
	"errors"
	"github.com/shaharia-lab/echoy/internal/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/shaharia-lab/echoy/internal/webui/mocks"
)

func createTestZip(t *testing.T) []byte {
	t.Helper()

	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	fileWriter, err := zipWriter.Create("index.html")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}

	_, err = fileWriter.Write([]byte("<html><body>Test</body></html>"))
	if err != nil {
		t.Fatalf("Failed to write to file in zip: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func TestDownloadFrontend(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "frontend-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testZipData := createTestZip(t)

	tests := []struct {
		name          string
		version       string
		mockResponses []MockResponse
		wantErr       bool
		errContains   string
		expectedFiles []string
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
			wantErr:       false,
			expectedFiles: []string{"index.html"},
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
			wantErr:       false,
			expectedFiles: []string{"index.html"},
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
			mockClient := mocks.NewMockHTTPClient(t) // Use the constructor from the mocks package
			for _, mr := range tt.mockResponses {
				mockResp := mr

				var respBody io.ReadCloser = http.NoBody
				if mockResp.Body != "" {
					respBody = io.NopCloser(bytes.NewBufferString(mockResp.Body))
				}

				var expectedResponse *http.Response
				var expectedError error

				if mockResp.Err != nil {
					expectedError = mockResp.Err
				} else {
					expectedResponse = &http.Response{
						StatusCode: mockResp.StatusCode,
						Body:       respBody,
						Header:     make(http.Header),
					}
				}

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.URL.String() == mockResp.URL
				})).Return(expectedResponse, expectedError).Times(1)
			}

			testDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			noOpLogger := logger.NewNoOpLogger()

			downloader := NewFrontendGitHubReleaseDownloader(testDir, mockClient, noOpLogger)
			err = downloader.DownloadFrontend(tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFrontend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && (err == nil || !errors.Is(err, err) && !strings.Contains(err.Error(), tt.errContains)) {
				t.Errorf("DownloadFrontend() error = %v, should contain: %v", err, tt.errContains)
			}

			if tt.expectedFiles != nil {
				verifyFilesInDirectory(t, testDir, tt.expectedFiles)
			}
		})
	}
}

func verifyFilesInDirectory(t *testing.T, dir string, expectedFiles []string) {
	t.Helper()

	for _, expectedFile := range expectedFiles {
		filePath := filepath.Join(dir, expectedFile)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist in directory %s", expectedFile, dir)
		} else if err != nil {
			t.Errorf("Error checking file %s: %v", expectedFile, err)
		}
	}
}

type MockResponse struct {
	URL        string
	StatusCode int
	Body       string
	Err        error
}
