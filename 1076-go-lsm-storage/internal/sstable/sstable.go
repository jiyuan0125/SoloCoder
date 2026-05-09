package sstable

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"lsm-storage/internal/bloom"
)

const (
	blockSize    = 32 * 1024
	footerSize   = 24
	magicNumber  = uint64(0x4c534d5353544142)
)

type BlockMeta struct {
	StartKey []byte
	Offset   uint64
}

type Footer struct {
	IndexOffset  uint64
	IndexSize    uint64
	BloomOffset  uint64
	BloomSize    uint64
	Magic        uint64
}

type Record struct {
	Key       []byte
	Value     []byte
	Tombstone bool
}

type Writer struct {
	path     string
	f        *os.File
	offset   uint64
	buffer   []Record
	bufSize  int
	blocks   []BlockMeta
	keys     [][]byte
}

func NewWriter(path string) (*Writer, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &Writer{
		path:   path,
		f:      f,
		offset: 0,
	}, nil
}

func (w *Writer) Append(key, value []byte, tombstone bool) {
	recSize := len(key) + len(value) + 9
	if w.bufSize+recSize > blockSize && len(w.buffer) > 0 {
		w.flushBlock()
	}
	w.buffer = append(w.buffer, Record{Key: key, Value: value, Tombstone: tombstone})
	w.bufSize += recSize
	w.keys = append(w.keys, key)
}

func (w *Writer) flushBlock() {
	if len(w.buffer) == 0 {
		return
	}

	var blockData []byte
	for _, rec := range w.buffer {
		blockData = appendRecord(blockData, rec)
	}

	w.blocks = append(w.blocks, BlockMeta{
		StartKey: w.buffer[0].Key,
		Offset:   w.offset,
	})

	w.f.Write(blockData)
	w.offset += uint64(len(blockData))
	w.buffer = nil
	w.bufSize = 0
}

func (w *Writer) Finish() error {
	if len(w.buffer) > 0 {
		w.flushBlock()
	}

	n := uint(len(w.keys))
	bitSize := bloom.OptimalBitSize(n, 0.01)
	k := bloom.OptimalHashCount(bitSize, n)
	bf := bloom.New(bitSize, k)
	for _, key := range w.keys {
		bf.Add(key)
	}
	bloomData := bf.Encode()

	indexData := make([]byte, 0)
	for _, meta := range w.blocks {
		indexData = appendRecord(indexData, Record{
			Key:       meta.StartKey,
			Value:     u64ToBytes(meta.Offset),
			Tombstone: false,
		})
	}

	indexOffset := w.offset
	indexSize := uint64(len(indexData))
	w.f.Write(indexData)
	w.offset += indexSize

	bloomOffset := w.offset
	bloomSize := uint64(len(bloomData))
	w.f.Write(bloomData)
	w.offset += bloomSize

	footer := Footer{
		IndexOffset: indexOffset,
		IndexSize:   indexSize,
		BloomOffset: bloomOffset,
		BloomSize:   bloomSize,
		Magic:       magicNumber,
	}

	var footerData [footerSize]byte
	binary.BigEndian.PutUint64(footerData[0:8], footer.IndexOffset)
	binary.BigEndian.PutUint64(footerData[8:16], footer.IndexSize)
	binary.BigEndian.PutUint64(footerData[16:24], magicNumber)
	w.f.Write(footerData[:])

	return w.f.Close()
}

type Reader struct {
	path     string
	f        *os.File
	footer   Footer
	index    []BlockMeta
	bf       *bloom.Filter
}

func NewReader(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	size := info.Size()
	if size < footerSize {
		f.Close()
		return nil, fmt.Errorf("invalid sstable: too small")
	}

	var footerData [footerSize]byte
	f.ReadAt(footerData[:], size-footerSize)

	footer := Footer{
		IndexOffset: binary.BigEndian.Uint64(footerData[0:8]),
		IndexSize:   binary.BigEndian.Uint64(footerData[8:16]),
		Magic:       binary.BigEndian.Uint64(footerData[16:24]),
	}

	if footer.Magic != magicNumber {
		f.Close()
		return nil, fmt.Errorf("invalid sstable: magic number mismatch")
	}

	indexData := make([]byte, footer.IndexSize)
	f.ReadAt(indexData, int64(footer.IndexOffset))
	index := parseIndex(indexData)

	r := &Reader{path: path, f: f, footer: footer, index: index}
	if err := r.loadBloom(); err != nil {
		f.Close()
		return nil, err
	}

	return r, nil
}

func (r *Reader) loadBloom() error {
	footer := r.footer
	if footer.BloomSize == 0 {
		return nil
	}
	bloomData := make([]byte, footer.BloomSize)
	_, err := r.f.ReadAt(bloomData, int64(footer.BloomOffset))
	if err != nil && err != io.EOF {
		return err
	}
	r.bf = bloom.Decode(bloomData)
	return nil
}

func (r *Reader) MayContain(key []byte) bool {
	if r.bf == nil {
		return true
	}
	return r.bf.MayContain(key)
}

func (r *Reader) Get(key []byte) ([]byte, bool, bool) {
	if !r.MayContain(key) {
		return nil, false, false
	}

	blockIdx := -1
	for i := 0; i < len(r.index); i++ {
		if compare(r.index[i].StartKey, key) <= 0 {
			blockIdx = i
		} else {
			break
		}
	}

	if blockIdx == -1 {
		return nil, false, false
	}

	blockOffset := r.index[blockIdx].Offset
	var nextOffset uint64
	if blockIdx+1 < len(r.index) {
		nextOffset = r.index[blockIdx+1].Offset
	} else {
		nextOffset = r.footer.IndexOffset
	}

	blockSize := int(nextOffset - blockOffset)
	blockData := make([]byte, blockSize)
	r.f.ReadAt(blockData, int64(blockOffset))

	records := parseRecords(blockData)
	for _, rec := range records {
		cmp := compare(rec.Key, key)
		if cmp == 0 {
			return rec.Value, rec.Tombstone, true
		} else if cmp > 0 {
			break
		}
	}

	return nil, false, false
}

func (r *Reader) Scan(start, end []byte) []Record {
	var result []Record
	for blockIdx := 0; blockIdx < len(r.index); blockIdx++ {
		meta := r.index[blockIdx]
		if end != nil && compare(meta.StartKey, end) > 0 {
			break
		}

		blockOffset := meta.Offset
		var nextOffset uint64
		if blockIdx+1 < len(r.index) {
			nextOffset = r.index[blockIdx+1].Offset
		} else {
			nextOffset = r.footer.IndexOffset
		}

		blockSize := int(nextOffset - blockOffset)
		blockData := make([]byte, blockSize)
		r.f.ReadAt(blockData, int64(blockOffset))
		records := parseRecords(blockData)

		for _, rec := range records {
			if start != nil && compare(rec.Key, start) < 0 {
				continue
			}
			if end != nil && compare(rec.Key, end) >= 0 {
				return result
			}
			result = append(result, rec)
		}
	}
	return result
}

func (r *Reader) Iterator() *Iterator {
	return &Iterator{r: r, blockIdx: 0, recIdx: 0, records: nil}
}

func (r *Reader) Close() error {
	return r.f.Close()
}

func (r *Reader) Path() string {
	return r.path
}

type Iterator struct {
	r        *Reader
	blockIdx int
	recIdx   int
	records  []Record
}

func (it *Iterator) Valid() bool {
	if it.records == nil {
		if it.blockIdx >= len(it.r.index) {
			return false
		}
		it.loadBlock()
	}
	return it.recIdx < len(it.records)
}

func (it *Iterator) loadBlock() {
	meta := it.r.index[it.blockIdx]
	blockOffset := meta.Offset
	var nextOffset uint64
	if it.blockIdx+1 < len(it.r.index) {
		nextOffset = it.r.index[it.blockIdx+1].Offset
	} else {
		nextOffset = it.r.footer.IndexOffset
	}
	blockSize := int(nextOffset - blockOffset)
	blockData := make([]byte, blockSize)
	it.r.f.ReadAt(blockData, int64(blockOffset))
	it.records = parseRecords(blockData)
	it.recIdx = 0
}

func (it *Iterator) Next() {
	it.recIdx++
	if it.recIdx >= len(it.records) {
		it.blockIdx++
		it.records = nil
	}
}

func (it *Iterator) Key() []byte {
	return it.records[it.recIdx].Key
}

func (it *Iterator) Value() []byte {
	return it.records[it.recIdx].Value
}

func (it *Iterator) Tombstone() bool {
	return it.records[it.recIdx].Tombstone
}

func (it *Iterator) Close() {
}

func appendRecord(buf []byte, rec Record) []byte {
	buf = append(buf, u32ToBytes(uint32(len(rec.Key)))...)
	buf = append(buf, u32ToBytes(uint32(len(rec.Value)))...)
	if rec.Tombstone {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	buf = append(buf, rec.Key...)
	buf = append(buf, rec.Value...)
	return buf
}

func parseRecords(data []byte) []Record {
	var records []Record
	i := 0
	for i < len(data) {
		if i+9 > len(data) {
			break
		}
		keyLen := int(binary.BigEndian.Uint32(data[i : i+4]))
		valLen := int(binary.BigEndian.Uint32(data[i+4 : i+8]))
		tombstone := data[i+8] == 1
		i += 9
		if i+keyLen+valLen > len(data) {
			break
		}
		key := append([]byte{}, data[i:i+keyLen]...)
		val := append([]byte{}, data[i+keyLen:i+keyLen+valLen]...)
		records = append(records, Record{Key: key, Value: val, Tombstone: tombstone})
		i += keyLen + valLen
	}
	return records
}

func parseIndex(data []byte) []BlockMeta {
	records := parseRecords(data)
	var metas []BlockMeta
	for _, rec := range records {
		offset := binary.BigEndian.Uint64(rec.Value)
		metas = append(metas, BlockMeta{StartKey: rec.Key, Offset: offset})
	}
	return metas
}

func compare(a, b []byte) int {
	la, lb := len(a), len(b)
	minLen := la
	if lb < minLen {
		minLen = lb
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if la < lb {
		return -1
	} else if la > lb {
		return 1
	}
	return 0
}

func u32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func u64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func GenPath(dir string, level int, seq int64) string {
	return filepath.Join(dir, fmt.Sprintf("L%d_%016d.sst", level, seq))
}

func List(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "L*_*.sst"))
	if err != nil {
		return nil, err
	}
	return matches, nil
}
