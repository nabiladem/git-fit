package main

import (
	"flag"
	"fmt"
	"git-fit/internal/compressor"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	// define command-line flags
	inputPath := flag.String("input", "", "Path to the input image file")
	outputPath := flag.String("output", "", "Path to save the compressed image")
	maxSize := flag.Int("maxsize", 1048576, "Maximum file size in bytes (default 1MB)") // 1MB
    outputFormat := flag.String("format", "", "Output image format (jpeg, png, or gif)")

    // usage message for flags
    flag.Usage = func() {
        fmt.Println("Usage: gitfit -input <input-image-file> -output <output-image-file> -maxsize <max size in bytes> -format <jpeg|png|gif>")
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

    // compress corresponding to the input type if not given in -output flag, jpeg by default
    if *outputFormat == "" {
        extension := strings.ToLower(filepath.Ext(*inputPath))

        switch extension {
        case ".png":
            *outputFormat = "png"
        case ".gif":
            *outputFormat = "gif"
        default:
            *outputFormat = "jpeg"
        }
    }

    // add the extension if not present
    if filepath.Ext(*outputPath) == "" {
        *outputPath = *outputPath + "." + *outputFormat
    }

	// call the function from the internal package to compress the image
	err := compressor.CompressImage(*inputPath, *outputPath, *maxSize, *outputFormat)

	if err != nil {
		fmt.Println("Error compressing image:", err)
		os.Exit(1)
	}

	fmt.Println("Image compressed successfully!")
}
