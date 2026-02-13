package gravatar

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestGenerateRandomState() - test the generateRandomState method
func TestGenerateRandomState(t *testing.T) {
	state := generateRandomState()
	if len(state) == 0 {
		t.Error("generated state is empty")
	}

	if len(state) != 32 { // 16 bytes hex encoded
		t.Errorf("expected state length 32, got %d", len(state))
	}
}

// TestUploadAvatar_NotAuthenticated() - test the UploadAvatar method when not authenticated
func TestUploadAvatar_NotAuthenticated(t *testing.T) {
	client := NewClient("id", "secret", "uri", false)
	err := client.UploadAvatar("test.jpg")
	if err == nil {
		t.Error("expected error for unauthenticated upload")
	}

	if err.Error() != "not authenticated - call Authenticate() first" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestUploadAvatar_Success() - test the UploadAvatar method when authenticated
func TestUploadAvatar_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-avatar.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if err := createTestImage(tmpFile.Name(), 100, 100, "jpg"); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	tmpFile.Close()

	// mock Gravatar API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/me/avatars" {
			t.Errorf("expected path /me/avatars, got %s", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer token, got %s", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// override base URL for testing
	originalBaseURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = originalBaseURL }()

	client := NewClient("id", "secret", "uri", true)
	client.AccessToken = "test-token"

	if err := client.UploadAvatar(tmpFile.Name()); err != nil {
		t.Fatalf("UploadAvatar failed: %v", err)
	}
}

// TestUploadAvatar_CropError() - test the UploadAvatar method when crop fails
func TestUploadAvatar_CropError(t *testing.T) {
	// mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	originalBaseURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = originalBaseURL }()

	client := NewClient("id", "secret", "uri", false)
	client.AccessToken = "test-token"

	// try to upload a file that doesn't exist - should fail during crop
	err := client.UploadAvatar("nonexistent.jpg")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
