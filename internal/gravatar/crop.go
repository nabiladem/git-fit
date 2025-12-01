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

// cropToSquare() - takes an image path and creates a square version by center-cropping, returns the path to the cropped image
// imagePath (string) - path to the image to crop
func cropToSquare(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// if already square, return original path
	if width == height {
		return imagePath, nil
	}

	// determine square size (use smaller dimension)
	size := width
	if height < width {
		size = height
	}

	// calculate crop offsets to center the crop
	xOffset := (width - size) / 2
	yOffset := (height - size) / 2

	// create cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cropped.Set(x, y, img.At(x+xOffset, y+yOffset))
		}
	}

	// create output path
	ext := filepath.Ext(imagePath)
	base := strings.TrimSuffix(imagePath, ext)
	croppedPath := base + "_square" + ext

	outFile, err := os.Create(croppedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// encode based on format, 100 is used because the image is already compressed and Gravatar might already have compressed it
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(outFile, cropped, &jpeg.Options{Quality: 100})
	case "png":
		err = png.Encode(outFile, cropped)
	default:
		err = jpeg.Encode(outFile, cropped, &jpeg.Options{Quality: 100})
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode cropped image: %v", err)
	}

	return croppedPath, nil
}
