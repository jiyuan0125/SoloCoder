package core

import (
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Writer struct {
	w      io.Writer
	format Format
	closed bool
}

func NewWriter(w io.Writer, format Format) *Writer {
	return &Writer{w: w, format: format}
}

func (tw *Writer) WriteHeader(hdr *Header) error {
	if tw.closed {
		return errors.New("archive/tar: write on closed writer")
	}
	if tw.format == FormatPOSIX {
		return tw.writePOSIXHeader(hdr)
	}
	return tw.writeGNUHeader(hdr)
}

func (tw *Writer) writePOSIXHeader(hdr *Header) error {
	if len(hdr.Name) > nameLen+prefixLen+1 {
		return tw.writePAXHeader(hdr)
	}
	block := make([]byte, blockSize)
	tw.fillHeader(block, hdr, FormatPOSIX)
	_, err := tw.w.Write(block)
	return err
}

func (tw *Writer) writeGNUHeader(hdr *Header) error {
	if len(hdr.Name) > nameLen {
		if err := tw.writeGNULongName(hdr.Name, TypeGNULongName); err != nil {
			return err
		}
	}
	if len(hdr.Linkname) > linknameLen {
		if err := tw.writeGNULongName(hdr.Linkname, 'K'); err != nil {
			return err
		}
	}
	block := make([]byte, blockSize)
	tw.fillHeader(block, hdr, FormatGNU)
	_, err := tw.w.Write(block)
	return err
}

func (tw *Writer) writeGNULongName(name string, typeflag byte) error {
	nameBytes := []byte(name)
	size := int64(len(nameBytes) + 1)
	padded := padToBlock(size)
	block := make([]byte, blockSize)
	copy(block[257:], magicGNU)
	copy(block[263:], versionGNU)
	copy(block[124:], octalString(size, sizeLen))
	block[156] = typeflag
	tw.computeAndSetChecksum(block)
	if _, err := tw.w.Write(block); err != nil {
		return err
	}
	data := make([]byte, padded)
	copy(data, nameBytes)
	_, err := tw.w.Write(data)
	return err
}

func (tw *Writer) fillHeader(block []byte, hdr *Header, format Format) {
	name := hdr.Name
	linkname := hdr.Linkname
	var prefix string
	if format == FormatPOSIX {
		prefix, name = splitPOSIXName(hdr.Name)
		if len(prefix) > prefixLen {
			prefix = prefix[:prefixLen]
		}
	}
	if len(name) > nameLen {
		name = name[:nameLen]
	}
	if len(linkname) > linknameLen {
		linkname = linkname[:linknameLen]
	}
	copy(block[0:], name)
	copy(block[100:], octalString(hdr.Mode, modeLen))
	copy(block[108:], octalString(int64(hdr.Uid), uidLen))
	copy(block[116:], octalString(int64(hdr.Gid), gidLen))
	copy(block[124:], octalString(hdr.Size, sizeLen))
	mtime := hdr.Mtime.Unix()
	if mtime < 0 {
		mtime = 0
	}
	copy(block[136:], octalString(mtime, mtimeLen))
	for i := 148; i < 148+chksumLen; i++ {
		block[i] = ' '
	}
	block[156] = hdr.Typeflag
	copy(block[157:], linkname)
	if format == FormatPOSIX {
		copy(block[257:], magicPOSIX)
		copy(block[263:], versionPOSIX)
	} else {
		copy(block[257:], magicGNU)
		copy(block[263:], versionGNU)
	}
	uname := hdr.Uname
	gname := hdr.Gname
	if len(uname) > unameLen {
		uname = uname[:unameLen]
	}
	if len(gname) > gnameLen {
		gname = gname[:gnameLen]
	}
	copy(block[265:], uname)
	copy(block[297:], gname)
	copy(block[329:], octalString(hdr.Devmajor, devmajorLen))
	copy(block[337:], octalString(hdr.Devminor, devminorLen))
	if format == FormatPOSIX && prefix != "" {
		copy(block[345:], prefix)
	}
	tw.computeAndSetChecksum(block)
}

func (tw *Writer) computeAndSetChecksum(block []byte) {
	chksum := computeChecksum(block)
	copy(block[148:], octalString(chksum, chksumLen))
	block[148+chksumLen-1] = 0
}

func (tw *Writer) writePAXHeader(hdr *Header) error {
	content := " path=" + hdr.Name + "\n"
	length := len(content) + 1
	for {
		tentative := strconv.Itoa(length) + content
		if len(tentative) == length {
			return tw.writePAXRecord(tentative, int64(len(tentative)), hdr)
		}
		length = len(tentative)
	}
}

func (tw *Writer) writePAXRecord(paxData string, paxSize int64, hdr *Header) error {
	paxHdr := &Header{
		Name:     "PaxHeaders/0",
		Mode:     0644,
		Typeflag: TypeXHeader,
		Size:     paxSize,
	}
	block := make([]byte, blockSize)
	tw.fillHeader(block, paxHdr, FormatPOSIX)
	if _, err := tw.w.Write(block); err != nil {
		return err
	}
	paxBytes := []byte(paxData)
	padded := padToBlock(paxSize)
	padding := make([]byte, padded-paxSize)
	if _, err := tw.w.Write(paxBytes); err != nil {
		return err
	}
	if _, err := tw.w.Write(padding); err != nil {
		return err
	}
	truncName := hdr.Name
	if len(truncName) > nameLen {
		truncName = truncName[:nameLen]
	}
	shortHdr := *hdr
	shortHdr.Name = truncName
	shortBlock := make([]byte, blockSize)
	tw.fillHeader(shortBlock, &shortHdr, FormatPOSIX)
	_, err := tw.w.Write(shortBlock)
	return err
}

func (tw *Writer) Write(b []byte) (int, error) {
	if tw.closed {
		return 0, errors.New("archive/tar: write on closed writer")
	}
	return tw.w.Write(b)
}

func (tw *Writer) Close() error {
	if tw.closed {
		return nil
	}
	tw.closed = true
	trailer := make([]byte, blockSize*2)
	_, err := tw.w.Write(trailer)
	return err
}

func ArchivePaths(paths []string, format Format) (*io.PipeReader, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		tw := NewWriter(pw, format)
		defer tw.Close()
		seenInodes := make(map[uint64]string)
		for _, path := range paths {
			if err := archivePathRecursive(tw, path, "", seenInodes); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()
	return pr, nil
}

func archivePathRecursive(tw *Writer, path, base string, seenInodes map[uint64]string) error {
	cleanPath := filepath.Clean(path)
	fi, err := os.Lstat(cleanPath)
	if err != nil {
		return err
	}
	name := filepath.Join(base, filepath.Base(cleanPath))
	if fi.Mode()&os.ModeSymlink != 0 {
		return archiveSymlink(tw, cleanPath, name, fi)
	}
	if fi.IsDir() {
		return archiveDirectory(tw, cleanPath, name, seenInodes)
	}
	if fi.Mode()&os.ModeNamedPipe != 0 {
		return archivePipe(tw, name, fi)
	}
	return archiveRegularOrHardlink(tw, cleanPath, name, fi, seenInodes)
}

func archiveSymlink(tw *Writer, path, name string, fi os.FileInfo) error {
	target, err := os.Readlink(path)
	if err != nil {
		return err
	}
	hdr := &Header{
		Name:     name,
		Typeflag: TypeSymlink,
		Linkname: target,
		Size:     0,
	}
	fillHeaderFromStat(hdr, fi, path)
	return tw.WriteHeader(hdr)
}

func archivePipe(tw *Writer, name string, fi os.FileInfo) error {
	hdr := &Header{
		Name:     name,
		Typeflag: TypeFifo,
		Size:     0,
	}
	fillHeaderFromStat(hdr, fi, "")
	return tw.WriteHeader(hdr)
}

func archiveRegularOrHardlink(tw *Writer, path, name string, fi os.FileInfo, seenInodes map[uint64]string) error {
	var inode uint64
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		inode = stat.Ino
		if linkname, exists := seenInodes[inode]; exists {
			hdr := &Header{
				Name:     name,
				Typeflag: TypeLink,
				Linkname: linkname,
				Size:     0,
			}
			fillHeaderFromStat(hdr, fi, path)
			return tw.WriteHeader(hdr)
		}
		seenInodes[inode] = name
	}
	hdr := &Header{Name: name}
	hdr.fillFromInfo(fi)
	fillHeaderFromStat(hdr, fi, path)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if hdr.Typeflag == TypeReg && hdr.Size > 0 {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
		padded := padToBlock(hdr.Size)
		padding := int(padded - hdr.Size)
		if padding > 0 {
			_, err = tw.Write(make([]byte, padding))
			return err
		}
	}
	return nil
}

func archiveDirectory(tw *Writer, path, name string, seenInodes map[uint64]string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}
	hdr := &Header{Name: name}
	hdr.fillFromInfo(fi)
	fillHeaderFromStat(hdr, fi, path)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		if err := archivePathRecursive(tw, childPath, name, seenInodes); err != nil {
			return err
		}
	}
	return nil
}

func fillHeaderFromStat(hdr *Header, fi os.FileInfo, path string) {
	if hdr.Mode == 0 {
		hdr.Mode = int64(fi.Mode().Perm())
	}
	if hdr.Mtime.IsZero() {
		hdr.Mtime = fi.ModTime()
	}
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		hdr.Uid = int(stat.Uid)
		hdr.Gid = int(stat.Gid)
		if u, err := user.LookupId(strconv.Itoa(hdr.Uid)); err == nil {
			hdr.Uname = u.Username
		}
		if g, err := user.LookupGroupId(strconv.Itoa(hdr.Gid)); err == nil {
			hdr.Gname = g.Name
		}
		hdr.Devmajor = int64(stat.Rdev >> 8)
		hdr.Devminor = int64(stat.Rdev & 0xff)
	}
	if strings.HasSuffix(hdr.Name, "/") && hdr.Typeflag == 0 {
		hdr.Typeflag = TypeDir
	}
}
