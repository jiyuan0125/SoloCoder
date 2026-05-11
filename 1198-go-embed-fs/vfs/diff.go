package vfs

import (
	"bytes"
	"io/fs"
	"sort"
	"strings"
)

type FileDiff struct {
	Path     string
	Status   string
	Details  string
}

func Compare(vfs *VirtualFileSystem) ([]FileDiff, error) {
	if !vfs.HasDiskLayer() || !vfs.HasEmbedLayer() {
		return nil, nil
	}

	var diskLayer, embedLayer *FSLayer
	for _, l := range vfs.Layers() {
		if l.Type == LayerDisk {
			diskLayer = l
		}
		if l.Type == LayerEmbed {
			embedLayer = l
		}
	}

	diskFiles, err := collectFiles(diskLayer.FS, ".")
	if err != nil {
		return nil, err
	}

	embedFiles, err := collectFiles(embedLayer.FS, ".")
	if err != nil {
		return nil, err
	}

	allFiles := make(map[string]bool)
	for f := range diskFiles {
		allFiles[f] = true
	}
	for f := range embedFiles {
		allFiles[f] = true
	}

	var diffs []FileDiff
	for path := range allFiles {
		inDisk := diskFiles[path]
		inEmbed := embedFiles[path]

		if inDisk && !inEmbed {
			diffs = append(diffs, FileDiff{
				Path:    path,
				Status:  "added",
				Details: "exists only on disk",
			})
		} else if !inDisk && inEmbed {
			diffs = append(diffs, FileDiff{
				Path:    path,
				Status:  "removed",
				Details: "exists only in embed",
			})
		} else {
			diskData, err1 := fs.ReadFile(diskLayer.FS, path)
			embedData, err2 := fs.ReadFile(embedLayer.FS, path)
			
			if err1 != nil || err2 != nil {
				diffs = append(diffs, FileDiff{
					Path:    path,
					Status:  "error",
					Details: "failed to read for comparison",
				})
				continue
			}

			if !bytes.Equal(diskData, embedData) {
				diffs = append(diffs, FileDiff{
					Path:    path,
					Status:  "modified",
					Details: "content differs",
				})
			}
		}
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Path < diffs[j].Path
	})

	return diffs, nil
}

func collectFiles(fsys fs.FS, dir string) (map[string]bool, error) {
	files := make(map[string]bool)
	
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		fullPath := dir
		if dir == "." {
			fullPath = e.Name()
		} else {
			fullPath = strings.Join([]string{dir, e.Name()}, "/")
		}

		if e.IsDir() {
			subFiles, err := collectFiles(fsys, fullPath)
			if err != nil {
				return nil, err
			}
			for f := range subFiles {
				files[f] = true
			}
		} else {
			files[fullPath] = true
		}
	}

	return files, nil
}
