package tar

import (
	"os"
	"time"
)

type Header struct {
	Name     string
	Linkname string
	Size     int64
	Mode     int64
	Uid      int
	Gid      int
	Uname    string
	Gname    string
	ModTime  time.Time
	Typeflag byte
	Xattrs   map[string]string
}

func (h *Header) FileInfo() os.FileInfo {
	return &headerFileInfo{h: h}
}

type headerFileInfo struct {
	h *Header
}

func (fi *headerFileInfo) Name() string {
	if len(fi.h.Name) == 0 {
		return ""
	}
	name := fi.h.Name
	if len(name) > 1 && name[len(name)-1] == '/' {
		name = name[:len(name)-1]
	}
	i := len(name) - 1
	for i >= 0 && name[i] != '/' {
		i--
	}
	if i >= 0 {
		name = name[i+1:]
	}
	return name
}

func (fi *headerFileInfo) Size() int64        { return fi.h.Size }
func (fi *headerFileInfo) Mode() os.FileMode  { return os.FileMode(fi.h.Mode) }
func (fi *headerFileInfo) ModTime() time.Time { return fi.h.ModTime }
func (fi *headerFileInfo) IsDir() bool        { return fi.h.Typeflag == TypeFlagDir }
func (fi *headerFileInfo) Sys() interface{}   { return fi.h }
