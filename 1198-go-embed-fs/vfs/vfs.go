package vfs

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotImplemented = errors.New("not implemented")

type VirtualFS interface {
	fs.FS
	fs.GlobFS
	fs.ReadDirFS
	fs.StatFS
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	Glob(pattern string) ([]string, error)
	Stat(name string) (fs.FileInfo, error)
}

type LayerType int

const (
	LayerEmbed LayerType = iota
	LayerDisk
)

type FSLayer struct {
	Type   LayerType
	FS     fs.FS
	Base   string
}

type Config struct {
	EmbedFS     fs.FS
	DiskDir     string
	UseHotReload bool
	ReloadChan  <-chan time.Time
}

type VirtualFileSystem struct {
	config     Config
	layers     []*FSLayer
	cache      map[string]*cacheEntry
	cacheMu    sync.RWMutex
}

type cacheEntry struct {
	data    []byte
	modTime time.Time
	layer   *FSLayer
}

func New(config Config) (*VirtualFileSystem, error) {
	if config.EmbedFS == nil && config.DiskDir == "" {
		return nil, errors.New("at least one of EmbedFS or DiskDir must be provided")
	}

	vfs := &VirtualFileSystem{
		config: config,
		cache:  make(map[string]*cacheEntry),
	}

	if config.DiskDir != "" {
		info, err := os.Stat(config.DiskDir)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, errors.New("DiskDir must be a directory")
		}
		vfs.layers = append(vfs.layers, &FSLayer{
			Type: LayerDisk,
			FS:   os.DirFS(config.DiskDir),
			Base: config.DiskDir,
		})
	}

	if config.EmbedFS != nil {
		vfs.layers = append(vfs.layers, &FSLayer{
			Type: LayerEmbed,
			FS:   config.EmbedFS,
		})
	}

	if config.UseHotReload {
		go vfs.watch()
	}

	return vfs, nil
}

func (v *VirtualFileSystem) normalizePath(name string) string {
	name = strings.ReplaceAll(name, string(os.PathSeparator), "/")
	name = path.Clean(name)
	if name == "." {
		return "."
	}
	if name == ".." || strings.HasPrefix(name, "../") {
		return name
	}
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	return name
}

func (v *VirtualFileSystem) watch() {
	if v.config.ReloadChan == nil {
		return
	}
	for range v.config.ReloadChan {
		v.Reload()
	}
}

func (v *VirtualFileSystem) Reload() {
	v.cacheMu.Lock()
	defer v.cacheMu.Unlock()
	v.cache = make(map[string]*cacheEntry)
}

func (v *VirtualFileSystem) findFile(name string) (*FSLayer, error) {
	name = v.normalizePath(name)
	
	if strings.Contains(name, "..") {
		return nil, &fs.PathError{Op: "findFile", Path: name, Err: fs.ErrInvalid}
	}

	for _, layer := range v.layers {
		_, err := fs.Stat(layer.FS, name)
		if err == nil {
			return layer, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return nil, &fs.PathError{Op: "findFile", Path: name, Err: fs.ErrNotExist}
}

func (v *VirtualFileSystem) Open(name string) (fs.File, error) {
	name = v.normalizePath(name)
	
	if strings.Contains(name, "..") {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	for _, layer := range v.layers {
		file, err := layer.FS.Open(name)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (v *VirtualFileSystem) ReadFile(name string) ([]byte, error) {
	name = v.normalizePath(name)
	
	if strings.Contains(name, "..") {
		return nil, &fs.PathError{Op: "readFile", Path: name, Err: fs.ErrInvalid}
	}

	v.cacheMu.RLock()
	entry, ok := v.cache[name]
	v.cacheMu.RUnlock()

	if ok {
		return entry.data, nil
	}

	layer, err := v.findFile(name)
	if err != nil {
		return nil, err
	}

	data, err := fs.ReadFile(layer.FS, name)
	if err != nil {
		return nil, err
	}

	var modTime time.Time
	info, err := fs.Stat(layer.FS, name)
	if err == nil {
		modTime = info.ModTime()
	}

	v.cacheMu.Lock()
	v.cache[name] = &cacheEntry{
		data:    data,
		modTime: modTime,
		layer:   layer,
	}
	v.cacheMu.Unlock()

	return data, nil
}

func (v *VirtualFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	name = v.normalizePath(name)
	
	if strings.Contains(name, "..") {
		return nil, &fs.PathError{Op: "readDir", Path: name, Err: fs.ErrInvalid}
	}

	type dirEntryKey struct {
		name  string
		isDir bool
	}

	seen := make(map[dirEntryKey]bool)
	var entries []fs.DirEntry

	for _, layer := range v.layers {
		dirs, err := fs.ReadDir(layer.FS, name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		for _, e := range dirs {
			key := dirEntryKey{name: e.Name(), isDir: e.IsDir()}
			if !seen[key] {
				seen[key] = true
				entries = append(entries, e)
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

func (v *VirtualFileSystem) Glob(pattern string) ([]string, error) {
	if strings.Contains(pattern, "**") {
		return nil, errors.New("** pattern not supported, use standard glob syntax")
	}

	seen := make(map[string]bool)
	var results []string

	for _, layer := range v.layers {
		matches, err := fs.Glob(layer.FS, pattern)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				results = append(results, m)
			}
		}
	}

	sort.Strings(results)
	return results, nil
}

func (v *VirtualFileSystem) Stat(name string) (fs.FileInfo, error) {
	name = v.normalizePath(name)
	
	if strings.Contains(name, "..") {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}

	for _, layer := range v.layers {
		info, err := fs.Stat(layer.FS, name)
		if err == nil {
			return info, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

func (v *VirtualFileSystem) Layers() []*FSLayer {
	return v.layers
}

func (v *VirtualFileSystem) HasDiskLayer() bool {
	for _, l := range v.layers {
		if l.Type == LayerDisk {
			return true
		}
	}
	return false
}

func (v *VirtualFileSystem) HasEmbedLayer() bool {
	for _, l := range v.layers {
		if l.Type == LayerEmbed {
			return true
		}
	}
	return false
}
