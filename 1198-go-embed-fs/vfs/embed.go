package vfs

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

type EmbedDirective struct {
	IncludeAll   bool
	Patterns     []string
	HasWildcards bool
}

func ParseEmbedDirective(line string) *EmbedDirective {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "//go:embed") {
		return nil
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return &EmbedDirective{}
	}

	directive := &EmbedDirective{}

	for _, pattern := range parts[1:] {
		if strings.HasPrefix(pattern, "all:") {
			pattern = strings.TrimPrefix(pattern, "all:")
			directive.IncludeAll = true
		}

		if strings.ContainsAny(pattern, "*?[") {
			directive.HasWildcards = true
		}

		directive.Patterns = append(directive.Patterns, pattern)
	}

	return directive
}

func ParseEmbedDirectivesFromFile(filePath string) ([]*EmbedDirective, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var directives []*EmbedDirective

	for _, commentGroup := range node.Comments {
		for _, comment := range commentGroup.List {
			if strings.HasPrefix(comment.Text, "//go:embed") {
				directive := ParseEmbedDirective(comment.Text)
				if directive != nil {
					directives = append(directives, directive)
				}
			}
		}
	}

	return directives, nil
}

func ParseEmbedDirectivesFromSource(pkgRoot string) ([]*EmbedDirective, error) {
	goFiles, err := filepath.Glob(filepath.Join(pkgRoot, "*.go"))
	if err != nil {
		return nil, err
	}

	var allDirectives []*EmbedDirective

	for _, goFile := range goFiles {
		directives, err := ParseEmbedDirectivesFromFile(goFile)
		if err != nil {
			continue
		}
		allDirectives = append(allDirectives, directives...)
	}

	return allDirectives, nil
}
