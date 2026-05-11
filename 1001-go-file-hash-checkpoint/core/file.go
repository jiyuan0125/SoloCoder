package core

import (
	"fmt"
	"io"
	"os"

	"hash-checkpoint/common"
)

type FileChunkReader struct {
	filePath  string
	file      *os.File
	chunkSize int64
	fileSize  int64
}

func NewFileChunkReader(filePath string, chunkSize int64) (*FileChunkReader, error) {
	if chunkSize <= 0 {
		chunkSize = common.DefaultChunkSize
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &FileChunkReader{
		filePath:  filePath,
		file:      file,
		chunkSize: chunkSize,
		fileSize:  info.Size(),
	}, nil
}

func (r *FileChunkReader) Close() error {
	return r.file.Close()
}

func (r *FileChunkReader) FileSize() int64 {
	return r.fileSize
}

func (r *FileChunkReader) ReadChunk(index int) ([]byte, int64, error) {
	chunkLen := CalculateChunkLen(index, r.fileSize, r.chunkSize)
	if chunkLen == 0 {
		return nil, 0, io.EOF
	}

	offset := int64(index) * r.chunkSize
	if offset >= r.fileSize {
		return nil, 0, io.EOF
	}

	buf := make([]byte, chunkLen)
	n, err := r.file.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, 0, err
	}
	if int64(n) != chunkLen {
		return nil, 0, fmt.Errorf("expected to read %d bytes, got %d", chunkLen, n)
	}

	return buf, chunkLen, nil
}

func (r *FileChunkReader) CalculateChunkHash(index int, algorithm common.HashAlgorithm) (string, int64, error) {
	chunkData, chunkLen, err := r.ReadChunk(index)
	if err != nil {
		return "", 0, err
	}
	if len(chunkData) == 0 {
		return "", 0, nil
	}
	hash, err := CalculateChunkHash(chunkData, algorithm)
	if err != nil {
		return "", 0, err
	}
	return hash, chunkLen, nil
}

func (r *FileChunkReader) CalculateWholeFileHash(algorithm common.HashAlgorithm) (string, error) {
	_, err := r.file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	return CalculateStreamHash(r.file, algorithm)
}

func CalculateFinalHashFromChunks(filePath string, chunkSize int64, algorithm common.HashAlgorithm) (string, error) {
	reader, err := NewFileChunkReader(filePath, chunkSize)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	if reader.FileSize() <= 0 {
		return common.GetEmptyFileHash(algorithm), nil
	}

	totalChunks := CalculateTotalChunks(reader.FileSize(), chunkSize)
	chunkHashes := make([]string, 0, totalChunks)

	for i := 0; i < totalChunks; i++ {
		hash, _, err := reader.CalculateChunkHash(i, algorithm)
		if err != nil {
			return "", err
		}
		chunkHashes = append(chunkHashes, hash)
	}

	return CalculateFinalHash(chunkHashes, algorithm)
}

func GetFileInfo(filePath string) (size int64, modTime interface{}, err error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, nil, err
	}
	return info.Size(), info.ModTime(), nil
}
