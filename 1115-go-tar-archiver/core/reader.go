package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Reader struct {
	r      io.Reader
	currHdr *Header
	currSize int64
	currRead int64
	eof    bool
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

func (tr *Reader) Next() (*Header, error) {
	if tr.currSize > 0 {
		if tr.currSize > tr.currRead {
			remaining := tr.currSize - tr.currRead
			if _, err := io.CopyN(io.Discard, tr.r, remaining); err != nil {
				return nil, err
			}
			tr.currRead = tr.currSize
		}
		padded := padToBlock(tr.currSize)
		padding := padded - tr.currSize
		if padding > 0 {
			if _, err := io.CopyN(io.Discard, tr.r, padding); err != nil {
				return nil, err
			}
		}
	}
	tr.currHdr = nil
	tr.currSize = 0
	tr.currRead = 0
	block := make([]byte, blockSize)
	for {
		if _, err := io.ReadFull(tr.r, block); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil, io.EOF
			}
			return nil, err
		}
		if isZeroBlock(block) {
			if _, err := io.ReadFull(tr.r, block); err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return nil, io.EOF
				}
				return nil, err
			}
			if isZeroBlock(block) {
				tr.eof = true
				return nil, io.EOF
			}
			continue
		}
		break
	}
	hdr, err := tr.parseHeader(block)
	if err != nil {
		return nil, err
	}
	if hdr.Typeflag == TypeGNULongName {
		name, err := readAllString(tr.r, hdr.Size)
		if err != nil {
			return nil, err
		}
		padded := padToBlock(hdr.Size)
		if padded > hdr.Size {
			io.CopyN(io.Discard, tr.r, padded-hdr.Size)
		}
		nextBlock := make([]byte, blockSize)
		if _, err := io.ReadFull(tr.r, nextBlock); err != nil {
			return nil, err
		}
		realHdr, err := tr.parseHeader(nextBlock)
		if err != nil {
			return nil, err
		}
		realHdr.Name = strings.TrimRight(name, "\x00")
		tr.currHdr = realHdr
		tr.currSize = realHdr.Size
		return realHdr, nil
	}
	if hdr.Typeflag == 'K' {
		linkname, err := readAllString(tr.r, hdr.Size)
		if err != nil {
			return nil, err
		}
		padded := padToBlock(hdr.Size)
		if padded > hdr.Size {
			io.CopyN(io.Discard, tr.r, padded-hdr.Size)
		}
		nextBlock := make([]byte, blockSize)
		if _, err := io.ReadFull(tr.r, nextBlock); err != nil {
			return nil, err
		}
		realHdr, err := tr.parseHeader(nextBlock)
		if err != nil {
			return nil, err
		}
		realHdr.Linkname = strings.TrimRight(linkname, "\x00")
		tr.currHdr = realHdr
		tr.currSize = realHdr.Size
		return realHdr, nil
	}
	if hdr.Typeflag == TypeXHeader {
		paxData, err := readAllString(tr.r, hdr.Size)
		if err != nil {
			return nil, err
		}
		padded := padToBlock(hdr.Size)
		if padded > hdr.Size {
			io.CopyN(io.Discard, tr.r, padded-hdr.Size)
		}
		nextBlock := make([]byte, blockSize)
		if _, err := io.ReadFull(tr.r, nextBlock); err != nil {
			return nil, err
		}
		realHdr, err := tr.parseHeader(nextBlock)
		if err != nil {
			return nil, err
		}
		parsePAXData(realHdr, paxData)
		tr.currHdr = realHdr
		tr.currSize = realHdr.Size
		return realHdr, nil
	}
	tr.currHdr = hdr
	tr.currSize = hdr.Size
	return hdr, nil
}

func (tr *Reader) parseHeader(block []byte) (*Header, error) {
	chksumField := strings.TrimRight(string(block[148:148+chksumLen]), " \x00")
	var chksum int64
	var err error
	if chksumField != "" {
		chksum, err = strconv.ParseInt(chksumField, 8, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid checksum: %v", err)
		}
	}
	computed := computeChecksum(block)
	if chksum != 0 && chksum != computed {
		return nil, errors.New("checksum mismatch")
	}
	format := parseMagic(block)
	hdr := &Header{Format: format}
	hdr.Name = strings.TrimRight(string(block[0:nameLen]), "\x00")
	hdr.Mode, _ = parseOctalString(string(block[100 : 100+modeLen]))
	uid, _ := parseOctalString(string(block[108 : 108+uidLen]))
	hdr.Uid = int(uid)
	gid, _ := parseOctalString(string(block[116 : 116+gidLen]))
	hdr.Gid = int(gid)
	hdr.Size, _ = parseOctalString(string(block[124 : 124+sizeLen]))
	mtime, _ := parseOctalString(string(block[136 : 136+mtimeLen]))
	hdr.Mtime = time.Unix(mtime, 0)
	if block[156] == 0 {
		hdr.Typeflag = TypeReg
	} else {
		hdr.Typeflag = block[156]
	}
	hdr.Linkname = strings.TrimRight(string(block[157:157+linknameLen]), "\x00")
	hdr.Uname = strings.TrimRight(string(block[265:265+unameLen]), "\x00")
	hdr.Gname = strings.TrimRight(string(block[297:297+gnameLen]), "\x00")
	hdr.Devmajor, _ = parseOctalString(string(block[329 : 329+devmajorLen]))
	hdr.Devminor, _ = parseOctalString(string(block[337 : 337+devminorLen]))
	if format == FormatPOSIX {
		prefix := strings.TrimRight(string(block[345:345+prefixLen]), "\x00")
		if prefix != "" {
			hdr.Name = prefix + "/" + hdr.Name
		}
	}
	return hdr, nil
}

func parsePAXData(hdr *Header, data string) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		idx := strings.Index(line, " ")
		if idx == -1 {
			continue
		}
		rest := line[idx+1:]
		eqIdx := strings.Index(rest, "=")
		if eqIdx == -1 {
			continue
		}
		key := rest[:eqIdx]
		value := rest[eqIdx+1:]
		if hdr.PAX == nil {
			hdr.PAX = make(map[string]string)
		}
		hdr.PAX[key] = value
		switch key {
		case "path":
			hdr.Name = value
		case "linkpath":
			hdr.Linkname = value
		case "size":
			if size, err := strconv.ParseInt(value, 10, 64); err == nil {
				hdr.Size = size
			}
		}
	}
}

func isZeroBlock(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

func (tr *Reader) Read(b []byte) (int, error) {
	if tr.currHdr == nil {
		return 0, io.EOF
	}
	if tr.currRead >= tr.currSize {
		return 0, io.EOF
	}
	remaining := tr.currSize - tr.currRead
	if int64(len(b)) > remaining {
		b = b[:remaining]
	}
	n, err := tr.r.Read(b)
	tr.currRead += int64(n)
	return n, err
}

func ExtractArchive(r io.Reader, destDir string) error {
	if destDir == "" {
		destDir = "."
	}
	tr := NewReader(r)
	hardlinks := make(map[string]string)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := extractEntry(tr, hdr, destDir, hardlinks); err != nil {
			return err
		}
	}
	return nil
}

func extractEntry(tr *Reader, hdr *Header, destDir string, hardlinks map[string]string) error {
	targetPath := filepath.Join(destDir, hdr.Name)
	switch hdr.Typeflag {
	case TypeDir:
		return extractDir(hdr, targetPath)
	case TypeReg:
		return extractRegular(tr, hdr, targetPath, hardlinks)
	case TypeSymlink:
		return extractSymlink(hdr, targetPath)
	case TypeLink:
		return extractHardlink(hdr, targetPath, destDir, hardlinks)
	case TypeFifo:
		return extractFifo(hdr, targetPath)
	default:
		if hdr.Typeflag == 0 {
			return extractRegular(tr, hdr, targetPath, hardlinks)
		}
		return nil
	}
}

func extractDir(hdr *Header, path string) error {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	if err := os.MkdirAll(path, os.FileMode(hdr.Mode)|0700); err != nil {
		return err
	}
	if err := os.Chmod(path, os.FileMode(hdr.Mode)); err != nil {
		return err
	}
	if err := os.Chtimes(path, hdr.Mtime, hdr.Mtime); err != nil {
		return err
	}
	return nil
}

func extractRegular(tr *Reader, hdr *Header, path string, hardlinks map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
	if err != nil {
		return err
	}
	defer f.Close()
	if hdr.Size > 0 {
		if _, err := io.CopyN(f, tr, hdr.Size); err != nil {
			return err
		}
	}
	if err := f.Chmod(os.FileMode(hdr.Mode)); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chtimes(path, hdr.Mtime, hdr.Mtime); err != nil {
		return err
	}
	hardlinks[hdr.Name] = path
	return nil
}

func extractSymlink(hdr *Header, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	if err := os.Symlink(hdr.Linkname, path); err != nil {
		return err
	}
	return nil
}

func extractHardlink(hdr *Header, path string, destDir string, hardlinks map[string]string) error {
	existingPath, ok := hardlinks[hdr.Linkname]
	if !ok {
		existingPath = filepath.Join(destDir, hdr.Linkname)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	if err := os.Link(existingPath, path); err != nil {
		return err
	}
	hardlinks[hdr.Name] = path
	return nil
}

func extractFifo(hdr *Header, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return syscall.Mkfifo(path, uint32(hdr.Mode))
}
