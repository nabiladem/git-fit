package main

import (
	"flag"
	"fmt"
	"os"
	"git-fit/internal/compressor"
)

func main() {

	// define command-line flags
	inputPath := flag.String("input", "", "Path to the input image file")
	outputPath := flag.String("output", "", "Path to save the compressed image")
	maxSize := flag.Int("maxsize", 1048576, "Maximum file size in bytes (default 1MB)") // 1MB

    // usage message to helping user with flags
    flag.Usage = func() {
        fmt.Println("Usage: gitfit -input <input-image-file> -output <output-image-file> -maxsize <max size in bytes>")
        fmt.Println("\nFlags:")
        flag.PrintDefaults()
    }

	flag.Parse() // parse the input

	// validate input
	if *inputPath == "" || *outputPath == "" {
		fmt.Println("Error: You must provide both -input and -output file paths.")
		flag.Usage()
		os.Exit(1)
	}

	// call the function from the internal package to compress the image
	err := compressor.CompressImage(*inputPath, *outputPath, *maxSize)

	if err != nil {
		fmt.Println("Error compressing image:", err)
		os.Exit(1)
	}

	fmt.Println("Image compressed successfully!")
}
