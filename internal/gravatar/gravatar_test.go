package gravatar

import (
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestUploadAvatar() - tests the UploadAvatar() function
/* t (*testing.T) - test context */
func TestUploadAvatar(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-image-*.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// write a real square JPEG image (100x100) to temp file for testing
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	col := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, col)
		}
	}

	if err := jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("failed to write JPEG to temp file: %v", err)
	}
	tmpFile.Close()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read back temp file: %v", err)
	}

	// mock Gravatar REST API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// check authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("missing or invalid Authorization header: %s", authHeader)
		}

		if authHeader != "Bearer test-access-token" {
			t.Errorf("wrong access token in header: %s", authHeader)
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("wrong content type: %s", contentType)
		}

		// parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			t.Errorf("failed to parse multipart form: %v", err)
		}

		file, _, err := r.FormFile("image")
		if err != nil {
			t.Errorf("failed to get file from form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileContent, _ := io.ReadAll(file)
		if string(fileContent) != string(content) {
			t.Errorf("file content mismatch")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// override API base URL for testing
	originalURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = originalURL }()

	// manually set access token for testing (skip OAuth flow)
	client := NewClient("test-client-id", "test-client-secret", "http://localhost:8080/callback", false)
	client.AccessToken = "test-access-token"

	if err := client.UploadAvatar(tmpFile.Name()); err != nil {
		t.Errorf("uploadAvatar failed: %v", err)
	}
}
