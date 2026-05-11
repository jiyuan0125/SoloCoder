package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/solocoder/brotli/pkg/api"
)

func main() {
	var (
		serverURL  string
		quality    int
		compress   bool
		decompress bool
		fileInput  string
		textInput  string
	)

	flag.StringVar(&serverURL, "server", "http://localhost:8104", "Brotli server URL")
	flag.IntVar(&quality, "q", 6, "Compression quality (1-11, default 6)")
	flag.BoolVar(&compress, "c", false, "Compress mode (default)")
	flag.BoolVar(&decompress, "d", false, "Decompress mode")
	flag.StringVar(&fileInput, "f", "", "Input file path")
	flag.StringVar(&textInput, "t", "", "Input text directly")

	flag.Parse()

	if decompress && compress {
		log.Fatal("Cannot specify both -c and -d")
	}

	quality = api.ValidateQuality(quality)

	var data []byte
	var err error

	if fileInput != "" {
		data, err = ioutil.ReadFile(fileInput)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
	} else if textInput != "" {
		data = []byte(textInput)
	} else if flag.NArg() > 0 {
		data = []byte(strings.Join(flag.Args(), " "))
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("Failed to read stdin: %v", err)
			}
		} else {
			flag.Usage()
			fmt.Println("\nExamples:")
			fmt.Println("  brotli-cli -c -t \"hello world\"")
			fmt.Println("  brotli-cli -c -f input.txt")
			fmt.Println("  brotli-cli -d -f compressed.br")
			fmt.Println("  echo \"hello\" | brotli-cli -c")
			os.Exit(1)
		}
	}

	if decompress {
		result, err := callDecompress(serverURL, data)
		if err != nil {
			log.Fatalf("Decompression failed: %v", err)
		}
		fmt.Print(string(result))
	} else {
		result, err := callCompress(serverURL, data, quality)
		if err != nil {
			log.Fatalf("Compression failed: %v", err)
		}
		os.Stdout.Write(result)
	}
}

func callCompress(serverURL string, data []byte, quality int) ([]byte, error) {
	req := api.CompressRequest{
		Data:    data,
		Quality: quality,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/compress", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result api.CompressResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Compressed: %d -> %d bytes (%.2f%%) with quality %d\n",
		result.OriginalSize, result.CompressedSize,
		float64(result.CompressedSize)/float64(result.OriginalSize)*100,
		result.Quality)

	return result.Data, nil
}

func callDecompress(serverURL string, data []byte) ([]byte, error) {
	req := api.DecompressRequest{
		Data: data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/decompress", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result api.DecompressResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Decompressed: %d -> %d bytes\n",
		result.CompressedSize, result.OriginalSize)

	return result.Data, nil
}
