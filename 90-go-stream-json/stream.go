package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	bufferSize          = 4 * 1024
	largeFileThreshold  = 100 * 1024 * 1024
	progressInterval    = 10 * 1024 * 1024
	gzipMagic1          = 0x1f
	gzipMagic2          = 0x8b
)

type ProgressReader struct {
	reader       io.Reader
	totalBytes   int64
	readBytes    int64
	lastProgress int64
	isLargeFile  bool
}

func NewProgressReader(r io.Reader, totalSize int64) *ProgressReader {
	return &ProgressReader{
		reader:       r,
		totalBytes:   totalSize,
		readBytes:    0,
		lastProgress: 0,
		isLargeFile:  totalSize > largeFileThreshold,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.readBytes += int64(n)

	if pr.isLargeFile && pr.totalBytes > 0 {
		currentProgress := pr.readBytes / progressInterval
		if currentProgress > pr.lastProgress {
			pr.lastProgress = currentProgress
			percent := float64(pr.readBytes) / float64(pr.totalBytes) * 100
			fmt.Fprintf(os.Stderr, "Progress: %.1f%% (%d MB / %d MB)\n",
				percent, pr.readBytes/(1024*1024), pr.totalBytes/(1024*1024))
		}
	}

	return n, err
}

func isGzipFile(file *os.File) (bool, error) {
	var header [2]byte
	_, err := file.Read(header[:])
	if err != nil {
		return false, err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return false, err
	}

	return header[0] == gzipMagic1 && header[1] == gzipMagic2, nil
}

func OpenInput(filename string) (io.ReadCloser, int64, error) {
	var file *os.File
	var err error
	var totalSize int64 = -1

	if filename == "" {
		file = os.Stdin
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return nil, 0, err
		}

		info, err := file.Stat()
		if err != nil {
			file.Close()
			return nil, 0, err
		}
		totalSize = info.Size()
	}

	isGzip, err := isGzipFile(file)
	if err != nil {
		if file != os.Stdin {
			file.Close()
		}
		return nil, 0, err
	}

	var reader io.Reader = file

	if isGzip {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			if file != os.Stdin {
				file.Close()
			}
			return nil, 0, err
		}
		reader = gzReader
	}

	bufReader := bufio.NewReaderSize(reader, bufferSize)

	if totalSize > 0 {
		progressReader := NewProgressReader(bufReader, totalSize)
		return &wrappedCloser{reader: progressReader, closer: file, isGzip: isGzip}, totalSize, nil
	}

	return &wrappedCloser{reader: bufReader, closer: file, isGzip: isGzip}, totalSize, nil
}

type wrappedCloser struct {
	reader io.Reader
	closer io.Closer
	isGzip bool
	gz     *gzip.Reader
}

func (wc *wrappedCloser) Read(p []byte) (int, error) {
	return wc.reader.Read(p)
}

func (wc *wrappedCloser) Close() error {
	if wc.closer != nil && wc.closer != os.Stdin {
		return wc.closer.Close()
	}
	return nil
}

type MatchHandler interface {
	HandleMatch(value interface{})
}

type QueryHandler struct {
	Results []interface{}
}

func (qh *QueryHandler) HandleMatch(value interface{}) {
	qh.Results = append(qh.Results, value)
}

type pathElement struct {
	kind  elementKind
	field string
	index int
}

type elementKind int

const (
	elemRoot elementKind = iota
	elemField
	elemIndex
)

type StreamProcessor struct {
	decoder *json.Decoder
	jp      *JSONPath
	handler MatchHandler
}

func NewStreamProcessor(r io.Reader, jp *JSONPath, handler MatchHandler) *StreamProcessor {
	return &StreamProcessor{
		decoder: json.NewDecoder(r),
		jp:      jp,
		handler: handler,
	}
}

func (sp *StreamProcessor) Process() error {
	var path []pathElement
	path = append(path, pathElement{kind: elemRoot})

	for {
		tok, err := sp.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := sp.processRootToken(tok, &path); err != nil {
			return err
		}
	}
	return nil
}

func (sp *StreamProcessor) processRootToken(tok json.Token, path *[]pathElement) error {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return sp.processObject(path)
		case '[':
			return sp.processArray(path)
		}
	default:
		if sp.matchesPath(*path) {
			sp.handleValue(t)
		}
	}
	return nil
}

func (sp *StreamProcessor) processObject(path *[]pathElement) error {
	for sp.decoder.More() {
		keyTok, err := sp.decoder.Token()
		if err != nil {
			return err
		}

		key, ok := keyTok.(string)
		if !ok {
			continue
		}

		*path = append(*path, pathElement{kind: elemField, field: key})

		valTok, err := sp.decoder.Token()
		if err != nil {
			return err
		}

		if err := sp.processValue(valTok, path); err != nil {
			return err
		}

		if len(*path) > 1 {
			*path = (*path)[:len(*path)-1]
		}
	}

	_, err := sp.decoder.Token()
	return err
}

func (sp *StreamProcessor) processArray(path *[]pathElement) error {
	index := 0
	for sp.decoder.More() {
		*path = append(*path, pathElement{kind: elemIndex, index: index})

		valTok, err := sp.decoder.Token()
		if err != nil {
			return err
		}

		if err := sp.processValue(valTok, path); err != nil {
			return err
		}

		if len(*path) > 1 {
			*path = (*path)[:len(*path)-1]
		}
		index++
	}

	_, err := sp.decoder.Token()
	return err
}

func (sp *StreamProcessor) processValue(tok json.Token, path *[]pathElement) error {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return sp.processObject(path)
		case '[':
			return sp.processArray(path)
		}
	default:
		if sp.matchesPath(*path) {
			sp.handleValue(tok)
		}
	}
	return nil
}

func (sp *StreamProcessor) handleValue(tok json.Token) {
	switch t := tok.(type) {
	case string:
		sp.handler.HandleMatch(t)
	case json.Number:
		if f, err := t.Float64(); err == nil {
			sp.handler.HandleMatch(f)
		} else {
			sp.handler.HandleMatch(t.String())
		}
	case bool:
		sp.handler.HandleMatch(t)
	case nil:
		sp.handler.HandleMatch(nil)
	}
}

func (sp *StreamProcessor) matchesPath(currentPath []pathElement) bool {
	if len(sp.jp.Segments) == 1 {
		return true
	}

	if len(currentPath) != len(sp.jp.Segments) {
		return false
	}

	for i := 1; i < len(sp.jp.Segments); i++ {
		seg := sp.jp.Segments[i]
		elem := currentPath[i]

		switch seg.Type {
		case SegField:
			if elem.kind != elemField || elem.field != seg.Value {
				return false
			}
		case SegIndex:
			if elem.kind != elemIndex || elem.index != seg.Index {
				return false
			}
		case SegWildcard:
			if elem.kind != elemIndex {
				return false
			}
		}
	}

	return true
}

type KeysHandler struct {
	keys map[string]bool
}

func NewKeysHandler() *KeysHandler {
	return &KeysHandler{
		keys: make(map[string]bool),
	}
}

func (kh *KeysHandler) HandleMatch(value interface{}) {
}

func (kh *KeysHandler) AddKey(key string) {
	kh.keys[key] = true
}

func (kh *KeysHandler) GetKeys() []string {
	keys := make([]string, 0, len(kh.keys))
	for k := range kh.keys {
		keys = append(keys, k)
	}
	return keys
}

type KeysProcessor struct {
	decoder *json.Decoder
	handler *KeysHandler
}

func NewKeysProcessor(r io.Reader, handler *KeysHandler) *KeysProcessor {
	return &KeysProcessor{
		decoder: json.NewDecoder(r),
		handler: handler,
	}
}

func (kp *KeysProcessor) Process() error {
	for {
		tok, err := kp.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := kp.processRootToken(tok); err != nil {
			return err
		}
	}
	return nil
}

func (kp *KeysProcessor) processRootToken(tok json.Token) error {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return kp.processObject()
		case '[':
			return kp.processArray()
		}
	}
	return nil
}

func (kp *KeysProcessor) processObject() error {
	for kp.decoder.More() {
		keyTok, err := kp.decoder.Token()
		if err != nil {
			return err
		}

		if key, ok := keyTok.(string); ok {
			kp.handler.AddKey(key)
		}

		valTok, err := kp.decoder.Token()
		if err != nil {
			return err
		}

		if err := kp.processValue(valTok); err != nil {
			return err
		}
	}

	_, err := kp.decoder.Token()
	return err
}

func (kp *KeysProcessor) processArray() error {
	for kp.decoder.More() {
		valTok, err := kp.decoder.Token()
		if err != nil {
			return err
		}

		if err := kp.processValue(valTok); err != nil {
			return err
		}
	}

	_, err := kp.decoder.Token()
	return err
}

func (kp *KeysProcessor) processValue(tok json.Token) error {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return kp.processObject()
		case '[':
			return kp.processArray()
		}
	}
	return nil
}
