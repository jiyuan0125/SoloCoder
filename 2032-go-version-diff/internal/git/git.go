package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repository struct {
	path string
}

func NewRepository(repoPath string) (*Repository, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, errors.New("仓库路径不存在")
	}

	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, errors.New("不是 Git 仓库")
	}

	return &Repository{path: absPath}, nil
}

func (r *Repository) Path() string {
	return r.path
}

func (r *Repository) Exists(ref string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("commit 不存在: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func (r *Repository) GetFileContent(ref, filePath string) (string, error) {
	if err := r.Exists(ref); err != nil {
		return "", err
	}

	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}
