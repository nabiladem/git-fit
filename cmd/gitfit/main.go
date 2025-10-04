package main

import (
	"fmt"
	"git-fit/internal/compressor"
)

func main() {

    // call the function from the internal package to compress the image
    err := compressor.CompressImage("input.jpg", "output.jpg", 1*1024*1024) // 1MB
    if err != nil {
        fmt.Println("Error compressing image:", err)
        return
    }
    fmt.Println("Image compressed successfully!")
}
