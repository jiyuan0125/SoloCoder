package sqlmigrate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type Migration struct {
	Version     string
	Description string
	UpSQL       string
	DownSQL     string
	UpPath      string
	DownPath    string
}

type Scanner struct {
	dir string
}

func NewScanner(dir string) *Scanner {
	return &Scanner{dir: dir}
}

func (s *Scanner) Scan() ([]*Migration, error) {
	if _, err := os.Stat(s.dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", s.dir)
	}

	upFiles := make(map[string]string)
	downFiles := make(map[string]string)
	versionSet := make(map[string]bool)

	versionRegex := regexp.MustCompile(`^(\d+)_([^.]+)\.up\.sql$`)
	downVersionRegex := regexp.MustCompile(`^(\d+)_([^.]+)\.down\.sql$`)

	err := filepath.Walk(s.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()

		// 处理 .up.sql 文件
		if upMatch := versionRegex.FindStringSubmatch(filename); upMatch != nil {
			version := upMatch[1]
			description := upMatch[2]

			if existingPath, exists := upFiles[version]; exists {
				return fmt.Errorf("duplicate version '%s' found: %s and %s", version, existingPath, path)
			}

			upFiles[version] = path
			versionSet[version] = true

			// 检查是否有对应的 down 文件
			downFilename := fmt.Sprintf("%s_%s.down.sql", version, description)
			downPath := filepath.Join(filepath.Dir(path), downFilename)
			if _, err := os.Stat(downPath); err == nil {
				downFiles[version] = downPath
			}
		}

		// 处理 .down.sql 文件
		if downMatch := downVersionRegex.FindStringSubmatch(filename); downMatch != nil {
			version := downMatch[1]
			description := downMatch[2]

			downFiles[version] = path

			// 检查是否有对应的 up 文件
			upFilename := fmt.Sprintf("%s_%s.up.sql", version, description)
			upPath := filepath.Join(filepath.Dir(path), upFilename)
			if _, err := os.Stat(upPath); err == nil {
				upFiles[version] = upPath
				versionSet[version] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 构建迁移列表
	var migrations []*Migration
	for version := range versionSet {
		migration, err := s.buildMigration(version, upFiles, downFiles)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	// 按 version 排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (s *Scanner) buildMigration(version string, upFiles, downFiles map[string]string) (*Migration, error) {
	migration := &Migration{
		Version: version,
	}

	// 读取 up.sql
	if upPath, exists := upFiles[version]; exists {
		migration.UpPath = upPath
		upContent, err := ioutil.ReadFile(upPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read up file %s: %w", upPath, err)
		}
		migration.UpSQL = string(upContent)

		// 从文件名提取 description
		filename := filepath.Base(upPath)
		descRegex := regexp.MustCompile(`^\d+_([^.]+)\.up\.sql$`)
		if match := descRegex.FindStringSubmatch(filename); match != nil {
			migration.Description = match[1]
		}
	}

	// 读取 down.sql
	if downPath, exists := downFiles[version]; exists {
		migration.DownPath = downPath
		downContent, err := ioutil.ReadFile(downPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read down file %s: %w", downPath, err)
		}
		migration.DownSQL = string(downContent)
	}

	return migration, nil
}
