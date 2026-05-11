package vfs

import (
	"strings"
)

type TreeNode struct {
	Name     string
	IsDir    bool
	Path     string
	Children []*TreeNode
}

func BuildTree(vfs *VirtualFileSystem, basePath string) (*TreeNode, error) {
	info, err := vfs.Stat(basePath)
	if err != nil {
		return nil, err
	}

	node := &TreeNode{
		Name:  info.Name(),
		IsDir: info.IsDir(),
		Path:  basePath,
	}

	if !info.IsDir() {
		return node, nil
	}

	entries, err := vfs.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		childPath := basePath
		if basePath == "." {
			childPath = entry.Name()
		} else {
			childPath = strings.Join([]string{basePath, entry.Name()}, "/")
		}

		child, err := BuildTree(vfs, childPath)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}
