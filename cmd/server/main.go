package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nabiladem/git-fit/internal/compressor"
)

// in-memory store for short-lived compressed files
type storedFile struct {
	Data     []byte
	Mime     string
	Filename string
	Expires  time.Time
	Token    string
}

// global file store
var (
	fileStore = struct {
		sync.Mutex
		m map[string]storedFile
	}{m: make(map[string]storedFile)}
)

// init() - initialize the janitor goroutine to clean up expired files
func init() {
	// janitor goroutine to remove expired items every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			fileStore.Lock()
			for k, v := range fileStore.m {
				if v.Expires.Before(now) {
					delete(fileStore.m, k)
				}
			}
			fileStore.Unlock()
		}
	}()
}

// main() - entry point
func main() {
	r := gin.Default()

	// Enable CORS with specific settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // React frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Allowed HTTP methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"}, // Allowed headers
		AllowCredentials: true, // Allow credentials like cookies (if needed)
	}))

	// POST /api/compress
	r.POST("/api/compress", func(c *gin.Context) {
		file, err := c.FormFile("avatar")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'avatar' file field"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
			return
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		tmp, err := os.CreateTemp("", "gitfit-*-upload"+ext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
			return
		}

		tmpPath := tmp.Name()
		defer func() {
			tmp.Close()
			os.Remove(tmpPath) // cleanup
		}()

		// copy uploaded content to temp file
		if _, err := io.Copy(tmp, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
			return
		}

		// optional form params: maxsize, format, quality
		maxSize := 1048576 // default 1MB
		if v := c.PostForm("maxsize"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				maxSize = n
			}
		}

		format := c.PostForm("format")
		if format == "" {
			format = "jpeg"
		}

		quality := 85
		if q := c.PostForm("quality"); q != "" {
			if n, err := strconv.Atoi(q); err == nil && n >= 1 && n <= 100 {
				quality = n
			}
		}

		// prepare output temp file with proper extension
		outExt := ".jpg"
		switch format {
		case "png":
			outExt = ".png"
		case "gif":
			outExt = ".gif"
		default:
			outExt = ".jpg"
		}

		outTmp, err := os.CreateTemp("", "gitfit-compressed-*"+outExt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create output temp file"})
			return
		}

		outPath := outTmp.Name()
		outTmp.Close()

		// run compression
		if err := compressor.CompressImage(tmpPath, outPath, maxSize, format, quality, false); err != nil {
			// remove output if present
			_ = os.Remove(outPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "compression failed", "detail": err.Error()})
			return
		}

		// stat compressed file to get size and mime type
		info, err := os.Stat(outPath)
		if err != nil {
			_ = os.Remove(outPath)
			_ = os.Remove(tmpPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stat compressed file"})
			return
		}

		mimeType := mime.TypeByExtension(filepath.Ext(outPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// read compressed file into memory
		data, err := os.ReadFile(outPath)
		if err != nil {
			_ = os.Remove(outPath)
			_ = os.Remove(tmpPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read compressed file"})
			return
		}

		// generate short-lived id and token
		idb := make([]byte, 16)
		_, _ = rand.Read(idb)
		id := hex.EncodeToString(idb)
		tokn := make([]byte, 16)
		_, _ = rand.Read(tokn)
		token := hex.EncodeToString(tokn)

		// store in memory for short period (5 minutes)
		fileStore.Lock()
		fileStore.m[id] = storedFile{
			Data:     data,
			Mime:     mimeType,
			Filename: filepath.Base(outPath),
			Expires:  time.Now().Add(5 * time.Minute),
			Token:    token,
		}
		fileStore.Unlock()

		// remove temp files immediately
		_ = os.Remove(outPath)
		_ = os.Remove(tmpPath)

		// build a download URL (preserve scheme and host from request)
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host
		downloadURL := fmt.Sprintf("%s://%s/api/download/%s?token=%s", scheme, host, id, token)

		// respond with JSON metadata including a signed download URL
		resp := gin.H{
			"filename":     filepath.Base(outPath),
			"size":         info.Size(),
			"mime":         mimeType,
			"message":      "compression successful",
			"download_url": downloadURL,
			"expires_in":   300, // seconds
		}

		c.JSON(http.StatusOK, resp)
	})

	// GET /api/download/:id?token=<token>
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

		// constant-time token compare
		if subtle.ConstantTimeCompare([]byte(token), []byte(f.Token)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid token"})
			return
		}

		// serve bytes
		c.Header("Content-Type", f.Mime)
		c.Header("Content-Disposition", "attachment; filename="+strconv.Quote(f.Filename))
		c.Data(http.StatusOK, f.Mime, f.Data)
	})

	// optionally serve your built React frontend
	r.Static("/assets", "./web/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	// Start the server
	addr := ":8080"
	fmt.Println("Starting server on", addr)
	if err := r.Run(addr); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}
