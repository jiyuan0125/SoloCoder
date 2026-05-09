package zstdlib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
	"zstd-tool/protocol"
)

type Compressor struct {
	dictionaries map[string][]byte
	mu           sync.RWMutex
}

func NewCompressor() *Compressor {
	return &Compressor{
		dictionaries: make(map[string][]byte),
	}
}

func (c *Compressor) LoadDictionary(name string, dict []byte) error {
	if len(dict) == 0 {
		return ErrEmptyDictionary
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.dictionaries[name] = dict
	return nil
}

func (c *Compressor) GetDictionary(name string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dict, ok := c.dictionaries[name]
	return dict, ok
}

func (c *Compressor) CompressStream(r io.Reader, w io.Writer, level protocol.CompressionLevel, dictName string) (originalSize, compressedSize int64, err error) {
	if !ValidateLevel(level) {
		return 0, 0, fmt.Errorf("invalid compression level: %d", level)
	}

	var dict []byte
	if dictName != "" {
		var ok bool
		dict, ok = c.GetDictionary(dictName)
		if !ok {
			return 0, 0, ErrDictionaryNotFound
		}
	}

	levelInt := LevelToInt[level]

	var dataBuffer bytes.Buffer
	tee := io.TeeReader(r, &dataBuffer)

	buf := make([]byte, DefaultBufferSize)
	for {
		n, readErr := tee.Read(buf)
		if n > 0 {
			originalSize += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return 0, 0, readErr
		}
	}

	var compressor *zstd.Encoder
	if len(dict) > 0 {
		compressor, err = zstd.NewWriter(nil,
			zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(levelInt)),
			zstd.WithEncoderDict(dict),
		)
	} else {
		compressor, err = zstd.NewWriter(nil,
			zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(levelInt)),
		)
	}
	if err != nil {
		return 0, 0, err
	}

	compressedData := compressor.EncodeAll(dataBuffer.Bytes(), nil)
	compressedSize = int64(len(compressedData))

	if originalSize > (1 << 32) {
		return 0, 0, errors.New("file too large, exceeds 4GB")
	}
	if compressedSize > (1 << 32) {
		return 0, 0, errors.New("compressed data too large")
	}

	if err := WriteHeader(w, uint32(originalSize), uint32(compressedSize), level); err != nil {
		return 0, 0, err
	}

	n, err := w.Write(compressedData)
	if err != nil {
		return 0, 0, err
	}
	compressedSize = int64(n)

	return originalSize, compressedSize, nil
}

func (c *Compressor) CompressFile(inputPath, outputPath string, level protocol.CompressionLevel, dictName string) (originalSize, compressedSize int64, err error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return 0, 0, err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return 0, 0, err
	}
	defer outputFile.Close()

	return c.CompressStream(inputFile, outputFile, level, dictName)
}

func (c *Compressor) DecompressStream(r io.Reader, w io.Writer, dictName string) (originalSize, compressedSize int64, err error) {
	header, err := ReadHeader(r)
	if err != nil {
		return 0, 0, err
	}

	var dict []byte
	if dictName != "" {
		var ok bool
		dict, ok = c.GetDictionary(dictName)
		if !ok {
			return 0, 0, ErrDictionaryNotFound
		}
	}

	compressedBuffer := make([]byte, header.CompressedSize)
	n, readErr := io.ReadFull(r, compressedBuffer)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return 0, 0, readErr
	}
	if uint32(n) != header.CompressedSize {
		return 0, 0, ErrCorruptedData
	}
	compressedSize = int64(n)

	var decompressed []byte
	if len(dict) > 0 {
		decoder, decErr := zstd.NewReader(nil, zstd.WithDecoderDicts(dict))
		if decErr != nil {
			return 0, 0, decErr
		}
		decompressed, err = decoder.DecodeAll(compressedBuffer[:n], nil)
	} else {
		decoder, decErr := zstd.NewReader(nil)
		if decErr != nil {
			return 0, 0, decErr
		}
		decompressed, err = decoder.DecodeAll(compressedBuffer[:n], nil)
	}

	if err != nil {
		if strings.Contains(err.Error(), "corrupt") || strings.Contains(err.Error(), "truncated") {
			return 0, 0, ErrCorruptedData
		}
		return 0, 0, err
	}

	originalSize = int64(len(decompressed))

	if originalSize != int64(header.OriginalSize) {
		return 0, 0, ErrCorruptedData
	}

	if _, err := w.Write(decompressed); err != nil {
		return 0, 0, err
	}

	return originalSize, compressedSize, nil
}

func (c *Compressor) DecompressFile(inputPath, outputPath string, dictName string) (originalSize, compressedSize int64, err error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return 0, 0, err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return 0, 0, err
	}
	defer outputFile.Close()

	return c.DecompressStream(inputFile, outputFile, dictName)
}

func (c *Compressor) CompressBatch(inputPaths []string, outputDir string, level protocol.CompressionLevel, dictName string) ([]BatchResult, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	results := make([]BatchResult, len(inputPaths))
	var wg sync.WaitGroup
	resultChan := make(chan BatchResult, len(inputPaths))

	semaphore := make(chan struct{}, 4)

	for i, inputPath := range inputPaths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := BatchResult{
				Index:        i,
				OriginalPath: path,
			}

			filename := filepath.Base(path)
			outputPath := filepath.Join(outputDir, filename+".zst")

			original, compressed, err := c.CompressFile(path, outputPath, level, dictName)
			if err != nil {
				result.Error = err.Error()
				result.Success = false
			} else {
				result.Success = true
				result.OriginalSize = original
				result.CompressedSize = compressed
				result.OutputPath = outputPath
			}

			resultChan <- result
		}(i, inputPath)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results[result.Index] = result
	}

	return results, nil
}

func (c *Compressor) DecompressBatch(inputPaths []string, outputDir string, dictName string) ([]BatchResult, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	results := make([]BatchResult, len(inputPaths))
	var wg sync.WaitGroup
	resultChan := make(chan BatchResult, len(inputPaths))

	semaphore := make(chan struct{}, 4)

	for i, inputPath := range inputPaths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := BatchResult{
				Index:        i,
				OriginalPath: path,
			}

			filename := filepath.Base(path)
			outputPath := filepath.Join(outputDir, strings.TrimSuffix(filename, ".zst"))

			original, compressed, err := c.DecompressFile(path, outputPath, dictName)
			if err != nil {
				result.Error = err.Error()
				result.Success = false
			} else {
				result.Success = true
				result.OriginalSize = original
				result.CompressedSize = compressed
				result.OutputPath = outputPath
			}

			resultChan <- result
		}(i, inputPath)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results[result.Index] = result
	}

	return results, nil
}

type BatchResult struct {
	Index          int
	OriginalPath   string
	OutputPath     string
	OriginalSize   int64
	CompressedSize int64
	Success        bool
	Error          string
}
