package main

import (
"flag"
"fmt"
"os"
"path/filepath"
"strings"

"github.com/nabiladem/git-fit/internal/compressor"
"github.com/nabiladem/git-fit/internal/gravatar"
)

// Config holds parsed command-line options
/* InputPath (string) - path of the input image file; OutputPath (string) - path to save the compressed image
   MaxSize (int) - maximum size of the image in bytes; OutputFormat (string) - jpeg, png, or gif
   Quality (int) - quality for JPEG compression; Verbose (bool) - enable verbose logging */
type Config struct {
	InputPath      string
	OutputPath     string
	MaxSize        int
	OutputFormat   string
	Quality        int
	Verbose        bool
	UploadGravatar bool
}
// main() - entry point
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
    // define command-line flags
	inputPath := flag.String("input", "", "Path to the input image file")
	outputPath := flag.String("output", "", "Path to save the compressed image")
	maxSize := flag.Int("maxsize", 1048576, "Maximum file size in bytes (default 1MB)")
	outputFormat := flag.String("format", "", "Output image format (jpeg, png, or gif)")
	quality := flag.Int("quality", 85, "JPEG compression quality (1-100; 85 by default)")
	verbose := flag.Bool("v", false, "Verbose logging enabled")
uploadGravatar := flag.Bool("upload-gravatar", false, "Upload compressed image to Gravatar")

    // custom usage message for flags
	flag.Usage = func() {
		fmt.Println("Usage: gitfit -input <input-image-file> -output <output-image-file> -maxsize <max size in bytes> " +
                    "-format <jpeg|png|gif> -quality <0-100> -v [for verbose logging]")
		fmt.Println("Example: gitfit -input input.jpeg -output output.jpeg -maxsize 1000000 -format jpeg -quality 85 -v")
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

    // Config struct populated with flag values
	return &Config {
		InputPath:    *inputPath,
		OutputPath:   *outputPath,
		MaxSize:      *maxSize,
		OutputFormat: *outputFormat,
		Quality:      *quality,
		Verbose:      *verbose,
	}
}

// validateConfig() - perform validations and sets defaults and returns if usage should be shown
/* cfg (*Config) - configuration to validate */
func validateConfig(cfg *Config) (bool, error) {
    // check if input and/or output path is missing
    if cfg.InputPath == "" || cfg.OutputPath == "" {
        // assume user knows about both flags if one is given
		if cfg.InputPath == "" && cfg.OutputPath == "" {
			return true, nil
		}

		return false, fmt.Errorf("you must provide both -input and -output file paths")
	}

	if _, err := os.Stat(cfg.InputPath); os.IsNotExist(err) {
		return false, fmt.Errorf("input file %s does not exist", cfg.InputPath)
	}

    // set default output format based on input file extension if not provided
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

    // append appropriate file extension to output path if missing
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
/* cfg (*Config) - configuration for compression */
// runCompress() - call the compressor with the provided Config
/* cfg (*Config) - configuration for compression */
func runCompress(cfg *Config) error {
	err := compressor.CompressImage(cfg.InputPath, cfg.OutputPath, cfg.MaxSize,
cfg.OutputFormat, cfg.Quality, cfg.Verbose)
	if err != nil {
		return err
	}

	if cfg.UploadGravatar {
		if cfg.Verbose {
			fmt.Println("Uploading to Gravatar...")
		}

		// Load OAuth credentials from environment
		clientID := os.Getenv("GRAVATAR_CLIENT_ID")
		clientSecret := os.Getenv("GRAVATAR_CLIENT_SECRET")
		redirectURI := os.Getenv("GRAVATAR_REDIRECT_URI")

		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("GRAVATAR_CLIENT_ID and GRAVATAR_CLIENT_SECRET environment variables must be set")
		}

		// Use default redirect URI if not specified
		if redirectURI == "" {
			redirectURI = "http://localhost:8080/callback"
		}

		// Create OAuth client
		client := gravatar.NewClient(clientID, clientSecret, redirectURI, cfg.Verbose)

		// Perform OAuth authentication
		if cfg.Verbose {
			fmt.Println("Starting OAuth authentication...")
			fmt.Println("Your browser will open for authorization.")
		}

		if err := client.Authenticate(); err != nil {
			return fmt.Errorf("OAuth authentication failed: %v", err)
		}

		// Upload avatar
		if err := client.UploadAvatar(cfg.OutputPath); err != nil {
			return fmt.Errorf("failed to upload to Gravatar: %v", err)
		}

		if cfg.Verbose {
			fmt.Println("Successfully uploaded to Gravatar!")
		}
	}

	return nil
}
