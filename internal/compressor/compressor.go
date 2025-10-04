package compressor

import (
    "fmt"
    "image"
    "image/jpeg"
    "os"
    "github.com/nfnt/resize"
)

// compress image to the target size
func CompressImage(inputPath string, outputPath string, maxSize int) error {

    // open the input image file
    file, err := os.Open(inputPath)

    if err != nil {
        return fmt.Errorf("failed to open input file: %v", err)
    }

    defer file.Close()

    // decode the image
    img, _, err := image.Decode(file)

    if err != nil {
        return fmt.Errorf("failed to decode image: %v", err)
    }

    // resize the image to fit within the maxSize
    resizedImg := resize.Resize(800, 0, img, resize.Lanczos3) // Resize to fit width 800px, keeping aspect ratio

    // create the output file
    outFile, err := os.Create(outputPath)

    if err != nil {
        return fmt.Errorf("failed to create output file: %v", err)
    }

    defer outFile.Close()

    // compress the resized image as JPEG
    err = jpeg.Encode(outFile, resizedImg, nil)
	
    if err != nil {
        return fmt.Errorf("failed to compress image: %v", err)
    }

    return nil
}
