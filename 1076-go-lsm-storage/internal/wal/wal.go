package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type WAL struct {
	path string
	f    *os.File
}

type Entry struct {
	Key       []byte
	Value     []byte
	Tombstone bool
}

func New(path string) (*WAL, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{path: path, f: f}, nil
}

func (w *WAL) Append(entry *Entry) error {
	buf := make([]byte, 0)
	buf = append(buf, u32ToBytes(uint32(len(entry.Key)))...)
	buf = append(buf, u32ToBytes(uint32(len(entry.Value)))...)
	if entry.Tombstone {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	buf = append(buf, entry.Key...)
	buf = append(buf, entry.Value...)
	_, err := w.f.Write(buf)
	return err
}

func (w *WAL) Close() error {
	if err := w.f.Sync(); err != nil {
		return err
	}
	return w.f.Close()
}

func (w *WAL) Delete() error {
	if err := w.f.Close(); err != nil {
		_ = os.Remove(w.path)
		return err
	}
	return os.Remove(w.path)
}

func (w *WAL) Path() string {
	return w.path
}

func Load(path string) ([]*Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []*Entry
	for {
		entry, err := readEntry(f)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func List(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.wal"))
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func readEntry(f *os.File) (*Entry, error) {
	header := make([]byte, 9)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, err
	}

	keyLen := bytesToU32(header[0:4])
	valLen := bytesToU32(header[4:8])
	tombstone := header[8] == 1

	body := make([]byte, int(keyLen)+int(valLen))
	if _, err := io.ReadFull(f, body); err != nil {
		return nil, err
	}

	return &Entry{
		Key:       body[:keyLen],
		Value:     body[keyLen:],
		Tombstone: tombstone,
	}, nil
}

func u32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func bytesToU32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func GenPath(dir string, seq int64) string {
	return filepath.Join(dir, fmt.Sprintf("%016d.wal", seq))
}
