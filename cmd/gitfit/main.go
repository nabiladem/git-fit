package main

import (
	"flag"
	"fmt"
	"git-fit/internal/compressor"
	"os"
	"path/filepath"
	"strings"
)

// Config holds parsed command-line options
type Config struct {
	InputPath    string
	OutputPath   string
	MaxSize      int
	OutputFormat string
	Quality      int
	Verbose      bool
}

func main() {
	cfg := parseFlags()
	showUsage, err := validateConfig(cfg)
	if showUsage {
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := runCompress(cfg); err != nil {
		fmt.Println("Error compressing image:", err)
		os.Exit(1)
	}

	fmt.Println("Image compressed successfully!")
}

// parseFlags() - extract flags into a Config struct
func parseFlags() *Config {
	inputPath := flag.String("input", "", "Path to the input image file")
	outputPath := flag.String("output", "", "Path to save the compressed image")
	maxSize := flag.Int("maxsize", 1048576, "Maximum file size in bytes (default 1MB)")
	outputFormat := flag.String("format", "", "Output image format (jpeg, png, or gif)")
	quality := flag.Int("quality", 85, "JPEG compression quality (1-100; 85 by default)")
	verbose := flag.Bool("v", false, "Verbose logging enabled")

	flag.Usage = func() {
		fmt.Println("Usage: gitfit -input <input-image-file> -output <output-image-file> -maxsize <max size in bytes> -format <jpeg|png|gif> -quality <0-100> -v [for verbose logging]")
		fmt.Println("Example: gitfit -input input.jpeg -output output.jpeg -maxsize 1000000 -format jpeg -quality 85 -v")
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	return &Config{
		InputPath:    *inputPath,
		OutputPath:   *outputPath,
		MaxSize:      *maxSize,
		OutputFormat: *outputFormat,
		Quality:      *quality,
		Verbose:      *verbose,
	}
}

// validateConfig() - perform validations and sets defaults and returns if usage should be shown and error
func validateConfig(cfg *Config) (bool, error) {
	if cfg.InputPath == "" || cfg.OutputPath == "" {
		if cfg.InputPath == "" && cfg.OutputPath == "" {
			return true, nil
		}
		return false, fmt.Errorf("you must provide both -input and -output file paths")
	}

	if _, err := os.Stat(cfg.InputPath); os.IsNotExist(err) {
		return false, fmt.Errorf("input file %s does not exist", cfg.InputPath)
	}

	if cfg.OutputFormat == "" {
		extension := strings.ToLower(filepath.Ext(cfg.InputPath))
		switch extension {
		case ".png":
			cfg.OutputFormat = "png"
		case ".gif":
			cfg.OutputFormat = "gif"
		default:
			cfg.OutputFormat = "jpeg"
		}
	}

	if filepath.Ext(cfg.OutputPath) == "" {
		cfg.OutputPath = cfg.OutputPath + "." + cfg.OutputFormat
	}

	if cfg.Quality <= 0 || cfg.Quality > 100 {
		return false, fmt.Errorf("value for -quality must be between 1 and 100 inclusive")
	}

	if cfg.Verbose && cfg.Quality == 85 {
		fmt.Println("Using default quality of 85 for JPEG compression.")
	}

	if cfg.Verbose {
		fmt.Printf("Input file: %s\nOutput file: %s\nMaximum size: %d\nOutput format: %s\n",
			cfg.InputPath, cfg.OutputPath, cfg.MaxSize, cfg.OutputFormat)
		if cfg.OutputFormat == "jpeg" {
			fmt.Println("Quality:", cfg.Quality)
		}
	}

	return false, nil
}

// runCompress() - call the compressor with the provided Config
func runCompress(cfg *Config) error {
	return compressor.CompressImage(cfg.InputPath, cfg.OutputPath, cfg.MaxSize, cfg.OutputFormat, cfg.Quality, cfg.Verbose)
}
