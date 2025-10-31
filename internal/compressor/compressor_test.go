package compressor

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// makeTestImage() - create a simple solid-color RGBA image of given dimensions and returns the created image
/* w (int) - width of the image; h (int) - height of the image */
func makeTestImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	col := color.RGBA{R: 200, G: 100, B: 50, A: 255}

	// fill with solid color
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, col)
		}
	}

	return img
}

// TestEncodeResizedToBuffer_SupportedFormats() - tests encoding for jpeg, png, and gif formats
/* t (*testing.T) - testing object */
func TestEncodeResizedToBuffer_SupportedFormats(t *testing.T) {
	img := makeTestImage(400, 300)

	// jpeg
	if buf, err := encodeResizedToBuffer(img, 100, "jpeg", 80); err != nil {
		t.Fatalf("jpeg encode failed: %v", err)
	} else if buf.Len() == 0 {
		t.Fatalf("jpeg buffer empty")
	}

	// png
	if buf, err := encodeResizedToBuffer(img, 100, "png", 80); err != nil {
		t.Fatalf("png encode failed: %v", err)
	} else if buf.Len() == 0 {
		t.Fatalf("png buffer empty")
	}

	// gif
	if buf, err := encodeResizedToBuffer(img, 100, "gif", 80); err != nil {
		t.Fatalf("gif encode failed: %v", err)
	} else if buf.Len() == 0 {
		t.Fatalf("gif buffer empty")
	}
}

// TestEncodeResizedToBuffer_UnsupportedFormat() - test error handling for unsupported format
/* t (*testing.T) - testing object */
func TestEncodeResizedToBuffer_UnsupportedFormat(t *testing.T) {
	img := makeTestImage(100, 100)

	if _, err := encodeResizedToBuffer(img, 50, "bmp", 80); err == nil {
		t.Fatalf("expected error for unsupported format, got nil")
	}
}

// TestFindBestWidthBinarySearchAndLinearRefine() - test the binary search and linear refine functions
/* t (*testing.T) - testing object */
func TestFindBestWidthBinarySearchAndLinearRefine(t *testing.T) {
	img := makeTestImage(800, 600)
	minWidth := 50
	maxWidth := 800

	// use a very large maxSize so the binary search will accept the largest width
	maxSize := 10 * 1024 * 1024 // 10 MB
	best, buf, err := findBestWidthBinarySearch(img, minWidth, maxWidth, maxSize, "jpeg", 80, false)
	if err != nil {
		t.Fatalf("binary search returned error: %v", err)
	}

	if best != maxWidth {
		t.Fatalf("expected best == maxWidth (%d), got %d", maxWidth, best)
	}

	if buf == nil || buf.Len() == 0 {
		t.Fatalf("expected non-empty buffer from binary search")
	}

	// linearRefine() starting at best should succeed (since size already <= maxSize)
	refined, err := linearRefine(img, best, minWidth, maxSize, "jpeg", 80, false)
	if err != nil {
		t.Fatalf("linearRefine returned error: %v", err)
	}

	if refined == nil || refined.Len() == 0 {
		t.Fatalf("linearRefine returned empty buffer")
	}
}

// TestLinearRefine_InvalidStartWidth() - test error handling for invalid start width
/* t (*testing.T) - testing object */
func TestLinearRefine_InvalidStartWidth(t *testing.T) {
	img := makeTestImage(200, 200)

	if _, err := linearRefine(img, 0, 10, 1000, "jpeg", 80, false); err == nil {
		t.Fatalf("expected error for invalid start width")
	}
}

// TestSaveBufferToFileAndLoadImage() - test saving a buffer to file and loading it back
/* t (*testing.T) - testing object */
func TestSaveBufferToFileAndLoadImage(t *testing.T) {
	td := t.TempDir()

	// create a small png buffer
	img := makeTestImage(120, 80)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	outPath := filepath.Join(td, "out.png")
	if err := saveBufferToFile(outPath, &buf); err != nil {
		t.Fatalf("saveBufferToFile error: %v", err)
	}

	// load it back using loadImage
	loaded, w, err := loadImage(outPath)
	if err != nil {
		t.Fatalf("loadImage failed: %v", err)
	}

	if loaded == nil || w == 0 {
		t.Fatalf("loaded image invalid")
	}
}

// TestLoadImage_MissingFile() - test error handling for missing input file
/* t (*testing.T) - testing object */
func TestLoadImage_MissingFile(t *testing.T) {
	if _, _, err := loadImage("does-not-exist-12345.png"); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

// TestCompressImage_EndToEnd() - end-to-end test of CompressImage()
/* t (*testing.T) - testing object */
func TestCompressImage_EndToEnd(t *testing.T) {
	td := t.TempDir()
	inPath := filepath.Join(td, "in.jpg")
	outPath := filepath.Join(td, "out.jpg")

	// create and write a test jpeg at high quality
	img := makeTestImage(640, 480)
	f, err := os.Create(inPath)
	if err != nil {
		t.Fatalf("create in file: %v", err)
	}

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 100}); err != nil {
		f.Close()
		t.Fatalf("jpeg encode: %v", err)
	}
	f.Close()

	// compress with a reasonably large maxSize so it should succeed without extreme shrinking
	if err := CompressImage(inPath, outPath, 5*1024*1024, "jpeg", 90, false); err != nil {
		t.Fatalf("CompressImage failed: %v", err)
	}

	// ensure output file exists and non-empty
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}

	if info.Size() == 0 {
		t.Fatalf("output file is empty")
	}
}
