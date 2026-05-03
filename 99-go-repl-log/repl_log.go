package replog

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	offsetSize     = 8
	lengthSize     = 8
	headerSize     = offsetSize + lengthSize
	maxFileSize    = 16 * 1024 * 1024
	baseFileNumber = 1
)

var (
	ErrGroupExists       = errors.New("consumer group already exists")
	ErrGroupNotExists    = errors.New("consumer group does not exist")
	ErrInvalidOffset     = errors.New("invalid offset")
	ErrOffsetInUse       = errors.New("offset is still in use by some consumer groups")
	ErrTruncated         = errors.New("log has been truncated, requested offset no longer exists")
	ErrCorruptedLog      = errors.New("corrupted log file")
)

type LogFile struct {
	Number     int
	FileName   string
	StartOffset int64
	EndOffset   int64
}

type Log struct {
	mu           sync.RWMutex
	dir          string
	currentFile  *os.File
	currentFileNum int
	currentOffset int64
	currentSize  int64
	files        []*LogFile
	groups       map[string]int64
	closed       bool
}

func Open(dir string) (*Log, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	log := &Log{
		dir:          dir,
		currentOffset: -1,
		files:        make([]*LogFile, 0),
		groups:       make(map[string]int64),
	}

	if err := log.loadLogFiles(); err != nil {
		return nil, err
	}

	if err := log.loadGroups(); err != nil {
		return nil, err
	}

	if len(log.files) == 0 {
		if err := log.createNewFile(); err != nil {
			return nil, err
		}
	} else {
		if err := log.openLastFile(); err != nil {
			return nil, err
		}
	}

	return log, nil
}

func (l *Log) loadLogFiles() error {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return err
	}

	var logFiles []*LogFile

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "log.") && !strings.HasSuffix(name, ".offset") {
			var fileNum int
			_, err := fmt.Sscanf(name, "log.%04d", &fileNum)
			if err != nil {
				continue
			}

			logFile := &LogFile{
				Number:     fileNum,
				FileName:   name,
			}

			startOffset, endOffset, err := l.scanLogFileOffsets(filepath.Join(l.dir, name))
			if err != nil {
				return err
			}
			logFile.StartOffset = startOffset
			logFile.EndOffset = endOffset

			logFiles = append(logFiles, logFile)
		}
	}

	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Number < logFiles[j].Number
	})

	l.files = logFiles

	if len(logFiles) > 0 {
		lastFile := logFiles[len(logFiles)-1]
		l.currentOffset = lastFile.EndOffset
	}

	return nil
}

func (l *Log) scanLogFileOffsets(filePath string) (startOffset int64, endOffset int64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return -1, -1, err
	}
	defer file.Close()

	startOffset = -1
	endOffset = -1

	header := make([]byte, headerSize)
	for {
		n, err := file.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return -1, -1, err
		}
		if n != headerSize {
			return -1, -1, ErrCorruptedLog
		}

		offset := int64(binary.BigEndian.Uint64(header[:offsetSize]))
		length := int64(binary.BigEndian.Uint64(header[offsetSize:]))

		if startOffset == -1 {
			startOffset = offset
		}
		endOffset = offset

		if _, err := file.Seek(length, io.SeekCurrent); err != nil {
			return -1, -1, err
		}
	}

	return startOffset, endOffset, nil
}

func (l *Log) loadGroups() error {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".offset") {
			groupName := strings.TrimSuffix(name, ".offset")
			filePath := filepath.Join(l.dir, name)

			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			if len(data) != 8 {
				return ErrCorruptedLog
			}

			offset := int64(binary.BigEndian.Uint64(data))
			l.groups[groupName] = offset
		}
	}

	return nil
}

func (l *Log) createNewFile() error {
	if l.currentFile != nil {
		if err := l.currentFile.Close(); err != nil {
			return err
		}
	}

	newNum := baseFileNumber
	if len(l.files) > 0 {
		newNum = l.files[len(l.files)-1].Number + 1
	}

	fileName := fmt.Sprintf("log.%04d", newNum)
	filePath := filepath.Join(l.dir, fileName)

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.currentFile = file
	l.currentFileNum = newNum
	l.currentSize = 0

	logFile := &LogFile{
		Number:     newNum,
		FileName:   fileName,
		StartOffset: l.currentOffset + 1,
		EndOffset:   l.currentOffset,
	}
	l.files = append(l.files, logFile)

	return nil
}

func (l *Log) openLastFile() error {
	if len(l.files) == 0 {
		return nil
	}

	lastFile := l.files[len(l.files)-1]
	filePath := filepath.Join(l.dir, lastFile.FileName)

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.currentFile = file
	l.currentFileNum = lastFile.Number
	l.currentSize = info.Size()

	return nil
}

func (l *Log) Append(entry []byte) (offset int64, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return -1, errors.New("log is closed")
	}

	entrySize := headerSize + int64(len(entry))
	if l.currentSize+entrySize > maxFileSize {
		if err := l.createNewFile(); err != nil {
			return -1, err
		}
	}

	newOffset := l.currentOffset + 1

	buf := make([]byte, headerSize+len(entry))
	binary.BigEndian.PutUint64(buf[:offsetSize], uint64(newOffset))
	binary.BigEndian.PutUint64(buf[offsetSize:headerSize], uint64(len(entry)))
	copy(buf[headerSize:], entry)

	if _, err := l.currentFile.Write(buf); err != nil {
		return -1, err
	}

	if err := l.currentFile.Sync(); err != nil {
		return -1, err
	}

	l.currentOffset = newOffset
	l.currentSize += entrySize

	if len(l.files) > 0 {
		l.files[len(l.files)-1].EndOffset = newOffset
	}

	return newOffset, nil
}

func (l *Log) CreateGroup(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return errors.New("log is closed")
	}

	if _, exists := l.groups[name]; exists {
		return ErrGroupExists
	}

	startOffset := int64(0)
	if len(l.files) > 0 {
		startOffset = l.files[0].StartOffset
	}

	l.groups[name] = startOffset

	return l.saveGroupOffset(name, startOffset)
}

func (l *Log) saveGroupOffset(name string, offset int64) error {
	fileName := fmt.Sprintf("%s.offset", name)
	filePath := filepath.Join(l.dir, fileName)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(offset))

	return os.WriteFile(filePath, buf, 0644)
}

func (l *Log) Read(group string, maxCount int) ([][]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, errors.New("log is closed")
	}

	offset, exists := l.groups[group]
	if !exists {
		return nil, ErrGroupNotExists
	}

	if maxCount <= 0 {
		return [][]byte{}, nil
	}

	if len(l.files) == 0 {
		return [][]byte{}, nil
	}

	if offset > l.currentOffset {
		return [][]byte{}, nil
	}

	firstFile := l.files[0]
	if offset < firstFile.StartOffset {
		return nil, ErrTruncated
	}

	var results [][]byte
	currentOffset := offset
	count := 0

	fileIdx := l.findFileIndexByOffset(currentOffset)
	if fileIdx < 0 {
		return nil, ErrInvalidOffset
	}

	for fileIdx < len(l.files) && count < maxCount {
		logFile := l.files[fileIdx]
		filePath := filepath.Join(l.dir, logFile.FileName)

		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		header := make([]byte, headerSize)
		readFromBeginning := false

		if currentOffset == logFile.StartOffset {
			readFromBeginning = true
		} else {
			for {
				n, err := file.Read(header)
				if err == io.EOF {
					break
				}
				if err != nil {
					file.Close()
					return nil, err
				}
				if n != headerSize {
					file.Close()
					return nil, ErrCorruptedLog
				}

				entryOffset := int64(binary.BigEndian.Uint64(header[:offsetSize]))
				length := int64(binary.BigEndian.Uint64(header[offsetSize:]))

				if entryOffset == currentOffset {
					_, err = file.Seek(-headerSize, io.SeekCurrent)
					if err != nil {
						file.Close()
						return nil, err
					}
					readFromBeginning = true
					break
				}

				if entryOffset > currentOffset {
					file.Close()
					return nil, ErrInvalidOffset
				}

				if _, err := file.Seek(length, io.SeekCurrent); err != nil {
					file.Close()
					return nil, err
				}
			}
		}

		if !readFromBeginning {
			file.Close()
			fileIdx++
			continue
		}

		for count < maxCount {
			n, err := file.Read(header)
			if err == io.EOF {
				break
			}
			if err != nil {
				file.Close()
				return nil, err
			}
			if n != headerSize {
				file.Close()
				return nil, ErrCorruptedLog
			}

			entryOffset := int64(binary.BigEndian.Uint64(header[:offsetSize]))
			length := int64(binary.BigEndian.Uint64(header[offsetSize:]))

			data := make([]byte, length)
			if _, err := io.ReadFull(file, data); err != nil {
				file.Close()
				return nil, err
			}

			results = append(results, data)
			currentOffset = entryOffset + 1
			count++

			if currentOffset > logFile.EndOffset {
				break
			}
		}

		file.Close()

		if currentOffset > logFile.EndOffset {
			fileIdx++
		}
	}

	if len(results) > 0 {
		l.groups[group] = currentOffset
		if err := l.saveGroupOffset(group, currentOffset); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (l *Log) findFileIndexByOffset(offset int64) int {
	for i, f := range l.files {
		if offset >= f.StartOffset && offset <= f.EndOffset {
			return i
		}
	}
	return -1
}

func (l *Log) Truncate(offset int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return errors.New("log is closed")
	}

	if offset < 0 {
		return ErrInvalidOffset
	}

	if len(l.files) == 0 {
		return nil
	}

	firstFile := l.files[0]
	if offset <= firstFile.StartOffset {
		return nil
	}

	for _, groupOffset := range l.groups {
		if groupOffset < offset {
			return ErrOffsetInUse
		}
	}

	targetFileIdx := -1
	for i, f := range l.files {
		if offset >= f.StartOffset && offset <= f.EndOffset+1 {
			targetFileIdx = i
			break
		}
	}

	if targetFileIdx < 0 {
		return ErrInvalidOffset
	}

	targetFile := l.files[targetFileIdx]
	needsRewrite := false
	keepFromOffset := offset

	if offset > targetFile.StartOffset {
		needsRewrite = true
	}

	filesToRemove := l.files[:targetFileIdx]
	l.files = l.files[targetFileIdx:]

	for _, f := range filesToRemove {
		filePath := filepath.Join(l.dir, f.FileName)
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}

	if needsRewrite {
		if err := l.rewriteFileFromOffset(targetFile, keepFromOffset); err != nil {
			return err
		}
	}

	if l.currentFile != nil {
		l.currentFile.Close()
	}
	if err := l.openLastFile(); err != nil {
		return err
	}

	return nil
}

func (l *Log) rewriteFileFromOffset(logFile *LogFile, fromOffset int64) error {
	filePath := filepath.Join(l.dir, logFile.FileName)
	tempPath := filePath + ".tmp"

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tempFile, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	header := make([]byte, headerSize)
	newStartOffset := int64(-1)
	newEndOffset := int64(-1)
	newSize := int64(0)

	for {
		n, err := file.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if n != headerSize {
			return ErrCorruptedLog
		}

		entryOffset := int64(binary.BigEndian.Uint64(header[:offsetSize]))
		length := int64(binary.BigEndian.Uint64(header[offsetSize:]))

		if entryOffset >= fromOffset {
			data := make([]byte, length)
			if _, err := io.ReadFull(file, data); err != nil {
				return err
			}

			if _, err := tempFile.Write(header); err != nil {
				return err
			}
			if _, err := tempFile.Write(data); err != nil {
				return err
			}

			if newStartOffset == -1 {
				newStartOffset = entryOffset
			}
			newEndOffset = entryOffset
			newSize += headerSize + length
		} else {
			if _, err := file.Seek(length, io.SeekCurrent); err != nil {
				return err
			}
		}
	}

	if err := tempFile.Sync(); err != nil {
		return err
	}
	tempFile.Close()
	file.Close()

	if err := os.Rename(tempPath, filePath); err != nil {
		return err
	}

	logFile.StartOffset = newStartOffset
	logFile.EndOffset = newEndOffset

	if l.currentFileNum == logFile.Number {
		l.currentSize = newSize
	}

	return nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	if l.currentFile != nil {
		if err := l.currentFile.Close(); err != nil {
			return err
		}
	}

	l.closed = true
	return nil
}

func (l *Log) CurrentOffset() int64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.currentOffset
}

func (l *Log) GroupOffset(group string) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	offset, exists := l.groups[group]
	if !exists {
		return -1, ErrGroupNotExists
	}
	return offset, nil
}
