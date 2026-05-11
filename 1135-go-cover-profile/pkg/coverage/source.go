package coverage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type SourceAnalysis struct {
	ExecutableLines map[int]bool
	Functions       []FunctionInfo
}

type FunctionInfo struct {
	Name      string
	StartLine int
	EndLine   int
}

func AnalyzeSource(filePath string) (*SourceAnalysis, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return AnalyzeSourceContent(string(content), filePath)
}

func AnalyzeSourceContent(content string, filePath string) (*SourceAnalysis, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analysis := &SourceAnalysis{
		ExecutableLines: make(map[int]bool),
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			info := FunctionInfo{
				Name:      node.Name.Name,
				StartLine: fset.Position(node.Pos()).Line,
				EndLine:   fset.Position(node.End()).Line,
			}
			analysis.Functions = append(analysis.Functions, info)

			if node.Body != nil {
				markExecutableLines(node.Body, fset, analysis)
			}

		case *ast.AssignStmt, *ast.IncDecStmt, *ast.GoStmt, *ast.DeferStmt,
			*ast.ReturnStmt, *ast.SendStmt, *ast.ExprStmt:
			pos := fset.Position(n.Pos())
			line := pos.Line
			analysis.ExecutableLines[line] = true

		case *ast.IfStmt:
			pos := fset.Position(node.Pos())
			analysis.ExecutableLines[pos.Line] = true
			if node.Body != nil {
				markExecutableLines(node.Body, fset, analysis)
			}
			if node.Else != nil {
				markExecutableLines(node.Else, fset, analysis)
			}

		case *ast.SwitchStmt:
			pos := fset.Position(node.Pos())
			analysis.ExecutableLines[pos.Line] = true
			if node.Body != nil {
				for _, stmt := range node.Body.List {
					if cc, ok := stmt.(*ast.CaseClause); ok {
						casePos := fset.Position(cc.Pos())
						analysis.ExecutableLines[casePos.Line] = true
						for _, s := range cc.Body {
							markExecutableLines(s, fset, analysis)
						}
					}
				}
			}

		case *ast.SelectStmt:
			pos := fset.Position(node.Pos())
			analysis.ExecutableLines[pos.Line] = true
			if node.Body != nil {
				for _, stmt := range node.Body.List {
					if cc, ok := stmt.(*ast.CommClause); ok {
						casePos := fsetPosition(cc.Pos(), fset)
						analysis.ExecutableLines[casePos] = true
						for _, s := range cc.Body {
							markExecutableLines(s, fset, analysis)
						}
					}
				}
			}

		case *ast.ForStmt, *ast.RangeStmt:
			pos := fset.Position(n.Pos())
			analysis.ExecutableLines[pos.Line] = true
			if body, ok := n.(interface{ GetBody() *ast.BlockStmt }); ok {
				if b := body.GetBody(); b != nil {
					markExecutableLines(b, fset, analysis)
				}
			} else if fs, ok := n.(*ast.ForStmt); ok && fs.Body != nil {
				markExecutableLines(fs.Body, fset, analysis)
			} else if rs, ok := n.(*ast.RangeStmt); ok && rs.Body != nil {
				markExecutableLines(rs.Body, fset, analysis)
			}

		case *ast.TypeSwitchStmt:
			pos := fset.Position(node.Pos())
			analysis.ExecutableLines[pos.Line] = true
			if node.Body != nil {
				for _, stmt := range node.Body.List {
					if cc, ok := stmt.(*ast.CaseClause); ok {
						casePos := fsetPosition(cc.Pos(), fset)
						analysis.ExecutableLines[casePos] = true
						for _, s := range cc.Body {
							markExecutableLines(s, fset, analysis)
						}
					}
				}
			}
		}

		return true
	})

	lines := strings.Split(content, "\n")
	for lineNum := 1; lineNum <= len(lines); lineNum++ {
		line := lines[lineNum-1]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "package ") ||
			strings.HasPrefix(trimmed, "import ") {
			delete(analysis.ExecutableLines, lineNum)
		}
	}

	return analysis, nil
}

func fsetPosition(pos token.Pos, fset *token.FileSet) int {
	return fset.Position(pos).Line
}

func markExecutableLines(n ast.Node, fset *token.FileSet, analysis *SourceAnalysis) {
	ast.Inspect(n, func(child ast.Node) bool {
		if child == nil {
			return false
		}

		switch child.(type) {
		case *ast.AssignStmt, *ast.IncDecStmt, *ast.GoStmt, *ast.DeferStmt,
			*ast.ReturnStmt, *ast.SendStmt, *ast.ExprStmt,
			*ast.IfStmt, *ast.SwitchStmt, *ast.SelectStmt,
			*ast.ForStmt, *ast.RangeStmt, *ast.TypeSwitchStmt:
			pos := fset.Position(child.Pos())
			analysis.ExecutableLines[pos.Line] = true
		}

		return true
	})
}
