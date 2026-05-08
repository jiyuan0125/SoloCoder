package commitlog

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	segmentExt    = ".log"
	indexExt      = ".index"
	indexEntrySize = 16
)

type segment struct {
	baseOffset  int64
	logFile     *os.File
	indexFile   *os.File
	logOffset   int64
	indexOffset int64
	position    int64
	lastWrite   time.Time
	mu          sync.RWMutex
	closed      bool
}

func segmentFileName(offset int64) string {
	return fmt.Sprintf("%020d%s", offset, segmentExt)
}

func indexFileName(offset int64) string {
	return fmt.Sprintf("%020d%s", offset, indexExt)
}

func parseSegmentOffset(name string) (int64, error) {
	base := strings.TrimSuffix(name, segmentExt)
	var offset int64
	_, err := fmt.Sscanf(base, "%020d", &offset)
	return offset, err
}

func openSegments(dir string) ([]*segment, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"+segmentExt))
	if err != nil {
		return nil, err
	}

	var segments []*segment
	for _, file := range files {
		name := filepath.Base(file)
		offset, err := parseSegmentOffset(name)
		if err != nil {
			continue
		}

		seg, err := openSegment(dir, offset)
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].baseOffset < segments[j].baseOffset
	})

	return segments, nil
}

func openSegment(dir string, baseOffset int64) (*segment, error) {
	logPath := filepath.Join(dir, segmentFileName(baseOffset))
	indexPath := filepath.Join(dir, indexFileName(baseOffset))

	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(indexPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logFile.Close()
		return nil, err
	}

	logStat, err := logFile.Stat()
	if err != nil {
		logFile.Close()
		indexFile.Close()
		return nil, err
	}

	indexStat, err := indexFile.Stat()
	if err != nil {
		logFile.Close()
		indexFile.Close()
		return nil, err
	}

	seg := &segment{
		baseOffset:  baseOffset,
		logFile:     logFile,
		indexFile:   indexFile,
		position:    logStat.Size(),
		logOffset:   logStat.Size(),
		indexOffset: indexStat.Size(),
		lastWrite:   logStat.ModTime(),
	}

	return seg, nil
}

func createSegment(dir string, baseOffset int64) (*segment, error) {
	seg, err := openSegment(dir, baseOffset)
	if err != nil {
		return nil, err
	}
	return seg, nil
}

func (s *segment) append(rec *record, indexInterval int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return 0, fmt.Errorf("segment closed")
	}

	encoded := rec.encode()
	oldPos := s.position

	if _, err := s.logFile.WriteAt(encoded, s.position); err != nil {
		return 0, err
	}

	s.position += int64(len(encoded))
	s.logOffset += int64(len(encoded))

	if s.position-oldPos >= int64(indexInterval) || oldPos == 0 {
		if err := s.writeIndexEntry(rec.offset, oldPos); err != nil {
			return 0, err
		}
	}

	s.lastWrite = time.Now()
	return rec.offset, nil
}

func (s *segment) writeIndexEntry(offset int64, pos int64) error {
	buf := make([]byte, indexEntrySize)
	binary.BigEndian.PutUint64(buf[0:8], uint64(offset))
	binary.BigEndian.PutUint64(buf[8:16], uint64(pos))

	if _, err := s.indexFile.WriteAt(buf, s.indexOffset); err != nil {
		return err
	}
	s.indexOffset += indexEntrySize
	return nil
}

func (s *segment) size() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.position
}

func (s *segment) canFit(rec *record, segmentSize int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.position+int64(rec.size()) <= segmentSize
}

func (s *segment) readAt(offset int64) (*record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("segment closed")
	}

	pos, err := s.lookup(offset)
	if err != nil {
		return nil, err
	}

	headerBuf := make([]byte, messageHeaderSize)
	if _, err := s.logFile.ReadAt(headerBuf, pos); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(headerBuf[0:4])
	if length == 0 || length > maxMessageSize {
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	fullBuf := make([]byte, messageHeaderSize+int(length))
	copy(fullBuf, headerBuf)
	if length > 0 {
		if _, err := s.logFile.ReadAt(fullBuf[messageHeaderSize:], pos+messageHeaderSize); err != nil && err != io.EOF {
			return nil, err
		}
	}

	rec := &record{}
	if err := rec.decode(fullBuf); err != nil {
		return nil, err
	}

	return rec, nil
}

func (s *segment) lookup(offset int64) (int64, error) {
	if s.indexOffset == 0 {
		return 0, nil
	}

	numEntries := s.indexOffset / indexEntrySize
	buf := make([]byte, indexEntrySize)

	var bestPos int64 = 0
	for i := int64(0); i < numEntries; i++ {
		if _, err := s.indexFile.ReadAt(buf, i*indexEntrySize); err != nil {
			return 0, err
		}
		entryOffset := int64(binary.BigEndian.Uint64(buf[0:8]))
		if entryOffset <= offset {
			bestPos = int64(binary.BigEndian.Uint64(buf[8:16]))
		} else {
			break
		}
	}

	return bestPos, nil
}

func (s *segment) findPosition(offset int64) (int64, error) {
	return s.lookup(offset)
}

func (s *segment) scanFrom(pos int64, maxMessages int, callback func(*record) bool) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return fmt.Errorf("segment closed")
	}

	headerBuf := make([]byte, messageHeaderSize)
	currentPos := pos

	for i := 0; i < maxMessages; i++ {
		if currentPos >= s.position {
			break
		}

		if _, err := s.logFile.ReadAt(headerBuf, currentPos); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		length := binary.BigEndian.Uint32(headerBuf[0:4])
		if length == 0 || length > maxMessageSize {
			currentPos += messageHeaderSize
			continue
		}

		fullBuf := make([]byte, messageHeaderSize+int(length))
		copy(fullBuf, headerBuf)
		if length > 0 {
			if _, err := s.logFile.ReadAt(fullBuf[messageHeaderSize:], currentPos+messageHeaderSize); err != nil && err != io.EOF {
				currentPos += messageHeaderSize + int64(length)
				continue
			}
		}

		rec := &record{}
		if err := rec.decode(fullBuf); err != nil {
			currentPos += messageHeaderSize + int64(length)
			continue
		}

		if !callback(rec) {
			break
		}

		currentPos += int64(messageHeaderSize + int(length))
	}

	return nil
}

func (s *segment) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	if s.logFile != nil {
		s.logFile.Sync()
		s.logFile.Close()
	}
	if s.indexFile != nil {
		s.indexFile.Sync()
		s.indexFile.Close()
	}
	return nil
}

func (s *segment) sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.logFile != nil {
		if err := s.logFile.Sync(); err != nil {
			return err
		}
	}
	if s.indexFile != nil {
		if err := s.indexFile.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (s *segment) delete(dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed {
		s.closed = true
		if s.logFile != nil {
			s.logFile.Close()
		}
		if s.indexFile != nil {
			s.indexFile.Close()
		}
	}

	logPath := filepath.Join(dir, segmentFileName(s.baseOffset))
	indexPath := filepath.Join(dir, indexFileName(s.baseOffset))

	if err := os.Remove(logPath); err != nil {
		return err
	}
	if err := os.Remove(indexPath); err != nil {
		return err
	}

	return nil
}
