package utils

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"databackup/pkg/models"
)

func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ReadMetadata(path string) (*models.Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Metadata{BackupRecords: []models.BackupRecord{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	var meta models.Metadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&meta); err != nil {
		if err == io.EOF {
			return &models.Metadata{BackupRecords: []models.BackupRecord{}}, nil
		}
		return nil, err
	}
	return &meta, nil
}

func WriteMetadata(path string, meta *models.Metadata) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(meta)
}

func ReadResumeInfo(path string) (*models.ResumeInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var info models.ResumeInfo
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func WriteResumeInfo(path string, info *models.ResumeInfo) error {
	info.LastModified = time.Now()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func EnsureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func FileChanged(srcInfo, lastInfo models.FileInfo) bool {
	if srcInfo.Size != lastInfo.Size {
		return true
	}
	return !srcInfo.ModTime.Equal(lastInfo.ModTime)
}

func CopyFile(src, dst string, chunkSize int64, progressCallback func(written int64, total int64)) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buffer := make([]byte, chunkSize)
	var totalWritten int64

	for {
		n, readErr := srcFile.Read(buffer)
		if n > 0 {
			_, writeErr := dstFile.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			totalWritten += int64(n)
			if progressCallback != nil {
				progressCallback(totalWritten, srcStat.Size())
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return dstFile.Sync()
}

func AddFileToZip(zw *zip.Writer, srcPath, relPath string, chunkSize int64, progressCallback func(written int64, total int64)) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(srcStat)
	if err != nil {
		return err
	}
	header.Name = relPath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	buffer := make([]byte, chunkSize)
	var totalWritten int64

	for {
		n, readErr := srcFile.Read(buffer)
		if n > 0 {
			_, writeErr := writer.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			totalWritten += int64(n)
			if progressCallback != nil {
				progressCallback(totalWritten, srcStat.Size())
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}

func ExtractFileFromZip(zipReader *zip.ReadCloser, fileName, dstPath string) error {
	var file *zip.File
	for _, f := range zipReader.File {
		if f.Name == fileName {
			file = f
			break
		}
	}

	if file == nil {
		return fmt.Errorf("file not found in zip: %s", fileName)
	}

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dstDir := filepath.Dir(dstPath)
	if err := EnsureDirExists(dstDir); err != nil {
		return err
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, rc)
	return err
}
