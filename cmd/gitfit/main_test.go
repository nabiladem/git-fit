package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// TestParseFlags() - tests the parseFlags function
func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "All Flags",
			args: []string{
				"-input", "in.jpg",
				"-output", "out.jpg",
				"-maxsize", "500",
				"-format", "png",
				"-quality", "90",
				"-v",
				"-upload-gravatar",
			},
			expected: Config{
				InputPath:      "in.jpg",
				OutputPath:     "out.jpg",
				MaxSize:        500,
				OutputFormat:   "png",
				Quality:        90,
				Verbose:        true,
				UploadGravatar: true,
			},
		},
		{
			name: "Defaults",
			args: []string{},
			expected: Config{
				MaxSize: 1048576,
				Quality: 85,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := parseFlags(tt.args)
			if cfg.InputPath != tt.expected.InputPath {
				t.Errorf("expected InputPath %s, got %s", tt.expected.InputPath, cfg.InputPath)
			}

			if cfg.OutputPath != tt.expected.OutputPath {
				t.Errorf("expected OutputPath %s, got %s", tt.expected.OutputPath, cfg.OutputPath)
			}

			if cfg.MaxSize != tt.expected.MaxSize {
				t.Errorf("expected MaxSize %d, got %d", tt.expected.MaxSize, cfg.MaxSize)
			}

			if cfg.OutputFormat != tt.expected.OutputFormat {
				t.Errorf("expected OutputFormat %s, got %s", tt.expected.OutputFormat, cfg.OutputFormat)
			}

			if cfg.Quality != tt.expected.Quality {
				t.Errorf("expected Quality %d, got %d", tt.expected.Quality, cfg.Quality)
			}

			if cfg.Verbose != tt.expected.Verbose {
				t.Errorf("expected Verbose %v, got %v", tt.expected.Verbose, cfg.Verbose)
			}

			if cfg.UploadGravatar != tt.expected.UploadGravatar {
				t.Errorf("expected UploadGravatar %v, got %v", tt.expected.UploadGravatar, cfg.UploadGravatar)
			}
		})
	}
}

// TestValidateConfig() - tests the validateConfig function
func TestValidateConfig(t *testing.T) {
	// create a dummy file for existence check
	tmpFile, err := os.CreateTemp("", "test-image-*.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name      string
		cfg       Config
		wantUsage bool
		wantErr   bool
	}{
		{
			name:      "Empty Paths",
			cfg:       Config{},
			wantUsage: true,
			wantErr:   false,
		},
		{
			name: "Missing Input",
			cfg: Config{
				OutputPath: "out.jpg",
			},
			wantUsage: false,
			wantErr:   true,
		},
		{
			name: "Missing Output",
			cfg: Config{
				InputPath: "in.jpg",
			},
			wantUsage: false,
			wantErr:   true,
		},
		{
			name: "Input Not Exist",
			cfg: Config{
				InputPath:  "nonexistent.jpg",
				OutputPath: "out.jpg",
			},
			wantUsage: false,
			wantErr:   true,
		},
		{
			name: "Invalid MaxSize",
			cfg: Config{
				InputPath:  tmpFile.Name(),
				OutputPath: "out.jpg",
				MaxSize:    0,
			},
			wantUsage: false,
			wantErr:   true,
		},
		{
			name: "Invalid Quality",
			cfg: Config{
				InputPath:  tmpFile.Name(),
				OutputPath: "out.jpg",
				MaxSize:    100,
				Quality:    101,
			},
			wantUsage: false,
			wantErr:   true,
		},
		{
			name: "Valid Config (Auto Format)",
			cfg: Config{
				InputPath:  tmpFile.Name(),
				OutputPath: "out.jpg",
				MaxSize:    100,
				Quality:    80,
			},
			wantUsage: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			showUsage, err := validateConfig(&tt.cfg)
			if showUsage != tt.wantUsage {
				t.Errorf("expected showUsage %v, got %v", tt.wantUsage, showUsage)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr && !tt.wantUsage && tt.cfg.OutputFormat == "" {
				// Check if format was inferred
				ext := filepath.Ext(tt.cfg.InputPath)
				if ext == ".jpg" && tt.cfg.OutputFormat != "jpeg" {
					t.Errorf("expected inferred format jpeg, got %s", tt.cfg.OutputFormat)
				}
			}
		})
	}
}

// TestRunCompress() - tests the runCompress function
func TestRunCompress(t *testing.T) {
	// Create a dummy image
	tmpIn, err := os.CreateTemp("", "test-in-*.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpIn.Name())

	if err := createTestImage(tmpIn.Name(), 100, 100, "jpg"); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	tmpIn.Close()

	tmpOut, err := os.CreateTemp("", "test-out-*.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpOut.Name())
	tmpOut.Close()

	cfg := &Config{
		InputPath:    tmpIn.Name(),
		OutputPath:   tmpOut.Name(),
		MaxSize:      1024 * 1024,
		OutputFormat: "jpeg",
		Quality:      80,
		Verbose:      true,
	}

	if err := runCompress(cfg); err != nil {
		t.Fatalf("runCompress failed: %v", err)
	}

	// check if output exists and has content
	info, err := os.Stat(tmpOut.Name())
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("output file is empty")
	}

	// test with mode Gravatar upload
	cfg.UploadGravatar = true
	
	// ensure env vars are unset
	os.Unsetenv("GRAVATAR_CLIENT_ID")
	if err := runCompress(cfg); err == nil {
		t.Error("expected error for missing env vars")
	}
}

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

	return jpeg.Encode(f, img, nil)
}

// TestRunCompress_InvalidFile() - tests the runCompress function with an invalid input file
func TestRunCompress_InvalidFile(t *testing.T) {
	cfg := &Config{
		InputPath:    "nonexistent.jpg",
		OutputPath:   "out.jpg",
		MaxSize:      1024,
		OutputFormat: "jpeg",
		Quality:      80,
	}

	if err := runCompress(cfg); err == nil {
		t.Error("expected error for nonexistent input file")
	}
}
