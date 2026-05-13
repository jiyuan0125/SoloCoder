package logger

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"log-rotate/internal/repository"
)

type LogQuerier struct {
	repo   *repository.Repository
	logDir string
}

func NewLogQuerier(repo *repository.Repository, logDir string) *LogQuerier {
	return &LogQuerier{
		repo:   repo,
		logDir: logDir,
	}
}

func (q *LogQuerier) QueryCurrent(limit int) ([]string, error) {
	currentPath := filepath.Join(q.logDir, "app.log")
	return q.readLinesFromFile(currentPath, limit)
}

func (q *LogQuerier) QueryArchive(fileName string, limit int) ([]string, error) {
	archive, err := q.repo.GetArchiveByName(fileName)
	if err != nil {
		return nil, fmt.Errorf("archive not found: %w", err)
	}

	var filePath string
	if archive.Compressed {
		filePath = filepath.Join(q.logDir, fileName+".gz")
	} else {
		filePath = filepath.Join(q.logDir, fileName)
	}

	if archive.Compressed {
		return q.queryCompressed(filePath, limit)
	}
	return q.readLinesFromFile(filePath, limit)
}

func (q *LogQuerier) queryCompressed(gzPath string, limit int) ([]string, error) {
	uncompressedPath := strings.TrimSuffix(gzPath, ".gz")

	if _, err := os.Stat(uncompressedPath); os.IsNotExist(err) {
		if err := q.decompressFile(gzPath, uncompressedPath); err != nil {
			return nil, fmt.Errorf("decompress failed: %w", err)
		}
	}

	lines, err := q.readLinesFromFile(uncompressedPath, limit)
	if err != nil {
		return lines, err
	}

	go func() {
		if err := q.compressAfterQuery(uncompressedPath); err != nil {
			fmt.Printf("recompress failed: %v\n", err)
		}
	}()

	return lines, nil
}

func (q *LogQuerier) decompressFile(gzPath, outPath string) error {
	gzFile, err := os.Open(gzPath)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	gr, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer gr.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gr)
	return err
}

func (q *LogQuerier) compressAfterQuery(path string) error {
	gzPath := path + ".gz"

	inFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(gzPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzw := gzip.NewWriter(outFile)
	if _, err := io.Copy(gzw, inFile); err != nil {
		gzw.Close()
		return err
	}

	if err := gzw.Close(); err != nil {
		return err
	}

	os.Remove(path)
	return nil
}

func (q *LogQuerier) readLinesFromFile(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	buf := make([]byte, 2*1024*1024)
	scanner.Buffer(buf, 2*1024*1024)

	count := 0
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}
