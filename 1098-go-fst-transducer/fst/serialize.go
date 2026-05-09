package fst

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrInvalidMagic   = errors.New("invalid FST magic number")
	ErrInvalidVersion = errors.New("invalid FST version")
	ErrCorruptedData  = errors.New("corrupted FST data")
)

const (
	MagicNumber uint32 = 0x46535420
	Version     uint32 = 1
)

func writeVarUint(w io.Writer, x uint64) (int, error) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], x)
	return w.Write(buf[:n])
}

func readVarUint(r io.ByteReader) (uint64, error) {
	return binary.ReadUvarint(r)
}

func writeString(w io.Writer, s string) (int, error) {
	runes := []rune(s)
	n1, err := writeVarUint(w, uint64(len(runes)))
	if err != nil {
		return n1, err
	}
	n2 := 0
	for _, r := range runes {
		n, err := writeVarUint(w, uint64(r))
		n2 += n
		if err != nil {
			return n1 + n2, err
		}
	}
	return n1 + n2, nil
}

func readString(r io.ByteReader) (string, error) {
	lenRunes, err := readVarUint(r)
	if err != nil {
		return "", err
	}
	runes := make([]rune, lenRunes)
	for i := range runes {
		r, err := readVarUint(r)
		if err != nil {
			return "", err
		}
		runes[i] = rune(r)
	}
	return string(runes), nil
}

type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data, pos: 0}
}

func (r *byteReader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		return n, io.EOF
	}
	return n, nil
}

func (f *FST) Serialize() ([]byte, error) {
	if f.Root == nil {
		return nil, nil
	}

	buf := make([]byte, 0)

	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], MagicNumber)
	binary.LittleEndian.PutUint32(header[4:8], Version)
	buf = append(buf, header...)

	stateIDMap := make(map[*State]int)
	nextID := 0
	var assignIDs func(*State)
	assignIDs = func(s *State) {
		if _, exists := stateIDMap[s]; exists {
			return
		}
		stateIDMap[s] = nextID
		nextID++
		for _, t := range s.Transitions {
			assignIDs(t.Next)
		}
	}
	assignIDs(f.Root)

	tempBuf := make([]byte, 0, 1024)
	writer := &byteBufWriter{buf: &tempBuf}

	writeVarUint(writer, uint64(nextID))
	writeVarUint(writer, uint64(f.Count))

	visited := make(map[int]bool)
	var serializeState func(*State)
	serializeState = func(s *State) {
		id := stateIDMap[s]
		if visited[id] {
			return
		}
		visited[id] = true

		flags := uint64(0)
		if s.IsFinal {
			flags |= 1
		}
		flags |= uint64(len(s.Transitions)) << 1
		writeVarUint(writer, flags)

		if s.IsFinal {
			writeVarUint(writer, s.FinalValue)
		}

		for _, t := range s.Transitions {
			writeVarUint(writer, uint64(t.Char))
			writeVarUint(writer, uint64(stateIDMap[t.Next]))
			writeVarUint(writer, t.Value)
		}

		for _, t := range s.Transitions {
			serializeState(t.Next)
		}
	}

	serializeState(f.Root)

	writeVarUint(writer, uint64(stateIDMap[f.Root]))

	buf = append(buf, tempBuf...)
	return buf, nil
}

type byteBufWriter struct {
	buf *[]byte
}

func (w *byteBufWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func (w *byteBufWriter) WriteByte(b byte) error {
	*w.buf = append(*w.buf, b)
	return nil
}

func Deserialize(data []byte) (*FST, error) {
	if len(data) < 8 {
		return nil, ErrCorruptedData
	}

	magic := binary.LittleEndian.Uint32(data[0:4])
	if magic != MagicNumber {
		return nil, ErrInvalidMagic
	}

	version := binary.LittleEndian.Uint32(data[4:8])
	if version != Version {
		return nil, ErrInvalidVersion
	}

	reader := newByteReader(data[8:])

	numStates, err := readVarUint(reader)
	if err != nil {
		return nil, ErrCorruptedData
	}

	count, err := readVarUint(reader)
	if err != nil {
		return nil, ErrCorruptedData
	}

	states := make([]*State, numStates)
	for i := range states {
		states[i] = &State{
			ID:          i,
			Transitions: make([]Transition, 0),
		}
	}

	for i := uint64(0); i < numStates; i++ {
		flags, err := readVarUint(reader)
		if err != nil {
			return nil, ErrCorruptedData
		}

		state := states[i]
		state.IsFinal = (flags & 1) != 0
		numTrans := int(flags >> 1)

		if state.IsFinal {
			finalVal, err := readVarUint(reader)
			if err != nil {
				return nil, ErrCorruptedData
			}
			state.FinalValue = finalVal
		}

		state.Transitions = make([]Transition, 0, numTrans)
		for j := 0; j < numTrans; j++ {
			charVal, err := readVarUint(reader)
			if err != nil {
				return nil, ErrCorruptedData
			}
			char := rune(charVal)

			nextID, err := readVarUint(reader)
			if err != nil {
				return nil, ErrCorruptedData
			}

			transVal, err := readVarUint(reader)
			if err != nil {
				return nil, ErrCorruptedData
			}

			if int(nextID) >= len(states) {
				return nil, ErrCorruptedData
			}

			state.Transitions = append(state.Transitions, Transition{
				Char:  char,
				Next:  states[nextID],
				Value: transVal,
			})
		}
	}

	rootID, err := readVarUint(reader)
	if err != nil {
		return nil, ErrCorruptedData
	}

	if int(rootID) >= len(states) {
		return nil, ErrCorruptedData
	}

	fst := &FST{
		Root:   states[rootID],
		States: states,
		Count:  int(count),
	}

	return fst, nil
}
