package archive

import (
	"path/filepath"
	"regexp"
	"time"
)

type FilterFunc func(path string, info FileInfo) bool

type RenameFunc func(originalPath, archivePath string) string

type Options struct {
	CompressionLevel int
	IncludePatterns  []string
	ExcludePatterns  []string
	Filters          []FilterFunc
	Renamer          RenameFunc
	TimeStart        time.Time
	TimeEnd          time.Time
	StripPrefix      string
	AddPrefix        string
	FollowSymlinks   bool
	IncludeEmptyDirs bool
}

type FileInfo interface {
	Name() string
	Size() int64
	Mode() uint32
	ModTime() time.Time
	IsDir() bool
	IsSymlink() bool
}

func (o *Options) ShouldInclude(path string, info FileInfo) bool {
	if len(o.IncludePatterns) > 0 {
		matched := false
		for _, pattern := range o.IncludePatterns {
			if matchGlob(pattern, path) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	for _, pattern := range o.ExcludePatterns {
		if matchGlob(pattern, path) {
			return false
		}
	}

	if !o.TimeStart.IsZero() && info.ModTime().Before(o.TimeStart) {
		return false
	}

	if !o.TimeEnd.IsZero() && info.ModTime().After(o.TimeEnd) {
		return false
	}

	for _, filter := range o.Filters {
		if !filter(path, info) {
			return false
		}
	}

	if !o.IncludeEmptyDirs && info.IsDir() {
		return false
	}

	return true
}

func (o *Options) TransformPath(originalPath string) string {
	result := originalPath

	if o.StripPrefix != "" {
		result = stripPrefix(result, o.StripPrefix)
	}

	if o.Renamer != nil {
		result = o.Renamer(originalPath, result)
	}

	if o.AddPrefix != "" {
		result = filepath.Join(o.AddPrefix, result)
	}

	return filepath.ToSlash(result)
}

func matchGlob(pattern, path string) bool {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(path)
	}
	return matched
}

func stripPrefix(path, prefix string) string {
	prefix = filepath.Clean(prefix)
	path = filepath.Clean(path)

	if len(prefix) == 0 || prefix == "." {
		return path
	}

	if len(path) > len(prefix) {
		if path[:len(prefix)] == prefix {
			if len(path) == len(prefix)+1 && path[len(prefix)] == filepath.Separator {
				return "."
			}
			if path[len(prefix)] == filepath.Separator {
				return path[len(prefix)+1:]
			}
		}
	}

	return path
}
