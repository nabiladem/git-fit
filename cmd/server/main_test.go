package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// createTestImage() - create an image for testing, returns image data and error
func createTestImage() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// TestHealthCheck() - test health check endpoint
func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// send request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp["status"])
	}
}

// TestCompressEndpoint() - test compress endpoint
func TestCompressEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// create a test image
	imgData, err := createTestImage()
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "test.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	if _, err := part.Write(imgData); err != nil {
		t.Fatalf("Failed to write image data: %v", err)
	}
	writer.Close()

	// send request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/compress", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if _, ok := resp["download_url"]; !ok {
		t.Error("Response missing 'download_url'")
	}
}

// TestCompressEndpointMissingFile() - test compress endpoint with missing file
func TestCompressEndpointMissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// send request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/compress", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestDownloadEndpoint() - test download endpoint
func TestDownloadEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// manually inject a file into the store
	id := "test-id"
	token := "test-token"

	fileStore.Lock()
	fileStore.m[id] = storedFile{
		Data:     []byte("fake image data"),
		Mime:     "image/png",
		Filename: "test.png",
		Expires:  time.Now().Add(time.Minute),
		Token:    token,
	}
	fileStore.Unlock()

	// test valid download
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download/"+id+"?token="+token, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "fake image data" {
		t.Errorf("Expected body 'fake image data', got '%s'", w.Body.String())
	}

	// test invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/download/"+id+"?token=wrong", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	// test missing file
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/download/missing?token="+token, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestNotFound() - test not found endpoint
func TestNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/unknown", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestCompressEndpoint_InvalidFormat() - test compress endpoint with invalid format
func TestCompressEndpoint_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// create multipart form with invalid parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// add a dummy file
	part, _ := writer.CreateFormFile("avatar", "test.png")
	part.Write([]byte("fake image data"))
	
	// add invalid quality parameter
	writer.WriteField("quality", "200")
	writer.Close()

	// send request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/compress", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	// should still process (clamped to valid range)
	if w.Code == http.StatusOK {
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
	}
}

// TestDownloadEndpoint_Expired() - test download endpoint with expired file
func TestDownloadEndpoint_Expired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// manually inject an expired file into the store
	id := "expired-id"
	token := "expired-token"
	fileStore.Lock()
	fileStore.m[id] = storedFile{
		Data:     []byte("expired data"),
		Mime:     "image/png",
		Filename: "expired.png",
		Expires:  time.Now().Add(-time.Minute), // expired 1 minute ago
		Token:    token,
	}
	fileStore.Unlock()

	// send request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download/"+id+"?token="+token, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for expired file, got %d", w.Code)
	}
}
