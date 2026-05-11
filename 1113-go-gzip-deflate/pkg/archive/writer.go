package archive

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"archiver/internal/gzip"
	"archiver/internal/tar"
)

type ArchiveWriter struct {
	dst         io.Writer
	gzWriter    *gzip.Writer
	tarWriter   *tar.Writer
	options     *Options
	closed      bool
	writtenSize int64
	fileCount   int
}

type FileEntry struct {
	Path       string
	Reader     io.Reader
	Size       int64
	Mode       int64
	ModTime    int64
	IsDir      bool
	LinkTarget string
	UserID     int
	GroupID    int
	UserName   string
	GroupName  string
}

func NewWriter(dst io.Writer, opts *Options) (*ArchiveWriter, error) {
	if opts == nil {
		opts = &Options{}
	}

	gzHeader := &gzip.Header{
		OS: gzip.OSUnix,
	}

	gz := gzip.NewWriterHeader(dst, gzHeader)

	return &ArchiveWriter{
		dst:       dst,
		gzWriter:  gz,
		tarWriter: tar.NewWriter(gz),
		options:   opts,
	}, nil
}

func (w *ArchiveWriter) AddFilesFromRoot(rootPath string) error {
	if w.closed {
		return errors.New("writer is closed")
	}

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}

	var entries []string
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		entries = append(entries, relPath)
		return nil
	})

	if err != nil {
		return err
	}

	sort.Strings(entries)

	for _, relPath := range entries {
		fullPath := filepath.Join(absRoot, relPath)
		info, err := os.Lstat(fullPath)
		if err != nil {
			return err
		}

		fileInfo := &osFileInfo{FileInfo: info}

		if !w.options.ShouldInclude(relPath, fileInfo) {
			continue
		}

		transformedPath := w.options.TransformPath(relPath)

		if info.IsDir() {
			if err := w.addDir(transformedPath, info); err != nil {
				return err
			}
		} else if isSymlink(info) {
			if w.options.FollowSymlinks {
				target, err := filepath.EvalSymlinks(fullPath)
				if err == nil {
					targetInfo, err := os.Stat(target)
					if err == nil && !targetInfo.IsDir() {
						f, err := os.Open(target)
						if err == nil {
							defer f.Close()
							entry := &FileEntry{
								Path:    transformedPath,
								Reader:  f,
								Size:    targetInfo.Size(),
								Mode:    int64(targetInfo.Mode()),
								ModTime: targetInfo.ModTime().Unix(),
								IsDir:   false,
							}
							if err := w.AddFile(entry); err != nil {
								return err
							}
							continue
						}
					}
				}
			}

			linkTarget, err := os.Readlink(fullPath)
			if err != nil {
				return err
			}

			entry := &FileEntry{
				Path:       transformedPath,
				Size:       0,
				Mode:       int64(info.Mode()),
				ModTime:    info.ModTime().Unix(),
				IsDir:      false,
				LinkTarget: linkTarget,
			}
			if err := w.AddFile(entry); err != nil {
				return err
			}
		} else {
			f, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer f.Close()

			entry := &FileEntry{
				Path:    transformedPath,
				Reader:  f,
				Size:    info.Size(),
				Mode:    int64(info.Mode()),
				ModTime: info.ModTime().Unix(),
				IsDir:   false,
			}
			if err := w.AddFile(entry); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *ArchiveWriter) addDir(path string, info os.FileInfo) error {
	header := tar.NewHeaderFromInfo(
		ensureTrailingSlash(path),
		0,
		int64(info.Mode()),
		info.ModTime().Unix(),
		true,
		"",
	)
	return w.tarWriter.WriteHeader(header)
}

func (w *ArchiveWriter) AddFile(entry *FileEntry) error {
	if w.closed {
		return errors.New("writer is closed")
	}

	header, err := createHeader(entry)
	if err != nil {
		return err
	}

	if err := w.tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if entry.Reader != nil && !entry.IsDir && entry.Size > 0 {
		n, err := io.Copy(w.tarWriter, entry.Reader)
		if err != nil {
			return err
		}
		w.writtenSize += n
	}

	if err := w.tarWriter.Flush(); err != nil {
		return err
	}

	w.fileCount++
	return nil
}

func createHeader(entry *FileEntry) (*tar.Header, error) {
	if entry.Path == "" {
		return nil, errors.New("path is required")
	}

	header := tar.NewHeaderFromInfo(
		entry.Path,
		entry.Size,
		entry.Mode,
		entry.ModTime,
		entry.IsDir,
		entry.LinkTarget,
	)

	header.Uid = entry.UserID
	header.Gid = entry.GroupID
	header.Uname = entry.UserName
	header.Gname = entry.GroupName

	return header, nil
}

func (w *ArchiveWriter) Close() error {
	if w.closed {
		return nil
	}

	if err := w.tarWriter.Close(); err != nil {
		return err
	}
	if err := w.gzWriter.Close(); err != nil {
		return err
	}

	w.closed = true
	return nil
}

func (w *ArchiveWriter) Stats() (fileCount int, writtenSize int64) {
	return w.fileCount, w.writtenSize
}

func ensureTrailingSlash(path string) string {
	if path == "" {
		return ""
	}
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

func isSymlink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}

type osFileInfo struct {
	os.FileInfo
}

func (f *osFileInfo) Mode() uint32 {
	return uint32(f.FileInfo.Mode())
}

func (f *osFileInfo) IsSymlink() bool {
	return isSymlink(f.FileInfo)
}

func ArchiveDirectory(rootDir string, opts *Options) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer, err := NewWriter(buf, opts)
	if err != nil {
		return nil, err
	}

	if err := writer.AddFilesFromRoot(rootDir); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ArchiveFiles(files map[string]string, opts *Options) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer, err := NewWriter(buf, opts)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, srcPath := range paths {
		arcPath := files[srcPath]
		info, err := os.Lstat(srcPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", srcPath, err)
		}

		fileInfo := &osFileInfo{FileInfo: info}
		if !shouldInclude(opts, arcPath, fileInfo) {
			continue
		}

		transformedPath := opts.TransformPath(arcPath)

		if info.IsDir() {
			if err := writer.addDir(transformedPath, info); err != nil {
				return nil, err
			}
		} else if isSymlink(info) {
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return nil, err
			}

			entry := &FileEntry{
				Path:       transformedPath,
				Size:       0,
				Mode:       int64(info.Mode()),
				ModTime:    info.ModTime().Unix(),
				IsDir:      false,
				LinkTarget: linkTarget,
			}
			if err := writer.AddFile(entry); err != nil {
				return nil, err
			}
		} else {
			f, err := os.Open(srcPath)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			entry := &FileEntry{
				Path:    transformedPath,
				Reader:  f,
				Size:    info.Size(),
				Mode:    int64(info.Mode()),
				ModTime: info.ModTime().Unix(),
				IsDir:   false,
			}
			if err := writer.AddFile(entry); err != nil {
				return nil, err
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func shouldInclude(opts *Options, path string, info FileInfo) bool {
	if opts == nil {
		return true
	}
	return opts.ShouldInclude(path, info)
}
