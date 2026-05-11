package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"tar-archiver/common"
)

var (
	serverURL = flag.String("server", "http://localhost:8302", "Server URL")
	format    = flag.String("format", common.FormatPOSIX, "Archive format: posix or gnu")
	extract   = flag.Bool("extract", false, "Extract mode")
	output    = flag.String("output", "", "Output file for archive or destination directory for extract")
)

func main() {
	flag.Parse()
	paths := flag.Args()

	if *extract {
		if len(paths) == 0 {
			fmt.Println("Usage: client --extract <tar-file> [--output <dest-dir>]")
			os.Exit(1)
		}
		if err := doExtract(paths[0], *output); err != nil {
			fmt.Fprintf(os.Stderr, "Extract failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if len(paths) == 0 {
			fmt.Println("Usage: client [--format posix|gnu] [--output <file.tar>] <path>...")
			os.Exit(1)
		}
		if err := doArchive(paths, *format, *output); err != nil {
			fmt.Fprintf(os.Stderr, "Archive failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func doArchive(paths []string, fmtStr, outputPath string) error {
	req := common.ArchiveRequest{
		Paths:  paths,
		Format: fmtStr,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(*serverURL+"/archive", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(errBody))
	}

	if outputPath == "" || outputPath == "-" {
		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func doExtract(tarPath, destDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	url := *serverURL + "/extract"
	if destDir != "" {
		url += "?dest=" + destDir
	}

	resp, err := http.Post(url, "application/octet-stream", f)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(errBody))
	}

	fmt.Println("Extraction completed")
	return nil
}
