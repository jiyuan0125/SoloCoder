package lsm

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	EntryTypeValue byte = 0
	EntryTypeTomb  byte = 1
)

type Entry struct {
	Key     string
	Value   string
	Deleted bool
	SeqNum  uint64
}

func (e *Entry) Size() int {
	return 1 + 8 + 4 + len(e.Key) + 4 + len(e.Value)
}

func (e *Entry) Encode() []byte {
	buf := make([]byte, e.Size())
	pos := 0
	
	if e.Deleted {
		buf[pos] = EntryTypeTomb
	} else {
		buf[pos] = EntryTypeValue
	}
	pos++
	
	binary.BigEndian.PutUint64(buf[pos:pos+8], e.SeqNum)
	pos += 8
	
	keyLen := uint32(len(e.Key))
	binary.BigEndian.PutUint32(buf[pos:pos+4], keyLen)
	pos += 4
	copy(buf[pos:pos+int(keyLen)], e.Key)
	pos += int(keyLen)
	
	valLen := uint32(len(e.Value))
	binary.BigEndian.PutUint32(buf[pos:pos+4], valLen)
	pos += 4
	copy(buf[pos:pos+int(valLen)], e.Value)
	
	return buf
}

func DecodeEntry(r io.Reader) (*Entry, error) {
	entry := &Entry{}
	
	var typeByte [1]byte
	if _, err := io.ReadFull(r, typeByte[:]); err != nil {
		return nil, err
	}
	
	entry.Deleted = typeByte[0] == EntryTypeTomb
	
	var seqBuf [8]byte
	if _, err := io.ReadFull(r, seqBuf[:]); err != nil {
		return nil, err
	}
	entry.SeqNum = binary.BigEndian.Uint64(seqBuf[:])
	
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	keyLen := binary.BigEndian.Uint32(lenBuf[:])
	keyBuf := make([]byte, keyLen)
	if _, err := io.ReadFull(r, keyBuf); err != nil {
		return nil, err
	}
	entry.Key = string(keyBuf)
	
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	valLen := binary.BigEndian.Uint32(lenBuf[:])
	valBuf := make([]byte, valLen)
	if _, err := io.ReadFull(r, valBuf); err != nil {
		return nil, err
	}
	entry.Value = string(valBuf)
	
	return entry, nil
}

type EntryList []*Entry

func (el EntryList) Len() int           { return len(el) }
func (el EntryList) Less(i, j int) bool { return el[i].Key < el[j].Key }
func (el EntryList) Swap(i, j int)      { el[i], el[j] = el[j], el[i] }

var ErrKeyExists = errors.New("key exists")
