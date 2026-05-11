package core

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Header struct {
	Name     string
	Mode     int64
	Uid      int
	Gid      int
	Size     int64
	Mtime    time.Time
	Typeflag byte
	Linkname string
	Uname    string
	Gname    string
	Devmajor int64
	Devminor int64
	PAX      map[string]string
	Format   Format
}

func (h *Header) fillFromInfo(fi os.FileInfo) {
	h.Mode = int64(fi.Mode().Perm())
	h.Mtime = fi.ModTime()
	if fi.IsDir() {
		h.Typeflag = TypeDir
		h.Size = 0
		if !strings.HasSuffix(h.Name, "/") {
			h.Name += "/"
		}
	} else if fi.Mode()&os.ModeSymlink != 0 {
		h.Typeflag = TypeSymlink
		h.Size = 0
	} else if fi.Mode()&os.ModeDevice != 0 {
		if fi.Mode()&os.ModeCharDevice != 0 {
			h.Typeflag = TypeChar
		} else {
			h.Typeflag = TypeBlock
		}
		h.Size = 0
	} else if fi.Mode()&os.ModeNamedPipe != 0 {
		h.Typeflag = TypeFifo
		h.Size = 0
	} else {
		h.Typeflag = TypeReg
		h.Size = fi.Size()
	}
}

func octalString(v int64, width int) string {
	s := strconv.FormatInt(v, 8)
	for len(s) < width-1 {
		s = "0" + s
	}
	return s
}

func parseOctalString(s string) (int64, error) {
	s = strings.TrimRight(s, " \x00")
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 8, 64)
}

func computeChecksum(block []byte) int64 {
	var sum int64
	for i := 0; i < 148; i++ {
		sum += int64(block[i])
	}
	for i := 148 + chksumLen; i < blockSize; i++ {
		sum += int64(block[i])
	}
	for i := 0; i < chksumLen; i++ {
		sum += int64(' ')
	}
	return sum
}

func padToBlock(n int64) int64 {
	return (n + blockSize - 1) &^ (blockSize - 1)
}

func splitPOSIXName(name string) (prefix, suffix string) {
	if len(name) <= nameLen {
		return "", name
	}
	idx := len(name) - nameLen
	for idx > 0 && name[idx] != '/' {
		idx--
	}
	if idx <= 0 || idx > prefixLen {
		return "", name[:nameLen]
	}
	return name[:idx], name[idx+1:]
}

func writeFullName(name string) []byte {
	buf := make([]byte, len(name)+1)
	copy(buf, name)
	return buf
}

func readAllString(r io.Reader, size int64) (string, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	n := len(buf)
	for n > 0 && buf[n-1] == 0 {
		n--
	}
	return string(buf[:n]), nil
}
