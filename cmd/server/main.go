package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nabiladem/git-fit/internal/compressor"
)

// main() - entry point
func main() {
	r := gin.Default()

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

		// respond with JSON metadata about the compressed file
		resp := gin.H{
			"filename": filepath.Base(outPath),
			"size":     info.Size(),
			"mime":     mimeType,
			"message":  "compression successful",
		}

		// schedule background cleanup of the temp files after 60s
		go func(p1, p2 string) {
			time.Sleep(60 * time.Second)
			_ = os.Remove(p1)
			_ = os.Remove(p2)
		}(outPath, tmpPath)

		// send JSON response
		c.JSON(http.StatusOK, resp)
	})

	// optionally serve your built React frontend
	r.Static("/", "./web/dist")

	addr := ":8080"
	fmt.Println("Starting server on", addr)
	if err := r.Run(addr); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}
