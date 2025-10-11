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
    const MinWidth = 100;

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

    width := img.Bounds().Dx()
    best := 0
    var buf *bytes.Buffer
    
    low, high := MinWidth, width
    for low <= high {
        mid := (low + high) / 2

        // resize to fit width while maintaining aspect ratio
        resizedImg := resize.Resize(uint(mid), 0, img, resize.Lanczos3)

        // encode the resized image to the desired format
        var binaryBuf bytes.Buffer
        switch outputFormat {
        case "jpeg":
            err = jpeg.Encode(&binaryBuf, resizedImg, &jpeg.Options{Quality: quality})
        case "png":
            err = png.Encode(&binaryBuf, resizedImg)
        case "gif":
            err = gif.Encode(&binaryBuf, resizedImg, nil)
        default:
            return fmt.Errorf("unsupported file format: %v. Supported formats are: jpeg, png, gif", outputFormat)
        }

        if err != nil {
            return fmt.Errorf("failed to encode image: %v", err)
        }

        size := binaryBuf.Len()

        if verbose {
            fmt.Printf("[binary] Trying width: %d -> Compressed size: %.2f KB\n", width, float64(size) / 1024.0)
        }
        
        if size > maxSize {
            high = mid - 1
        } else {
            best = mid
            buf = &binaryBuf
            low = mid + 1
        }
    }
    
    if best == 0 {
        return fmt.Errorf("cannot compress image to the desired size of %d bytes", maxSize)
    }

    step := best / 20
    for width = best; width >= MinWidth; width -= step {
        // resize to fit width while maintaining aspect ratio
        resizedImg := resize.Resize(uint(width), 0, img, resize.Lanczos3)

        // encode the resized image to the desired format
        var binaryBuf bytes.Buffer
        switch outputFormat {
        case "jpeg":
            err = jpeg.Encode(&binaryBuf, resizedImg, &jpeg.Options{Quality: quality})
        case "png":
            err = png.Encode(&binaryBuf, resizedImg)
        case "gif":
            err = gif.Encode(&binaryBuf, resizedImg, nil)
        default:
            return fmt.Errorf("unsupported file format: %v. Supported formats are: jpeg, png, gif", outputFormat)
        }

        if err != nil {
            return fmt.Errorf("failed to encode image: %v", err)
        }

        size := binaryBuf.Len()

        if verbose {
            fmt.Printf("[linear] Trying width: %d -> Compressed size: %.2f KB\n", width, float64(size) / 1024.0)
        }
        
        if size <= maxSize {
            buf = &binaryBuf
            break
        }
    }

    if buf == nil {
        return fmt.Errorf("cannot compress image to the desired size of %d bytes", maxSize)
    }

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
