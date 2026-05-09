package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultServerURL = "http://localhost:8080"
)

type Client struct {
	serverURL string
	client    *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimRight(serverURL, "/"),
		client:    &http.Client{},
	}
}

func (c *Client) compressFile(inputPath, outputPath, level, dictName string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(inputPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	if level != "" {
		if err := writer.WriteField("level", level); err != nil {
			return fmt.Errorf("failed to write level field: %v", err)
		}
	}

	if dictName != "" {
		if err := writer.WriteField("dict", dictName); err != nil {
			return fmt.Errorf("failed to write dict field: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", c.serverURL+"/api/compress", &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	originalSize := resp.Header.Get("X-Original-Size")
	compressedSize := resp.Header.Get("X-Compressed-Size")

	if _, err := io.Copy(outputFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("Compressed: %s -> %s (original: %s bytes, compressed: %s bytes)\n",
		inputPath, outputPath, originalSize, compressedSize)

	return nil
}

func (c *Client) decompressFile(inputPath, outputPath, dictName string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(inputPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	if dictName != "" {
		if err := writer.WriteField("dict", dictName); err != nil {
			return fmt.Errorf("failed to write dict field: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", c.serverURL+"/api/decompress", &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	originalSize := resp.Header.Get("X-Original-Size")
	compressedSize := resp.Header.Get("X-Compressed-Size")

	if _, err := io.Copy(outputFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("Decompressed: %s -> %s (compressed: %s bytes, original: %s bytes)\n",
		inputPath, outputPath, compressedSize, originalSize)

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [options]")
		fmt.Println("Commands: compress, decompress")
		os.Exit(1)
	}

	command := os.Args[1]

	compressCmd := flag.NewFlagSet("compress", flag.ExitOnError)
	compressInput := compressCmd.String("input", "", "Input file path")
	compressOutput := compressCmd.String("output", "", "Output file or directory path")
	compressLevel := compressCmd.String("level", "default", "Compression level: fastest, default, best")
	compressDict := compressCmd.String("dict", "", "Dictionary name to use")
	compressBatch := compressCmd.String("batch", "", "Directory to batch compress")
	compressServer := compressCmd.String("server", defaultServerURL, "Server URL")

	decompressCmd := flag.NewFlagSet("decompress", flag.ExitOnError)
	decompressInput := decompressCmd.String("input", "", "Input file path")
	decompressOutput := decompressCmd.String("output", "", "Output file or directory path")
	decompressDict := decompressCmd.String("dict", "", "Dictionary name to use")
	decompressBatch := decompressCmd.String("batch", "", "Directory to batch decompress")
	decompressServer := decompressCmd.String("server", defaultServerURL, "Server URL")

	switch command {
	case "compress":
		if err := compressCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse command: %v\n", err)
			os.Exit(1)
		}

		client := NewClient(*compressServer)

		if *compressBatch != "" {
			files, err := filepath.Glob(filepath.Join(*compressBatch, "*"))
			if err != nil {
				fmt.Printf("Failed to list files: %v\n", err)
				os.Exit(1)
			}

			var validFiles []string
			for _, f := range files {
				info, err := os.Stat(f)
				if err == nil && !info.IsDir() {
					validFiles = append(validFiles, f)
				}
			}

			if len(validFiles) == 0 {
				fmt.Println("No files found in batch directory")
				os.Exit(1)
			}

			if *compressOutput == "" {
				*compressOutput = *compressBatch
			}

			if err := os.MkdirAll(*compressOutput, 0755); err != nil {
				fmt.Printf("Failed to create output directory: %v\n", err)
				os.Exit(1)
			}

			for i, inputPath := range validFiles {
				filename := filepath.Base(inputPath)
				outputPath := filepath.Join(*compressOutput, filename+".zst")

				fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(validFiles), filename)

				if err := client.compressFile(inputPath, outputPath, *compressLevel, *compressDict); err != nil {
					fmt.Printf("Error compressing %s: %v\n", inputPath, err)
				}
			}

			fmt.Printf("Batch compression complete. Processed %d files.\n", len(validFiles))
		} else {
			if *compressInput == "" {
				fmt.Println("Error: --input is required")
				os.Exit(1)
			}

			if *compressOutput == "" {
				*compressOutput = *compressInput + ".zst"
			}

			if info, err := os.Stat(*compressOutput); err == nil && info.IsDir() {
				filename := filepath.Base(*compressInput)
				*compressOutput = filepath.Join(*compressOutput, filename+".zst")
			}

			if err := client.compressFile(*compressInput, *compressOutput, *compressLevel, *compressDict); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

	case "decompress":
		if err := decompressCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse command: %v\n", err)
			os.Exit(1)
		}

		client := NewClient(*decompressServer)

		if *decompressBatch != "" {
			files, err := filepath.Glob(filepath.Join(*decompressBatch, "*.zst"))
			if err != nil {
				fmt.Printf("Failed to list files: %v\n", err)
				os.Exit(1)
			}

			if len(files) == 0 {
				fmt.Println("No .zst files found in batch directory")
				os.Exit(1)
			}

			if *decompressOutput == "" {
				*decompressOutput = *decompressBatch
			}

			if err := os.MkdirAll(*decompressOutput, 0755); err != nil {
				fmt.Printf("Failed to create output directory: %v\n", err)
				os.Exit(1)
			}

			for i, inputPath := range files {
				filename := filepath.Base(inputPath)
				originalFilename := filename
				if len(originalFilename) > 4 && originalFilename[len(originalFilename)-4:] == ".zst" {
					originalFilename = originalFilename[:len(originalFilename)-4]
				}
				outputPath := filepath.Join(*decompressOutput, originalFilename)

				fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(files), filename)

				if err := client.decompressFile(inputPath, outputPath, *decompressDict); err != nil {
					fmt.Printf("Error decompressing %s: %v\n", inputPath, err)
				}
			}

			fmt.Printf("Batch decompression complete. Processed %d files.\n", len(files))
		} else {
			if *decompressInput == "" {
				fmt.Println("Error: --input is required")
				os.Exit(1)
			}

			if *decompressOutput == "" {
				originalFilename := filepath.Base(*decompressInput)
				if len(originalFilename) > 4 && originalFilename[len(originalFilename)-4:] == ".zst" {
					originalFilename = originalFilename[:len(originalFilename)-4]
				}
				dir := filepath.Dir(*decompressInput)
				*decompressOutput = filepath.Join(dir, originalFilename)
			}

			if info, err := os.Stat(*decompressOutput); err == nil && info.IsDir() {
				originalFilename := filepath.Base(*decompressInput)
				if len(originalFilename) > 4 && originalFilename[len(originalFilename)-4:] == ".zst" {
					originalFilename = originalFilename[:len(originalFilename)-4]
				}
				*decompressOutput = filepath.Join(*decompressOutput, originalFilename)
			}

			if err := client.decompressFile(*decompressInput, *decompressOutput, *decompressDict); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: client <command> [options]")
		fmt.Println("Commands: compress, decompress")
		os.Exit(1)
	}
}
