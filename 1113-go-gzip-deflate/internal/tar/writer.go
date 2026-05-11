package tar

import (
	"fmt"
	"io"
	"strings"
)

const (
	BlockSize = 512

	TypeReg  = '0'
	TypeRegA = '\x00'
	TypeLink = '1'
	TypeSymlink = '2'
	TypeChar = '3'
	TypeBlock = '4'
	TypeDir = '5'
	TypeFifo = '6'
	TypeCont = '7'
	TypeGNULongName = 'L'
	TypeGNULongLink = 'K'
)

type Header struct {
	Name     string
	LinkName string
	Size     int64
	Mode     int64
	Uid      int
	Gid      int
	Uname    string
	Gname    string
	ModTime  int64
	Typeflag byte
	Devmajor int64
	Devminor int64
}

type Writer struct {
	dst          io.Writer
	currentFile  *currentFileInfo
	fileCount    int
	closed       bool
}

type currentFileInfo struct {
	size       int64
	written    int64
	header     *Header
}

func NewWriter(dst io.Writer) *Writer {
	return &Writer{
		dst: dst,
	}
}

func (w *Writer) WriteHeader(hdr *Header) error {
	if w.closed {
		return fmt.Errorf("tar writer is closed")
	}

	if w.currentFile != nil {
		return fmt.Errorf("cannot write new header while previous file is incomplete")
	}

	hdrCopy := *hdr

	if hdrCopy.Typeflag == 0 {
		if hdrCopy.LinkName != "" {
			hdrCopy.Typeflag = TypeSymlink
		} else if strings.HasSuffix(hdrCopy.Name, "/") {
			hdrCopy.Typeflag = TypeDir
		} else {
			hdrCopy.Typeflag = TypeReg
		}
	}

	if hdrCopy.Typeflag == TypeDir && !strings.HasSuffix(hdrCopy.Name, "/") {
		hdrCopy.Name += "/"
	}

	if hdrCopy.Typeflag == TypeDir {
		hdrCopy.Size = 0
	}

	if len(hdrCopy.Name) > 100 || len(hdrCopy.LinkName) > 100 {
		longName := hdrCopy.Name
		if len(longName) > 100 {
			longNameHdr := &Header{
				Name:     "././@LongLink",
				Size:     int64(len(longName) + 1),
				Typeflag: TypeGNULongName,
				Mode:     0,
				Uid:      0,
				Gid:      0,
				ModTime:  0,
			}
			if err := w.writeRawHeader(longNameHdr); err != nil {
				return err
			}
			data := []byte(longName)
			data = append(data, 0)
			w.dst.Write(data)
			padding := BlockSize - (len(data) % BlockSize)
			if padding != BlockSize {
				w.dst.Write(make([]byte, padding))
			}
		}

		longLink := hdrCopy.LinkName
		if len(longLink) > 100 {
			longLinkHdr := &Header{
				Name:     "././@LongLink",
				Size:     int64(len(longLink) + 1),
				Typeflag: TypeGNULongLink,
				Mode:     0,
				Uid:      0,
				Gid:      0,
				ModTime:  0,
			}
			if err := w.writeRawHeader(longLinkHdr); err != nil {
				return err
			}
			data := []byte(longLink)
			data = append(data, 0)
			w.dst.Write(data)
			padding := BlockSize - (len(data) % BlockSize)
			if padding != BlockSize {
				w.dst.Write(make([]byte, padding))
			}
		}
	}

	if err := w.writeRawHeader(&hdrCopy); err != nil {
		return err
	}

	w.currentFile = &currentFileInfo{
		size:   hdrCopy.Size,
		header: &hdrCopy,
	}
	w.fileCount++

	return nil
}

func (w *Writer) writeRawHeader(hdr *Header) error {
	var buf [BlockSize]byte

	writeString(buf[0:100], hdr.Name)
	writeNumeric(buf[100:108], hdr.Mode, 8)
	writeNumeric(buf[108:116], int64(hdr.Uid), 8)
	writeNumeric(buf[116:124], int64(hdr.Gid), 8)
	writeNumeric(buf[124:136], hdr.Size, 12)
	writeNumeric(buf[136:148], hdr.ModTime, 12)

	for i := 148; i < 156; i++ {
		buf[i] = ' '
	}

	buf[156] = hdr.Typeflag
	writeString(buf[157:257], hdr.LinkName)

	copy(buf[257:263], []byte("ustar\x00"))
	copy(buf[263:265], []byte("00"))

	writeString(buf[265:297], hdr.Uname)
	writeString(buf[297:329], hdr.Gname)
	writeNumeric(buf[329:337], hdr.Devmajor, 8)
	writeNumeric(buf[337:345], hdr.Devminor, 8)

	if len(hdr.Name) > 100 && len(hdr.Name) <= 256 {
		prefix := hdr.Name
		for idx := len(prefix) - 100; idx >= 0; idx-- {
			if prefix[idx] == '/' {
				prefix = hdr.Name[:idx]
				name := hdr.Name[idx+1:]
				if len(name) <= 100 && len(prefix) <= 155 {
					writeString(buf[0:100], name)
					writeString(buf[345:500], prefix)
					break
				}
			}
		}
	}

	checksum := computeChecksum(&buf)
	writeChecksum(buf[148:156], checksum)

	_, err := w.dst.Write(buf[:])
	return err
}

func writeString(buf []byte, s string) {
	for i := 0; i < len(buf); i++ {
		if i < len(s) {
			buf[i] = s[i]
		} else {
			buf[i] = 0
		}
	}
}

func writeNumeric(buf []byte, n int64, size int) {
	if n < 0 {
		n = 0
	}

	octal := fmt.Sprintf("%0*o", size-1, n)
	if len(octal) > size-1 {
		octal = octal[len(octal)-(size-1):]
	}

	for i := 0; i < len(octal); i++ {
		if i < len(buf) {
			buf[i] = octal[i]
		}
	}

	for i := len(octal); i < len(buf)-1; i++ {
		buf[i] = '0'
	}
	if len(buf) > 0 {
		buf[len(buf)-1] = 0
	}
}

func writeChecksum(buf []byte, sum int) {
	octal := fmt.Sprintf("%06o", sum)
	for i := 0; i < 6 && i < len(octal); i++ {
		buf[i] = octal[i]
	}
	buf[6] = 0
	buf[7] = ' '
}

func computeChecksum(buf *[BlockSize]byte) int {
	var sum int
	for i := 0; i < BlockSize; i++ {
		sum += int(buf[i])
	}
	return sum
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, fmt.Errorf("tar writer is closed")
	}
	if w.currentFile == nil {
		return 0, fmt.Errorf("no file header written")
	}

	remaining := w.currentFile.size - w.currentFile.written
	if remaining < int64(len(p)) {
		p = p[:remaining]
	}

	n, err = w.dst.Write(p)
	w.currentFile.written += int64(n)
	return n, err
}

func (w *Writer) Flush() error {
	if w.closed {
		return fmt.Errorf("tar writer is closed")
	}

	if w.currentFile != nil {
		if w.currentFile.written < w.currentFile.size {
			return fmt.Errorf("file not fully written")
		}

		remainder := w.currentFile.written % BlockSize
		if remainder > 0 {
			padding := BlockSize - remainder
			if _, err := w.dst.Write(make([]byte, padding)); err != nil {
				return err
			}
		}
		w.currentFile = nil
	}

	return nil
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}

	if w.currentFile != nil {
		if err := w.Flush(); err != nil {
			return err
		}
	}

	trailer := make([]byte, BlockSize*2)
	if _, err := w.dst.Write(trailer); err != nil {
		return err
	}

	w.closed = true
	return nil
}

func (w *Writer) FileCount() int {
	return w.fileCount
}

func NewHeaderFromInfo(name string, size int64, mode int64, modTime int64, isDir bool, linkName string) *Header {
	hdr := &Header{
		Name:    name,
		Size:    size,
		Mode:    mode,
		ModTime: modTime,
	}

	if isDir {
		hdr.Typeflag = TypeDir
		if !strings.HasSuffix(hdr.Name, "/") {
			hdr.Name += "/"
		}
		hdr.Size = 0
	} else if linkName != "" {
		hdr.Typeflag = TypeSymlink
		hdr.LinkName = linkName
		hdr.Size = 0
	} else {
		hdr.Typeflag = TypeReg
	}

	if hdr.Mode == 0 {
		if hdr.Typeflag == TypeDir {
			hdr.Mode = 0755
		} else {
			hdr.Mode = 0644
		}
	}

	return hdr
}
