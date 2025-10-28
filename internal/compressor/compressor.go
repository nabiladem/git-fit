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
/* outputPath (string) - path of the output image
   maxSize (int) - maximum size of the image in bytes; outputFormat (string) - jpeg, png, or gif */
func CompressImage(inputPath string, outputPath string, maxSize int, outputFormat string, quality int, verbose bool) error {
	const MinWidth = 100

	if verbose {
		fmt.Println("Starting compression...")
	}

	// load and decode image
	img, width, err := loadImage(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load image: %v", err)
	}

	if verbose {
		fmt.Println("Searching for best width (binary search)...")
	}

	best, buf, err := findBestWidthBinarySearch(img, MinWidth, width, maxSize, outputFormat, quality, verbose)
	if err != nil {
		return err
	}

	if best == 0 || buf == nil {
		return fmt.Errorf("cannot compress image to the desired size of %d bytes", maxSize)
	}

	// linear refinement to try slightly smaller widths in steps
	if verbose {
		fmt.Println("Refining result (linear search)...")
	}

	refinedBuf, err := linearRefine(img, best, MinWidth, maxSize, outputFormat, quality, verbose)
	if err == nil && refinedBuf != nil {
		buf = refinedBuf
	}

	if buf == nil {
		return fmt.Errorf("cannot compress image to the desired size of %d bytes", maxSize)
	}

	if verbose {
		fmt.Println("Saving compressed image...")
	}

	if err := saveBufferToFile(outputPath, buf); err != nil {
		return fmt.Errorf("failed to write compressed image to file: %v", err)
	}

	return nil
}

// loadImage() - open and decode an image from disk and returns the image and its width
/* inputPath (string) - path of the input image */
func loadImage(inputPath string) (image.Image, int, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode image from %s: %v", inputPath, err)
	}

	width := img.Bounds().Dx()
	return img, width, nil
}

// encodeResizedToBuffer() - resize image to target width and encode it to a bytes.Buffer in the requested format
/* img (image.Image) - input image; width (int) - target width
   outputFormat (string) - jpeg, png, or gif; quality (int) - JPEG quality */
func encodeResizedToBuffer(img image.Image, width int, outputFormat string, quality int) (*bytes.Buffer, error) {
	resizedImg := resize.Resize(uint(width), 0, img, resize.Lanczos3)

	var binaryBuf bytes.Buffer
	var err error
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&binaryBuf, resizedImg, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&binaryBuf, resizedImg)
	case "gif":
		err = gif.Encode(&binaryBuf, resizedImg, nil)
	default:
		return nil, fmt.Errorf("unsupported file format: %v. Supported formats are: jpeg, png, gif", outputFormat)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %v", err)
	}

	return &binaryBuf, nil
}

// findBestWidthBinarySearch() - perform a binary search on width to find the largest width that yields <= maxSize
/* img (image.Image) - input image; minWidth (int) - minimum width; maxWidth (int) - maximum width
   maxSize (int) - maximum size in bytes; outputFormat (string) - jpeg, png, or gif; quality (int) - JPEG quality */
func findBestWidthBinarySearch(img image.Image, minWidth, maxWidth, maxSize int, outputFormat string, quality int, verbose bool) (int, *bytes.Buffer, error) {
	low, high := minWidth, maxWidth
	best := 0
	var bestBuf *bytes.Buffer

	for low <= high {
		mid := (low + high) / 2
		buf, err := encodeResizedToBuffer(img, mid, outputFormat, quality)
		if err != nil {
			return 0, nil, err
		}

		size := buf.Len()
		if verbose {
			fmt.Printf("[binary] Trying width: %d -> Compressed size: %.2f KB\n", mid, float64(size)/1024.0)
		}

		if size > maxSize {
			high = mid - 1
		} else {
			best = mid
			bestBuf = buf
			low = mid + 1
		}
	}

	return best, bestBuf, nil
}

// linearRefine() - perform a linear search downward from startWidth to minWidth in small steps to try to meet maxSize
/* img (image.Image) - input image; startWidth (int) - starting width; minWidth (int) - minimum width
   maxSize (int) - maximum size in bytes; outputFormat (string) - jpeg, png, or gif; quality (int) - JPEG quality */
func linearRefine(img image.Image, startWidth, minWidth, maxSize int, outputFormat string, quality int, verbose bool) (*bytes.Buffer, error) {
	if startWidth <= 0 {
		return nil, fmt.Errorf("invalid start width")
	}

	step := startWidth / 20
	if step < 1 {
		step = 1
	}

	for w := startWidth; w >= minWidth; w -= step {
		buf, err := encodeResizedToBuffer(img, w, outputFormat, quality)
		if err != nil {
			return nil, err
		}

		size := buf.Len()
		if verbose {
			fmt.Printf("[linear] Trying width: %d -> Compressed size: %.2f KB\n", w, float64(size)/1024.0)
		}

		if size <= maxSize {
			return buf, nil
		}
	}

	return nil, fmt.Errorf("no linear refinement found")
}

// saveBufferToFile() - write the content of buf to a file at outputPath
/* outputPath (string) - path of the output image; buf (*bytes.Buffer) - buffer containing image data */
func saveBufferToFile(outputPath string, buf *bytes.Buffer) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = buf.WriteTo(outFile)
	return err
}
