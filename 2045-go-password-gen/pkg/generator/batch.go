package generator

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func (g *Generator) GenerateBatch(count int, length int, passwordType PasswordType, specialChars string, separator string, dictFile string) ([]string, error) {
	if count > MaxBatch {
		return nil, errors.New("单次最多 10000 个")
	}
	if count <= 0 {
		return nil, errors.New("数量必须大于 0")
	}

	results := make([]string, 0, count)
	seen := make(map[string]bool)

	maxAttempts := count * 10
	attempts := 0

	for len(results) < count && attempts < maxAttempts {
		var password string
		var err error

		if passwordType == TypeReadable {
			password, err = g.GenerateReadable(length, separator, dictFile)
		} else {
			password, err = g.Generate(length, passwordType, specialChars)
		}

		if err != nil {
			return nil, err
		}

		if !seen[password] {
			seen[password] = true
			results = append(results, password)
		}

		attempts++
	}

	if len(results) < count {
		return nil, fmt.Errorf("无法生成 %d 个不重复的密码（只生成了 %d 个）", count, len(results))
	}

	return results, nil
}

func (g *Generator) WriteToFile(passwords []string, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, password := range passwords {
		_, err := writer.WriteString(password + "\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	return nil
}
