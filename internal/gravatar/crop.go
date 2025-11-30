package gravatar

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// cropToSquare - takes an image path and creates a square version by center-cropping
// returns the path to the cropped image
func cropToSquare(imagePath string) (string, error) {
	// Open the image
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %v", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If already square, return original path
	if width == height {
		return imagePath, nil
	}

	// Determine square size (use smaller dimension)
	size := width
	if height < width {
		size = height
	}

	// Calculate crop offsets to center the crop
	xOffset := (width - size) / 2
	yOffset := (height - size) / 2

	// Create cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cropped.Set(x, y, img.At(x+xOffset, y+yOffset))
		}
	}

	// Create output path
	ext := filepath.Ext(imagePath)
	base := strings.TrimSuffix(imagePath, ext)
	croppedPath := base + "_square" + ext

	// Save cropped image
	outFile, err := os.Create(croppedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Encode based on format
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(outFile, cropped, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(outFile, cropped)
	default:
		err = jpeg.Encode(outFile, cropped, &jpeg.Options{Quality: 95})
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode cropped image: %v", err)
	}

	return croppedPath, nil
}
