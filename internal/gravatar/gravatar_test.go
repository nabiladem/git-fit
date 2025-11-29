package gravatar

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUploadAvatar(t *testing.T) {
	// Create a temporary image file
	tmpFile, err := os.CreateTemp("", "test-image-*.jpg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("fake image content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Mock Gravatar REST API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Missing or invalid Authorization header: %s", authHeader)
		}

		if authHeader != "Bearer test-access-token" {
			t.Errorf("Wrong access token in header: %s", authHeader)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("Wrong content type: %s", contentType)
		}

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			t.Errorf("Failed to parse multipart form: %v", err)
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("Failed to get file from form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileContent, _ := io.ReadAll(file)
		if string(fileContent) != string(content) {
			t.Errorf("File content mismatch")
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// Override API base URL for testing
	originalURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = originalURL }()

	client := NewClient("test-client-id", "test-client-secret", "http://localhost:8080/callback", false)
	// Manually set access token for testing (skip OAuth flow)
	client.AccessToken = "test-access-token"

	if err := client.UploadAvatar(tmpFile.Name()); err != nil {
		t.Errorf("UploadAvatar failed: %v", err)
	}
}
