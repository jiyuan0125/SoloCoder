package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"time"
)

type LogEntry struct {
	Timestamp  time.Time
	Lines      []string
	FileIndex  int
	LineNumber int64
}

type ProgressUpdate struct {
	Lines    int64
	Progress float64
}

type LogFileReader struct {
	filename      string
	file          *os.File
	reader        *bufio.Reader
	currentEntry  *LogEntry
	nextLine      string
	hasNextLine   bool
	eof           bool
	lineNumber    int64
	bytesConsumed int64
	fileIndex     int
	timeFormat    string
}

type LogEntryHeap []*LogEntry

func (h LogEntryHeap) Len() int           { return len(h) }
func (h LogEntryHeap) Less(i, j int) bool {
	if h[i].Timestamp.Equal(h[j].Timestamp) {
		return h[i].FileIndex < h[j].FileIndex
	}
	return h[i].Timestamp.Before(h[j].Timestamp)
}
func (h LogEntryHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *LogEntryHeap) Push(x interface{}) {
	*h = append(*h, x.(*LogEntry))
}

func (h *LogEntryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type LogMerger struct {
	readers       []*LogFileReader
	outputFile    *os.File
	writer        *bufio.Writer
	outputPath    string
	timeFormat    string
	totalLines    int64
	outputWritten int64
}

func NewLogMerger(inputFiles []string, outputFile string, timeFormat string, isResume bool) (*LogMerger, error) {
	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05"
	}

	var outFile *os.File
	var outputWritten int64
	var err error

	if isResume {
		if info, statErr := os.Stat(outputFile); statErr == nil {
			outFile, err = os.OpenFile(outputFile, os.O_RDWR, 0644)
			if err != nil {
				return nil, fmt.Errorf("打开输出文件失败: %v", err)
			}
			outputWritten = info.Size()
			_, err = outFile.Seek(0, io.SeekEnd)
			if err != nil {
				outFile.Close()
				return nil, fmt.Errorf("定位输出文件末尾失败: %v", err)
			}
		} else {
			outFile, err = os.Create(outputFile)
			if err != nil {
				return nil, fmt.Errorf("创建输出文件失败: %v", err)
			}
			outputWritten = 0
		}
	} else {
		outFile, err = os.Create(outputFile)
		if err != nil {
			return nil, fmt.Errorf("创建输出文件失败: %v", err)
		}
		outputWritten = 0
	}

	merger := &LogMerger{
		outputFile:    outFile,
		writer:        bufio.NewWriter(outFile),
		outputPath:    outputFile,
		timeFormat:    timeFormat,
		outputWritten: outputWritten,
	}

	for i, filename := range inputFiles {
		reader, err := NewLogFileReader(filename, i, timeFormat)
		if err != nil {
			merger.Close()
			return nil, err
		}
		merger.readers = append(merger.readers, reader)
	}

	return merger, nil
}

func NewLogFileReader(filename string, fileIndex int, timeFormat string) (*LogFileReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %s, %v", filename, err)
	}

	reader := &LogFileReader{
		filename:      filename,
		file:          file,
		reader:        bufio.NewReader(file),
		fileIndex:     fileIndex,
		timeFormat:    timeFormat,
		bytesConsumed: 0,
	}

	return reader, nil
}

func (r *LogFileReader) Close() {
	if r.file != nil {
		r.file.Close()
	}
}

func (r *LogFileReader) HasEntry() bool {
	if r.currentEntry != nil {
		return true
	}
	entry, err := r.readEntry()
	if err != nil || entry == nil {
		return false
	}
	r.currentEntry = entry
	return true
}

func (r *LogFileReader) PeekEntry() *LogEntry {
	if r.currentEntry != nil {
		return r.currentEntry
	}
	entry, _ := r.readEntry()
	r.currentEntry = entry
	return entry
}

func (r *LogFileReader) NextEntry() *LogEntry {
	entry := r.currentEntry
	r.currentEntry = nil
	return entry
}

func (r *LogFileReader) readLine() (string, int, error) {
	var line []byte
	var totalBytes int

	for {
		partial, isPrefix, err := r.reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				if len(line) > 0 {
					return string(line), totalBytes, nil
				}
				return "", 0, io.EOF
			}
			return "", 0, err
		}

		line = append(line, partial...)
		totalBytes += len(partial)

		if !isPrefix {
			totalBytes += 1
			return string(line), totalBytes, nil
		}
	}
}

func (r *LogFileReader) readEntry() (*LogEntry, error) {
	var entry *LogEntry

	for {
		var line string
		var bytesRead int
		var err error

		if r.hasNextLine {
			line = r.nextLine
			r.hasNextLine = false
			bytesRead = len(line) + 1
		} else {
			line, bytesRead, err = r.readLine()
			if err != nil {
				if err == io.EOF {
					if entry != nil {
						return entry, nil
					}
					r.eof = true
					return nil, nil
				}
				return nil, err
			}
			r.lineNumber++
		}

		timestamp, err := r.parseTimestamp(line)
		if err == nil {
			if entry != nil {
				r.nextLine = line
				r.hasNextLine = true
				return entry, nil
			}
			entry = &LogEntry{
				Timestamp:  timestamp,
				Lines:      []string{line},
				FileIndex:  r.fileIndex,
				LineNumber: r.lineNumber,
			}
			r.bytesConsumed += int64(bytesRead)
		} else {
			if entry != nil {
				entry.Lines = append(entry.Lines, line)
				r.bytesConsumed += int64(bytesRead)
			}
		}
	}
}

func (r *LogFileReader) parseTimestamp(line string) (time.Time, error) {
	if len(line) < len(r.timeFormat) {
		return time.Time{}, fmt.Errorf("行太短")
	}

	tsStr := line[:len(r.timeFormat)]
	ts, err := time.ParseInLocation(r.timeFormat, tsStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *LogFileReader) GetPosition() int64 {
	return r.bytesConsumed
}

func (r *LogFileReader) SeekTo(offset int64) error {
	_, err := r.file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	r.reader = bufio.NewReader(r.file)
	r.currentEntry = nil
	r.hasNextLine = false
	r.eof = false
	r.lineNumber = 0
	r.bytesConsumed = offset
	return nil
}

func (m *LogMerger) Close() {
	for _, r := range m.readers {
		r.Close()
	}
	if m.writer != nil {
		m.writer.Flush()
	}
	if m.outputFile != nil {
		m.outputFile.Close()
	}
}

func (m *LogMerger) Merge(progressCh chan<- ProgressUpdate, stopCh <-chan struct{}) error {
	h := &LogEntryHeap{}
	heap.Init(h)

	for _, reader := range m.readers {
		if reader.HasEntry() {
			heap.Push(h, reader.PeekEntry())
		}
	}

	var totalLines int64
	var processedLines int64

	for h.Len() > 0 {
		select {
		case <-stopCh:
			return nil
		default:
		}

		entry := heap.Pop(h).(*LogEntry)

		for _, line := range entry.Lines {
			lineWithNewline := line + "\n"
			if _, err := m.writer.WriteString(lineWithNewline); err != nil {
				return fmt.Errorf("写入输出失败: %v", err)
			}
			processedLines++
			m.outputWritten += int64(len(lineWithNewline))
		}

		m.writer.Flush()
		totalLines++
		m.totalLines = processedLines

		reader := m.readers[entry.FileIndex]
		reader.NextEntry()
		if reader.HasEntry() {
			heap.Push(h, reader.PeekEntry())
		}

		if progressCh != nil && totalLines%100 == 0 {
			progress := m.calculateProgress()
			progressCh <- ProgressUpdate{
				Lines:    processedLines,
				Progress: progress,
			}
		}
	}

	if progressCh != nil {
		progressCh <- ProgressUpdate{
			Lines:    processedLines,
			Progress: 100.0,
		}
	}

	return nil
}

func (m *LogMerger) calculateProgress() float64 {
	var totalSize int64
	var processedSize int64

	for _, reader := range m.readers {
		info, err := os.Stat(reader.filename)
		if err != nil {
			continue
		}
		totalSize += info.Size()
		processedSize += reader.GetPosition()
	}

	if totalSize == 0 {
		return 0.0
	}

	return float64(processedSize) / float64(totalSize) * 100.0
}

func (m *LogMerger) TotalLines() int64 {
	return m.totalLines
}

func (m *LogMerger) OutputPosition() int64 {
	return m.outputWritten
}

func (m *LogMerger) CreateCheckpoint() *Checkpoint {
	ckpt := &Checkpoint{
		TimeFormat:   m.timeFormat,
		OutputFile:   m.outputPath,
		OutputPos:  m.outputWritten,
		Files:      make([]FileCheckpoint, len(m.readers)),
	}

	for i, reader := range m.readers {
		ckpt.Files[i] = FileCheckpoint{
			Filename: reader.filename,
			Position: reader.GetPosition(),
		}
	}

	return ckpt
}

func (m *LogMerger) RestoreCheckpoint(ckpt *Checkpoint) {
	if ckpt.TimeFormat != "" {
		m.timeFormat = ckpt.TimeFormat
	}

	for _, fileCkpt := range ckpt.Files {
		for _, reader := range m.readers {
			if reader.filename == fileCkpt.Filename {
				reader.SeekTo(fileCkpt.Position)
				break
			}
		}
	}
}
