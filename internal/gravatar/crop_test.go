package gravatar

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestImage() - creates a test image
/* path (string) - path to save test image
   width (int) - width of test image
   height (int) - height of test image
   format (string) - format of test image */
func createTestImage(path string, width, height int, format string) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if format == "png" {
		return png.Encode(f, img)
	}
	
	return jpeg.Encode(f, img, nil)
}

// TestCropToSquare() - tests the cropToSquare function
func TestCropToSquare(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "crop_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		width    int
		height   int
		format   string
		expected int // expected square size
	}{
		{"Landscape", 100, 50, "jpg", 50},
		{"Portrait", 50, 100, "jpg", 50},
		{"Square", 80, 80, "png", 80},
	}

	// test crop to square
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, tt.name+"."+tt.format)
			if err := createTestImage(filename, tt.width, tt.height, tt.format); err != nil {
				t.Fatalf("failed to create test image: %v", err)
			}

			croppedPath, err := cropToSquare(filename)
			if err != nil {
				t.Fatalf("cropToSquare failed: %v", err)
			}

			// verify output file exists
			f, err := os.Open(croppedPath)
			if err != nil {
				t.Fatalf("failed to open cropped image: %v", err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Fatalf("failed to decode cropped image: %v", err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.expected || bounds.Dy() != tt.expected {
				t.Errorf("expected size %dx%d, got %dx%d", tt.expected, tt.expected, bounds.Dx(), bounds.Dy())
			}
		})
	}
}
