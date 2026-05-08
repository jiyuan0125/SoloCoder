package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Server     string
	Content    string
	ErrorLevel string
	Size       int
	LogoPath   string
	OutputPath string
}

func main() {
	cfg := parseFlags()

	if cfg.Content == "" {
		fmt.Println("Error: content is required")
		flag.Usage()
		os.Exit(1)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("content", cfg.Content)
	if cfg.ErrorLevel != "" {
		_ = writer.WriteField("error_level", cfg.ErrorLevel)
	}
	if cfg.Size > 0 {
		_ = writer.WriteField("size", strconv.Itoa(cfg.Size))
	}

	if cfg.LogoPath != "" {
		logoFile, err := os.Open(cfg.LogoPath)
		if err != nil {
			fmt.Printf("Error: failed to open logo file: %v\n", err)
			os.Exit(1)
		}
		defer logoFile.Close()

		part, err := writer.CreateFormFile("logo", filepath.Base(cfg.LogoPath))
		if err != nil {
			fmt.Printf("Error: failed to create form file: %v\n", err)
			os.Exit(1)
		}
		if _, err := io.Copy(part, logoFile); err != nil {
			fmt.Printf("Error: failed to copy logo file: %v\n", err)
			os.Exit(1)
		}
	}

	if err := writer.Close(); err != nil {
		fmt.Printf("Error: failed to close multipart writer: %v\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", cfg.Server+"/generate", &body)
	if err != nil {
		fmt.Printf("Error: failed to create request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: failed to send request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: server returned status %d: %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = "qrcode.png"
	}

	if err := os.WriteFile(outputPath, respBody, 0644); err != nil {
		fmt.Printf("Error: failed to write output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("QR code generated successfully and saved to: %s\n", outputPath)
}

func parseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.Server, "server", "http://localhost:8080", "QR code server URL")
	flag.StringVar(&cfg.Content, "content", "", "Content to encode in QR code (required)")
	flag.StringVar(&cfg.ErrorLevel, "error-level", "", "Error correction level: L, M, Q, H (default: M)")
	flag.IntVar(&cfg.Size, "size", 0, "QR code size in pixels (default: 300)")
	flag.StringVar(&cfg.LogoPath, "logo", "", "Path to logo image file (PNG or JPEG)")
	flag.StringVar(&cfg.OutputPath, "output", "", "Output file path (default: qrcode.png)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	return cfg
}
