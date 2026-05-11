package gzip

import (
	"encoding/binary"
	"io"
	"time"

	"archiver/internal/deflate"
)

const (
	FlagText    = 0x01
	FlagHeaderCRC = 0x02
	FlagExtra   = 0x04
	FlagName    = 0x08
	FlagComment = 0x10
	FlagReserved = 0xE0

	CompressionMethodDeflate = 8

	OSUnknown   = 255
	OSUnix      = 3
)

type Header struct {
	Comment string
	Extra   []byte
	ModTime time.Time
	Name    string
	OS      byte
}

type Writer struct {
	dst         io.Writer
	deflateWriter *deflate.Writer
	header      *Header
	headerWritten bool
	closed      bool
	crc32       uint32
	size        uint32
}

var crc32Table []uint32

func init() {
	crc32Table = make([]uint32, 256)
	for i := range crc32Table {
		crc := uint32(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
		crc32Table[i] = crc
	}
}

func NewWriter(dst io.Writer) *Writer {
	return NewWriterHeader(dst, &Header{
		OS: OSUnknown,
	})
}

func NewWriterHeader(dst io.Writer, h *Header) *Writer {
	if h == nil {
		h = &Header{
			OS: OSUnknown,
		}
	}
	return &Writer{
		dst:         dst,
		deflateWriter: deflate.NewWriter(dst),
		header:      h,
		crc32:       0xFFFFFFFF,
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}

	if !w.headerWritten {
		if err := w.writeHeader(); err != nil {
			return 0, err
		}
		w.headerWritten = true
	}

	w.updateCRC32(p)
	w.size += uint32(len(p))

	return w.deflateWriter.Write(p)
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}

	if !w.headerWritten {
		if err := w.writeHeader(); err != nil {
			return err
		}
		w.headerWritten = true
	}

	if err := w.deflateWriter.Close(); err != nil {
		return err
	}

	w.crc32 = ^w.crc32

	if err := w.writeTrailer(); err != nil {
		return err
	}

	w.closed = true
	return nil
}

func (w *Writer) writeHeader() error {
	flags := byte(0)
	if w.header.Name != "" {
		flags |= FlagName
	}
	if w.header.Comment != "" {
		flags |= FlagComment
	}
	if len(w.header.Extra) > 0 {
		flags |= FlagExtra
	}

	var modTime uint32
	if !w.header.ModTime.IsZero() {
		modTime = uint32(w.header.ModTime.Unix())
	}

	header := []byte{
		0x1f, 0x8b,
		CompressionMethodDeflate,
		flags,
		byte(modTime), byte(modTime >> 8), byte(modTime >> 16), byte(modTime >> 24),
		0,
		w.header.OS,
	}

	if _, err := w.dst.Write(header); err != nil {
		return err
	}

	if len(w.header.Extra) > 0 {
		extraLen := len(w.header.Extra)
		w.dst.Write([]byte{byte(extraLen), byte(extraLen >> 8)})
		w.dst.Write(w.header.Extra)
	}

	if w.header.Name != "" {
		w.dst.Write([]byte(w.header.Name))
		w.dst.Write([]byte{0})
	}

	if w.header.Comment != "" {
		w.dst.Write([]byte(w.header.Comment))
		w.dst.Write([]byte{0})
	}

	return nil
}

func (w *Writer) writeTrailer() error {
	trailer := make([]byte, 8)
	binary.LittleEndian.PutUint32(trailer[0:4], w.crc32)
	binary.LittleEndian.PutUint32(trailer[4:8], w.size)

	_, err := w.dst.Write(trailer)
	return err
}

func (w *Writer) updateCRC32(data []byte) {
	for _, b := range data {
		w.crc32 = crc32Table[(w.crc32^uint32(b))&0xFF] ^ (w.crc32 >> 8)
	}
}
