package main

import (
    "image"
    "image/color"
    "image/jpeg"
    "os"
    "path/filepath"
    "testing"
)

// helper to create a small JPEG file for tests
func writeTestJPEG(path string, w, h, quality int) error {
    img := image.NewRGBA(image.Rect(0, 0, w, h))
    col := color.RGBA{R: 180, G: 120, B: 80, A: 255}

	// fill with solid color
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            img.Set(x, y, col)
        }
    }

    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

// TestValidateConfig_ShowUsageWhenEmpty() - test that validateConfig signals to show usage when config is empty
/* t (*testing.T) - testing object */
func TestValidateConfig_ShowUsageWhenEmpty(t *testing.T) {
    cfg := &Config{}
    show, err := validateConfig(cfg)
    if !show || err != nil {
        t.Fatalf("expected showUsage=true and err=nil for empty cfg; got show=%v err=%v", show, err)
    }
}

// TestValidateConfig_MissingOneFlag() - test that validateConfig returns error when one of input/output is missing
/* t (*testing.T) - testing object */
func TestValidateConfig_MissingOneFlag(t *testing.T) {
    cfg := &Config{OutputPath: "out.jpg"}
    show, err := validateConfig(cfg)
    if show || err == nil {
        t.Fatalf("expected showUsage=false and err!=nil when one of input/output is missing; got show=%v err=%v", show, err)
    }
}

// TestValidateConfig_InputFileDoesNotExist() - test that validateConfig returns error when input file does not exist
/* t (*testing.T) - testing object */
func TestValidateConfig_InputFileDoesNotExist(t *testing.T) {
    cfg := &Config{InputPath: "no-such-file-xyz.jpg", OutputPath: "out.jpg"}
    _, err := validateConfig(cfg)
    if err == nil {
        t.Fatalf("expected error for non-existent input file")
    }
}

// TestValidateConfig_DefaultFormatAndExtension() - test that validateConfig sets default output format and appends extension
/* t (*testing.T) - testing object */
func TestValidateConfig_DefaultFormatAndExtension(t *testing.T) {
    td := t.TempDir()
    in := filepath.Join(td, "in.png")

    // create empty file to satisfy Stat
    if err := os.WriteFile(in, []byte(""), 0644); err != nil {
        t.Fatalf("failed to create input file: %v", err)
    }

    out := filepath.Join(td, "out")
    cfg := &Config{InputPath: in, OutputPath: out, OutputFormat: "", Quality: 85}
    show, err := validateConfig(cfg)
    if show || err != nil {
        t.Fatalf("unexpected validateConfig result: show=%v err=%v", show, err)
    }

    if cfg.OutputFormat != "png" {
        t.Fatalf("expected OutputFormat=png, got %s", cfg.OutputFormat)
    }

    if filepath.Ext(cfg.OutputPath) != ".png" {
        t.Fatalf("expected output path to have .png extension; got %s", cfg.OutputPath)
    }
}

// TestValidateConfig_QualityRange() - test that validateConfig returns error for invalid quality range
/* t (*testing.T) - testing object */
func TestValidateConfig_QualityRange(t *testing.T) {
    td := t.TempDir()
    in := filepath.Join(td, "in.jpg")
    if err := writeTestJPEG(in, 20, 20, 80); err != nil {
        t.Fatalf("failed to write in.jpg: %v", err)
    }

    cfg := &Config{InputPath: in, OutputPath: filepath.Join(td, "out.jpg"), Quality: 0}
    _, err := validateConfig(cfg)
    if err == nil {
        t.Fatalf("expected error for invalid quality range")
    }
}

// TestRunCompress_EndToEnd() - end-to-end test of runCompress()
/* t (*testing.T) - testing object */
func TestRunCompress_EndToEnd(t *testing.T) {
    td := t.TempDir()
    in := filepath.Join(td, "in.jpg")
    out := filepath.Join(td, "out.jpg")
    if err := writeTestJPEG(in, 320, 240, 100); err != nil {
        t.Fatalf("failed to write test jpeg: %v", err)
    }

	// run compression
    cfg := &Config{InputPath: in, OutputPath: out, MaxSize: 5 * 1024 * 1024, OutputFormat: "jpeg", Quality: 90, Verbose: false}
    if err := runCompress(cfg); err != nil {
        t.Fatalf("runCompress failed: %v", err)
    }

    fi, err := os.Stat(out)
    if err != nil {
        t.Fatalf("expected output file, got error: %v", err)
    }

    if fi.Size() == 0 {
        t.Fatalf("output file is empty")
    }
}
