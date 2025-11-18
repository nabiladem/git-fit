package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nabiladem/git-fit/internal/compressor"
)

func setupRouter() *gin.Engine {
	os.Setenv("FRONTEND_URL", "http://localhost:5173")

	startTime := time.Now()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) { c.Next() })

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"uptime":    time.Since(startTime).Truncate(time.Second).String(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.POST("/api/compress", func(c *gin.Context) {
		file, err := c.FormFile("avatar")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'avatar' file field"})
			return
		}

		src, _ := file.Open()
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		tmp, _ := os.CreateTemp("", "gitfit-*-upload"+ext)
		tmpPath := tmp.Name()
		io.Copy(tmp, src)
		tmp.Close()

		outTmp, _ := os.CreateTemp("", "gitfit-compressed-*.jpg")
		outPath := outTmp.Name()
		outTmp.Close()

		_ = compressor.CompressImage(tmpPath, outPath, 1048576, "jpeg", 85, false)

		data, _ := os.ReadFile(outPath)
		info, _ := os.Stat(outPath)

		id := "testfile"
		token := "tokentest"

		fileStore.Lock()
		fileStore.m[id] = storedFile{
			Data:     data,
			Mime:     "image/jpeg",
			Filename: "compressed.jpg",
			Expires:  time.Now().Add(5 * time.Minute),
			Token:    token,
		}
		fileStore.Unlock()

		os.Remove(tmpPath)
		os.Remove(outPath)

		c.JSON(200, gin.H{
			"filename":     "compressed.jpg",
			"size":         info.Size(),
			"mime":         "image/jpeg",
			"download_url": "http://test/api/download/testfile?token=tokentest",
			"expires_in":   300,
		})
	})

	r.GET("/api/download/:id", func(c *gin.Context) {
		id := c.Param("id")
		token := c.Query("token")

		fileStore.Lock()
		f, ok := fileStore.m[id]
		fileStore.Unlock()

		if !ok || time.Now().After(f.Expires) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found or expired"})
			return
		}

		if token != f.Token {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid token"})
			return
		}

		c.Data(200, f.Mime, f.Data)
	})

	os.MkdirAll("./web/dist/assets", 0755)
	os.WriteFile("./web/dist/assets/test.txt", []byte("asset file"), 0644)
	os.WriteFile("./web/dist/index.html", []byte("<html>OK</html>"), 0644)

	r.Static("/assets", "./web/dist/assets")

	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[:5] == "/api/" {
			c.JSON(404, gin.H{
				"error":   "not found",
				"message": "API endpoint does not exist",
			})
			return
		}
		c.File("./web/dist/index.html")
	})

	return r
}

func TestHealth(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCompressFlow(t *testing.T) {
	r := setupRouter()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("avatar", "test.jpg")
	part.Write([]byte("fakeimage"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/compress", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("compress endpoint returned %d: %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest("GET", "/api/download/testfile?token=tokentest", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Fatalf("download failed with %d", w2.Code)
	}

	if string(w2.Body.Bytes()) != "fakeimage" {
		t.Fatalf("download returned unexpected data: %q", w2.Body.String())
	}
}

func TestDownloadInvalidToken(t *testing.T) {
	r := setupRouter()

	fileStore.Lock()
	fileStore.m["abc"] = storedFile{
		Data:     []byte("xyz"),
		Mime:     "image/jpeg",
		Filename: "x.jpg",
		Token:    "real",
		Expires:  time.Now().Add(1 * time.Minute),
	}
	fileStore.Unlock()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/download/abc?token=wrong", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestAPINotFound(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/does-not-exist", nil)
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestStaticAssets(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/assets/test.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("assets returned %d", w.Code)
	}

	if w.Body.String() != "asset file" {
		t.Fatalf("unexpected asset data: %q", w.Body.String())
	}
}

func TestFrontendFallback(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some/random/path", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for frontend fallback, got %d", w.Code)
	}

	if w.Body.String() != "<html>OK</html>" {
		t.Fatalf("fallback returned unexpected content: %q", w.Body.String())
	}
}
