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
	"github.com/joho/godotenv"
	"github.com/nabiladem/git-fit/internal/compressor"
)

// storedFile struct holds data for a compressed file
/* Data ([]byte) - file data; Mime (string) - MIME type of the file
   Filename (string) - original filename; Expires (time.Time) - expiration time
   Token (string) - access token for download authorization */
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

// init() - initialize janitor goroutine to clean up expired files
func init() {
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
	// load .env file if present
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	startTime := time.Now() // track uptime

	// new Gin router with no default middleware
	r := gin.New()

	// add logging and recovery middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// custom log format: [TIME] STATUS METHOD PATH (LATENCY)
		return fmt.Sprintf("[%s] %d | %13v | %s | %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))
	r.Use(gin.Recovery())

	// enable CORS (for local dev with Vite frontend)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// GET /api/health
	// simple health check for uptime and status
	r.GET("/api/health", func(c *gin.Context) {
		uptime := time.Since(startTime).Truncate(time.Second)
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"uptime":    uptime.String(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// POST /api/compress
	// sends a compressed image file in response
	// returns JSON with download URL
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

		// create temp file to store upload
		ext := filepath.Ext(file.Filename)
		tmp, err := os.CreateTemp("", "gitfit-*-upload"+ext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
			return
		}

		tmpPath := tmp.Name()
		defer func() {
			tmp.Close()
			os.Remove(tmpPath)
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

		// determine output extension
		outExt := ".jpg"
		switch format {
		case "png":
			outExt = ".png"
		case "gif":
			outExt = ".gif"
		}

		// create output temp file
		outTmp, err := os.CreateTemp("", "gitfit-compressed-*"+outExt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create output temp file"})
			return
		}

		outPath := outTmp.Name()
		outTmp.Close()

		// run compression
		if err := compressor.CompressImage(tmpPath, outPath, maxSize, format, quality, false); err != nil {
			_ = os.Remove(outPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "compression failed", "detail": err.Error()})
			return
		}

		// get output file info
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

		data, err := os.ReadFile(outPath)
		if err != nil {
			_ = os.Remove(outPath)
			_ = os.Remove(tmpPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read compressed file"})
			return
		}

		// generate short-lived id + token
		idb := make([]byte, 16)
		_, _ = rand.Read(idb)
		id := hex.EncodeToString(idb)
		tokn := make([]byte, 16)
		_, _ = rand.Read(tokn)
		token := hex.EncodeToString(tokn)

		// store in memory (expires in 5 min)
		fileStore.Lock()
		fileStore.m[id] = storedFile{
			Data:     data,
			Mime:     mimeType,
			Filename: filepath.Base(outPath),
			Expires:  time.Now().Add(5 * time.Minute),
			Token:    token,
		}
		fileStore.Unlock()

		// cleanup temp files
		_ = os.Remove(outPath)
		_ = os.Remove(tmpPath)

		// build a download URL
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		host := c.Request.Host
		downloadURL := fmt.Sprintf("%s://%s/api/download/%s?token=%s", scheme, host, id, token)

		resp := gin.H{
			"filename":     filepath.Base(outPath),
			"size":         info.Size(),
			"mime":         mimeType,
			"message":      "compression successful",
			"download_url": downloadURL,
			"expires_in":   300,
		}
		c.JSON(http.StatusOK, resp)
	})

	// GET /api/download/:id
	// serves the compressed file if token is valid
	r.GET("/api/download/:id", func(c *gin.Context) {
		id := c.Param("id")
		token := c.Query("token")

		// lookup file
		fileStore.Lock()
		f, ok := fileStore.m[id]
		fileStore.Unlock()

		if !ok || time.Now().After(f.Expires) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found or expired"})
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(f.Token)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid token"})
			return
		}

		c.Header("Content-Type", f.Mime)
		c.Header("Content-Disposition", "attachment; filename="+strconv.Quote(f.Filename))
		c.Data(http.StatusOK, f.Mime, f.Data)
	})

	// serve static frontend files
	r.Static("/assets", "./web/dist/assets")

	// handle unknown API routes with JSON 404
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[:5] == "/api/" {
			c.JSON(http.StatusNotFound, gin.H {
				"error":   "not found",
				"message": "API endpoint does not exist",
			})
			return
		}
		// otherwise, serve React index.html (for client-side routing)
		c.File("./web/dist/index.html")
	})

	addr := ":" + port
	fmt.Println("🚀 Server running on", addr)
	if err := r.Run(addr); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}
