package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/example/go-ast-analyzer/internal/api"
)

type complexityVisitor struct {
	filename    string
	fset        *token.FileSet
	pkgName     string
	results     *[]api.FunctionComplexity
}

func AnalyzeComplexity(filename string, fset *token.FileSet, file *ast.File) []api.FunctionComplexity {
	results := []api.FunctionComplexity{}
	v := &complexityVisitor{
		filename: filepath.Base(filename),
		fset:     fset,
		pkgName:  file.Name.Name,
		results:  &results,
	}
	ast.Walk(v, file)
	return results
}

func (v *complexityVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			return nil
		}
		complexity := 1 + countBranchComplexity(n.Body)
		var funcName string
		if n.Recv != nil && len(n.Recv.List) > 0 {
			recvType := typeName(n.Recv.List[0].Type)
			funcName = fmt.Sprintf("%s.%s.%s", v.pkgName, recvType, n.Name.Name)
		} else {
			funcName = fmt.Sprintf("%s.%s", v.pkgName, n.Name.Name)
		}
		*v.results = append(*v.results, api.FunctionComplexity{
			Name:       funcName,
			Line:       v.fset.Position(n.Pos()).Line,
			Complexity: complexity,
		})
		return v
	case *ast.FuncLit:
		complexity := 1 + countBranchComplexity(n.Body)
		pos := v.fset.Position(n.Pos())
		*v.results = append(*v.results, api.FunctionComplexity{
			Name:       fmt.Sprintf("%s.literal@%d:%d", v.pkgName, pos.Line, pos.Column),
			Line:       pos.Line,
			Complexity: complexity,
		})
		return v
	}
	return v
}

func countBranchComplexity(body *ast.BlockStmt) int {
	count := 0
	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.FuncLit, *ast.FuncDecl:
			return false
		case *ast.IfStmt:
			count++
		case *ast.ForStmt, *ast.RangeStmt:
			count++
		case *ast.CaseClause:
			count++
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				count++
			}
		case *ast.UnaryExpr:
			if n.Op == token.ARROW {
				count++
			}
		}
		return true
	})
	return count
}

func typeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return typeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return typeName(t.X)
	case *ast.IndexListExpr:
		return typeName(t.X)
	}
	return fmt.Sprintf("%T", expr)
}

func parseSyntaxErrors(filename string, src interface{}, fset *token.FileSet, err error) []api.SyntaxError {
	if err == nil {
		return nil
	}
	var errors []api.SyntaxError
	if list, ok := err.(interface {
		Errors() []error
	}); ok {
		for _, e := range list.Errors() {
			errors = append(errors, parseSingleError(filename, fset, e))
		}
		return errors
	}
	return []api.SyntaxError{parseSingleError(filename, fset, err)}
}

func parseSingleError(filename string, fset *token.FileSet, err error) api.SyntaxError {
	msg := err.Error()
	line := 1
	parts := strings.SplitN(msg, ":", 3)
	if len(parts) >= 2 {
		if l, err2 := strconv.Atoi(parts[1]); err2 == nil {
			line = l
		}
	}
	return api.SyntaxError{
		Filename: filepath.Base(filename),
		Line:     line,
		Message:  msg,
	}
}
