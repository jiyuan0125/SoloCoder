package server

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	
	"disk-usage/internal/protocol"
)

type Scanner struct {
	MaxDepth   int
	MinSize    int64
	IncludeAll bool
	warnings   []string
}

func NewScanner(maxDepth int, minSize int64, includeAll bool) *Scanner {
	return &Scanner{
		MaxDepth:   maxDepth,
		MinSize:    minSize,
		IncludeAll: includeAll,
		warnings:   make([]string, 0),
	}
}

func (s *Scanner) Scan(rootPath string) (*protocol.ScanResult, error) {
	startTime := time.Now()
	
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	
	if !info.IsDir() {
		return nil, err
	}
	
	rootDepth := len(strings.Split(filepath.Clean(absPath), string(os.PathSeparator)))
	
	result := &protocol.ScanResult{
		Success:     true,
		Directories: make([]protocol.DirInfo, 0),
		Warnings:    make([]string, 0),
	}
	
	s.warnings = make([]string, 0)
	
	rootInfo, err := s.scanDirectory(absPath, rootDepth, rootDepth)
	if err != nil {
		result.Success = false
	}
	
	result.TotalSize = rootInfo.Size
	result.TotalFiles = rootInfo.Files
	result.TotalDirs = rootInfo.Dirs
	result.Warnings = s.warnings
	result.ScanTime = time.Since(startTime).Milliseconds()
	
	if s.MaxDepth > 0 {
		result.Directories = s.collectTopLevelDirs(absPath, rootDepth)
	}
	
	sort.Slice(result.Directories, func(i, j int) bool {
		return result.Directories[i].Size > result.Directories[j].Size
	})
	
	return result, nil
}

func (s *Scanner) scanDirectory(path string, currentDepth, baseDepth int) (*protocol.DirInfo, error) {
	dirInfo := &protocol.DirInfo{
		Path:  path,
		Depth: currentDepth - baseDepth,
	}
	
	entries, err := os.ReadDir(path)
	if err != nil {
		s.warnings = append(s.warnings, "无法访问目录: "+path+", 错误: "+err.Error())
		return dirInfo, err
	}
	
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		
		if !s.IncludeAll && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		
		if entry.IsDir() {
			dirInfo.Dirs++
			subInfo, _ := s.scanDirectory(entryPath, currentDepth+1, baseDepth)
			dirInfo.Size += subInfo.Size
			dirInfo.Files += subInfo.Files
			dirInfo.Dirs += subInfo.Dirs
		} else {
			info, err := entry.Info()
			if err != nil {
				s.warnings = append(s.warnings, "无法获取文件信息: "+entryPath+", 错误: "+err.Error())
				continue
			}
			
			dirInfo.Size += info.Size()
			dirInfo.Files++
		}
	}
	
	return dirInfo, nil
}

func (s *Scanner) collectTopLevelDirs(path string, baseDepth int) []protocol.DirInfo {
	dirs := make([]protocol.DirInfo, 0)
	
	entries, err := os.ReadDir(path)
	if err != nil {
		return dirs
	}
	
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		
		if entry.IsDir() {
			entryPath := filepath.Join(path, entry.Name())
			
			if !s.IncludeAll && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			
			currentDepth := len(strings.Split(filepath.Clean(entryPath), string(os.PathSeparator)))
			
			info, err := s.scanDirectory(entryPath, currentDepth, baseDepth)
			if err != nil {
				continue
			}
			
			if s.MinSize > 0 && info.Size < s.MinSize {
				continue
			}
			
			if s.MaxDepth > 0 && info.Depth <= s.MaxDepth {
				dirs = append(dirs, *info)
				dirs = append(dirs, s.collectSubDirs(entryPath, currentDepth, baseDepth)...)
			}
		}
	}
	
	return dirs
}

func (s *Scanner) collectSubDirs(path string, currentDepth, baseDepth int) []protocol.DirInfo {
	dirs := make([]protocol.DirInfo, 0)
	
	depth := currentDepth - baseDepth
	if depth >= s.MaxDepth {
		return dirs
	}
	
	entries, err := os.ReadDir(path)
	if err != nil {
		return dirs
	}
	
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		
		if entry.IsDir() {
			entryPath := filepath.Join(path, entry.Name())
			
			if !s.IncludeAll && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			
			subDepth := currentDepth + 1
			depthDiff := subDepth - baseDepth
			
			info, err := s.scanDirectory(entryPath, subDepth, baseDepth)
			if err != nil {
				continue
			}
			
			if s.MinSize > 0 && info.Size < s.MinSize {
				continue
			}
			
			if depthDiff <= s.MaxDepth {
				dirs = append(dirs, *info)
				dirs = append(dirs, s.collectSubDirs(entryPath, subDepth, baseDepth)...)
			}
		}
	}
	
	return dirs
}
