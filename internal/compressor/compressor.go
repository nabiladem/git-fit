package compressor

import (
    "fmt"
    "image"
    "image/jpeg"
    "image/png"
    "image/gif"
    "os"
    "github.com/nfnt/resize"
)

// CompressImage() - compress image to the target size
/* inputPath (string) - path of the input image; outputPath (string) - path of the input image
// maxSize (int) - maximum size of the image in byes; outputFormat (string) - jpeg, png, or gif */
func CompressImage(inputPath string, outputPath string, maxSize int, outputFormat string, quality int, verbose bool) error {

    if verbose {
        fmt.Println("Starting compression...")
    }

    // open the input image file
    file, err := os.Open(inputPath)

    if err != nil {
        return fmt.Errorf("failed to open input file: %v", err)
    }

    defer file.Close()

    if verbose {
        fmt.Println("Decoding image...")
    }

    // decode the image
    img, _, err := image.Decode(file)

    if err != nil {
        return fmt.Errorf("failed to decode image: %v", err)
    }

    // resize the image to fit within the maxSize
    resizedImg := resize.Resize(800, 0, img, resize.Lanczos3) // resize to fit width 800px

    // create the output file
    outFile, err := os.Create(outputPath)

    if err != nil {
        return fmt.Errorf("failed to create output file: %v", err)
    }

    defer outFile.Close()

    if verbose {
        fmt.Println("Compressing and saving image...")
    }
    
    // compress the resized image to the given or inferred output format
    switch outputFormat {
    case "jpeg":
        err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: quality})
    case "png":
        err = png.Encode(outFile, resizedImg)
    case "gif":
        err = gif.Encode(outFile, resizedImg, nil)
    default:
        return fmt.Errorf("unsupported file type: %v", outputFormat)
    }
	
    if err != nil {
        return fmt.Errorf("failed to compress image: %v", err)
    }

    return nil
}
