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
    outputFormat := flag.String("format", "", "Output image format (jpeg, png, or gif)") // given or inferred
    quality := flag.Int("quality", 85, "JPEG compression quality (1-100; 85 by default)")
    verbose := flag.Bool("v", false, "Verbose logging enabled")

    // usage message for flags
    flag.Usage = func() {
        fmt.Println("Usage: gitfit -input <input-image-file> -output <output-image-file> -maxsize <max size in bytes> " +
         "-format <jpeg|png|gif> -quality <0-100> -v [for verbose logging]")
        fmt.Println("Example: gitfit -input input.jpeg -output output.jpeg -maxsize 1000000 -format jpeg -quality 85 -v")
        fmt.Println("Flags:")
        flag.PrintDefaults()
    }

	flag.Parse() // parse the input

    // validate input
	if *inputPath == "" || *outputPath == "" {
        // assume user knows about both flags if one is given
        if !(*inputPath == "" && *outputPath == "") {
		    fmt.Println("Error: You must provide both -input and -output file paths.")
            os.Exit(1)
        }

		flag.Usage()
		os.Exit(1)
	}

    // check if input file exists
    if _, err := os.Stat(*inputPath); os.IsNotExist(err) {
        fmt.Printf("Error: Input file %s does not exist.\n", *inputPath)
        os.Exit(1)
    }

    // warn if default quality is being used
    if *verbose && *quality == 85 {
        fmt.Println("Using default quality of 85 for JPEG compression.")
    }

    // compress corresponding to the input format if not given in -output flag, jpeg by default
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

    // add the format extension if not present
    if filepath.Ext(*outputPath) == "" {
        *outputPath = *outputPath + "." + *outputFormat
    }

    // validate quality input for jpeg
    if *quality <= 0 || *quality > 100 {
        fmt.Println("Error: Value for -quality must be between 1 and 100 inclusive.")
        os.Exit(1)
    }

    // print settings
    if *verbose {
        fmt.Printf("Input file: %s\nOutput file: %s\nMaximum size: %d\nOutput format: %s\n",
         *inputPath, *outputPath, *maxSize, *outputFormat)
        
        if *outputFormat == "jpeg" {
            fmt.Println("Quality:", *quality)
        }
	}

	// call the function from the internal package to compress the image
	err := compressor.CompressImage(*inputPath, *outputPath, *maxSize, *outputFormat, *quality, *verbose)

	if err != nil {
		fmt.Println("Error compressing image:", err)
		os.Exit(1)
	}

	fmt.Println("Image compressed successfully!")
}
