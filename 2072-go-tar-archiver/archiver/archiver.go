package archiver

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxOldFormatSize = 2 * 1024 * 1024 * 1024

type ArchiveResult struct {
	ArchivePath string
	FilesArchived int
	TotalBytes  int64
}

type ExtractResult struct {
	FilesExtracted []string
	FilesSkipped   []string
}

type Progress struct {
	ProcessedBytes int64
	TotalBytes     int64
}

type fileEntry struct {
	Name       string
	Size       int64
	IsDir      bool
	ModTime    time.Time
	Mode       int64
}

func CreateArchive(sourceDir, archivePath string, incrementalDate *time.Time, password string, progressChan chan<- Progress) (*ArchiveResult, error) {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist")
	}

	if info, err := os.Stat(archivePath); err == nil && !info.IsDir() {
		if !isValidTarGz(archivePath) {
			return nil, fmt.Errorf("existing file exists but is not a valid tar.gz")
		}
	}

	tempArchive := archivePath + ".tmp"
	if err := os.RemoveAll(tempArchive); err != nil {
		return nil, err
	}

	totalBytes, err := calculateTotalBytes(sourceDir, incrementalDate)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(tempArchive)
	if err != nil {
		return nil, err
	}

	gzWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzWriter)

	var processedBytes int64
	var filesArchived int
	var mu sync.Mutex

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == sourceDir {
			return nil
		}

		if incrementalDate != nil && info.ModTime().Before(*incrementalDate) {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return addDirectoryToArchive(tarWriter, relPath, info)
		} else {
			mu.Lock()
			processedBytes += info.Size()
			mu.Unlock()

			if progressChan != nil {
				progressChan <- Progress{ProcessedBytes: processedBytes, TotalBytes: totalBytes}
			}

			err = addFileToArchive(tarWriter, path, relPath, info)
			if err != nil {
				return err
			}
			mu.Lock()
			filesArchived++
			mu.Unlock()
		}
		return nil
	})

	if err != nil {
		file.Close()
		os.Remove(tempArchive)
		return nil, err
	}

	if err := tarWriter.Close(); err != nil {
		file.Close()
		os.Remove(tempArchive)
		return nil, err
	}

	if err := gzWriter.Close(); err != nil {
		file.Close()
		os.Remove(tempArchive)
		return nil, err
	}

	if err := file.Close(); err != nil {
		os.Remove(tempArchive)
		return nil, err
	}

	if password != "" {
		encryptedPath := tempArchive + ".enc"
		if err := encryptFile(tempArchive, encryptedPath, password); err != nil {
			os.Remove(tempArchive)
			os.Remove(encryptedPath)
			return nil, err
		}
		if err := os.Rename(encryptedPath, archivePath); err != nil {
			os.Remove(tempArchive)
			os.Remove(encryptedPath)
			return nil, err
		}
		os.Remove(tempArchive)
	} else {
		if err := os.Rename(tempArchive, archivePath); err != nil {
			os.Remove(tempArchive)
			return nil, err
		}
	}

	return &ArchiveResult{
		ArchivePath: archivePath,
		FilesArchived: filesArchived,
		TotalBytes:  totalBytes,
	}, nil
}

func calculateTotalBytes(sourceDir string, incrementalDate *time.Time) (int64, error) {
	var total int64
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == sourceDir {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if incrementalDate != nil && info.ModTime().Before(*incrementalDate) {
			return nil
		}

		total += info.Size()
		return nil
	})
	return total, err
}

func addDirectoryToArchive(tw *tar.Writer, name string, info os.FileInfo) error {
	header := &tar.Header{
		Name:     name + "/",
		Mode:     int64(info.Mode()),
		ModTime:  info.ModTime(),
		Typeflag: tar.TypeDir,
	}

	return tw.WriteHeader(header)
}

func addFileToArchive(tw *tar.Writer, fullPath, relPath string, info os.FileInfo) error {
	header := &tar.Header{
		Name:    relPath,
		Mode:    int64(info.Mode()),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	if info.Size() > maxOldFormatSize {
		header.Format = tar.FormatPAX
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func ExtractArchive(archivePath, targetDir string, password string) (*ExtractResult, error) {
	result := &ExtractResult{
		FilesExtracted: []string{},
		FilesSkipped:   []string{},
	}

	var tempArchive string
	if password != "" {
		tempArchive = archivePath + ".tmp"
		if err := decryptFile(archivePath, tempArchive, password); err != nil {
			return nil, err
		}
		defer os.Remove(tempArchive)
		archivePath = tempArchive
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("not a valid gzip archive: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	extractedFiles := []string{}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			rollbackExtracted(extractedFiles, targetDir)
			return nil, fmt.Errorf("corrupted archive: %v", err)
		}

		targetPath := filepath.Join(targetDir, header.Name)

		if _, err := os.Stat(targetPath); err == nil {
			result.FilesSkipped = append(result.FilesSkipped, header.Name)
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				rollbackExtracted(extractedFiles, targetDir)
				return nil, fmt.Errorf("failed to create directory %s: %v", header.Name, err)
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				rollbackExtracted(extractedFiles, targetDir)
				return nil, fmt.Errorf("failed to create parent directory for %s: %v", header.Name, err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				rollbackExtracted(extractedFiles, targetDir)
				return nil, fmt.Errorf("failed to create file %s: %v", header.Name, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				rollbackExtracted(extractedFiles, targetDir)
				return nil, fmt.Errorf("failed to extract file %s: %v", header.Name, err)
			}

			outFile.Close()

			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				rollbackExtracted(extractedFiles, targetDir)
				return nil, fmt.Errorf("failed to set permissions for %s: %v", header.Name, err)
			}

			if err := os.Chtimes(targetPath, header.AccessTime, header.ModTime); err != nil {
				rollbackExtracted(extractedFiles, targetDir)
				return nil, fmt.Errorf("failed to set times for %s: %v", header.Name, err)
			}

			extractedFiles = append(extractedFiles, targetPath)
			result.FilesExtracted = append(result.FilesExtracted, header.Name)
		}
	}

	return result, nil
}

func rollbackExtracted(files []string, baseDir string) {
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if info, err := os.Stat(file); err == nil {
			if info.IsDir() {
				os.RemoveAll(file)
			} else {
				os.Remove(file)
			}
		}
	}
}

func ListArchive(archivePath, password string) ([]*tar.Header, error) {
	var tempArchive string
	if password != "" {
		tempArchive = archivePath + ".tmp"
		if err := decryptFile(archivePath, tempArchive, password); err != nil {
			return nil, err
		}
		defer os.Remove(tempArchive)
		archivePath = tempArchive
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	entries := []*tar.Header{}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("corrupted archive: %v", err)
		}
		entries = append(entries, header)
	}

	return entries, nil
}

func isValidTarGz(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return false
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	_, err = tarReader.Next()
	return err == nil || err == io.EOF
}

func CalculateArchiveChecksum(archivePath, password string) (string, error) {
	var tempArchive string
	if password != "" {
		tempArchive = archivePath + ".tmp"
		if err := decryptFile(archivePath, tempArchive, password); err != nil {
			return "", err
		}
		defer os.Remove(tempArchive)
		archivePath = tempArchive
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func VerifyArchiveIntegrity(archivePath, password string) error {
	entries, err := ListArchive(archivePath, password)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}
		if strings.Contains(entry.Name, "..") {
			return errors.New("potentially malicious archive entry: contains path traversal")
		}
	}

	return nil
}

func GetArchiveEntryCount(archivePath, password string) (int, error) {
	entries, err := ListArchive(archivePath, password)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}
