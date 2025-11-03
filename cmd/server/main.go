package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"

    "github.com/gin-gonic/gin"
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

        if _, err := io.Copy(tmp, src); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
            return
        }

        info, _ := os.Stat(tmpPath)

        c.JSON(http.StatusOK, gin.H{
            "filename": file.Filename,
            "size":     info.Size(),
            "message":  "upload successful",
        })
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
