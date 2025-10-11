package compressor

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

// CompressImage() - compress image to the target size
/* inputPath (string) - path of the input image; outputPath (string) - path of the input image
// maxSize (int) - maximum size of the image in byes; outputFormat (string) - jpeg, png, or gif */
func CompressImage(inputPath string, outputPath string, maxSize int, outputFormat string, quality int, verbose bool) error {

    const minWidth = 100;

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

    bounds := img.Bounds()
    width := bounds.Dx()

    for width > minWidth {
        // resize to fit width while maintaining aspect ratio
        resizedImg := resize.Resize(uint(width), 0, img, resize.Lanczos3)

        // encode the resized image to the desired format
        var buf bytes.Buffer
        switch outputFormat {
        case "jpeg":
            err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: quality})
        case "png":
            err = png.Encode(&buf, resizedImg)
        case "gif":
            err = gif.Encode(&buf, resizedImg, nil)
        default:
            return fmt.Errorf("unsupported file format: %v. Supported formats are: jpeg, png, gif", outputFormat)
        }

        if err != nil {
            return fmt.Errorf("failed to encode image: %v", err)
        }

        if verbose {
            fmt.Printf("Trying width: %d -> Compressed size: %.2f bytes\n", width, float64(buf.Len()))
        }

        if buf.Len() <= maxSize {
            // create the output file
            outFile, err := os.Create(outputPath)

            if err != nil {
                return fmt.Errorf("failed to create output file: %v", err)
            }

            defer outFile.Close()

            if verbose {
                fmt.Println("Compressing and saving image...")
            }

            _, err = buf.WriteTo(outFile)
            if err != nil {
                return fmt.Errorf("failed to write compressed image to file: %v", err)
            }

            return nil
        }

        width = width * 90 / 100 // reduce width by 10%
    }

    return fmt.Errorf("failed to compress image under %d bytes", maxSize)
}
