package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SourceFileInfo struct {
	Path           string
	BaseName       string
	HasOSSuffix    bool
	HasARCHSuffix  bool
	GOOS           string
	GOARCH         string
	BuildTags      []string
	ShouldBuild    map[string]bool
}

var fileNameRegex = regexp.MustCompile(`^(.+?)(?:_([a-z0-9]+))?(?:_([a-z0-9]+))?\.go$`)

var buildTagRegex = regexp.MustCompile(`^//\s*\+build\s+(.+)$`)

func ScanSourceFiles(dir string) ([]*SourceFileInfo, error) {
	var files []*SourceFileInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		fileInfo, err := AnalyzeSourceFile(path)
		if err != nil {
			return err
		}

		files = append(files, fileInfo)
		return nil
	})

	return files, err
}

func AnalyzeSourceFile(path string) (*SourceFileInfo, error) {
	info := &SourceFileInfo{
		Path:        path,
		ShouldBuild: make(map[string]bool),
	}

	base := filepath.Base(path)
	info.BaseName = base

	match := fileNameRegex.FindStringSubmatch(base)
	if match != nil {
		if len(match) >= 3 && match[2] != "" {
			if IsValidGOOS(match[2]) {
				info.HasOSSuffix = true
				info.GOOS = match[2]
			} else if IsValidGOARCH(match[2]) {
				info.HasARCHSuffix = true
				info.GOARCH = match[2]
			}
		}
		if len(match) >= 4 && match[3] != "" {
			if info.HasOSSuffix && IsValidGOARCH(match[3]) {
				info.HasARCHSuffix = true
				info.GOARCH = match[3]
			}
		}
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "//") {
			break
		}

		tagMatch := buildTagRegex.FindStringSubmatch(line)
		if tagMatch != nil {
			info.BuildTags = append(info.BuildTags, tagMatch[1])
		}
	}

	for _, target := range GetDefaultTargets() {
		key := target.GOOS + "/" + target.GOARCH
		info.ShouldBuild[key] = ShouldBuildForPlatform(info, target.GOOS, target.GOARCH)
	}

	return info, nil
}

func ShouldBuildForPlatform(info *SourceFileInfo, goos, goarch string) bool {
	if info.HasOSSuffix && info.GOOS != goos {
		return false
	}
	if info.HasARCHSuffix && info.GOARCH != goarch {
		return false
	}

	if len(info.BuildTags) == 0 {
		return true
	}

	for _, tagExpr := range info.BuildTags {
		result, err := EvaluateBuildTag(tagExpr, goos, goarch)
		if err != nil {
			return false
		}
		if !result {
			return false
		}
	}

	return true
}

func DetectCGODependency(dir string) (bool, []string, error) {
	var cgoFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.Contains(string(content), `"C"`) ||
			strings.Contains(string(content), "import \"C\"") {
			cgoFiles = append(cgoFiles, path)
		}

		return nil
	})

	return len(cgoFiles) > 0, cgoFiles, err
}

func ValidatePlatformFiles(files []*SourceFileInfo, goos, goarch string) []string {
	var warnings []string
	key := goos + "/" + goarch

	for _, file := range files {
		if file.HasOSSuffix || file.HasARCHSuffix {
			if file.GOOS != "" && file.GOOS != goos {
				if file.ShouldBuild[key] {
					warnings = append(warnings,
						fmt.Sprintf("File %s has OS suffix %s but claims to build for %s",
							file.Path, file.GOOS, goos))
				}
			}
			if file.GOARCH != "" && file.GOARCH != goarch {
				if file.ShouldBuild[key] {
					warnings = append(warnings,
						fmt.Sprintf("File %s has ARCH suffix %s but claims to build for %s",
							file.Path, file.GOARCH, goarch))
				}
			}
		}

		if len(file.BuildTags) > 0 {
			expected := ShouldBuildForPlatform(file, goos, goarch)
			if file.ShouldBuild[key] != expected {
				warnings = append(warnings,
					fmt.Sprintf("File %s build tag mismatch for %s/%s",
						file.Path, goos, goarch))
			}
		}
	}

	return warnings
}
