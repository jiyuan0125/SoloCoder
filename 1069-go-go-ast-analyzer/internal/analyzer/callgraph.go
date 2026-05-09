package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/example/go-ast-analyzer/internal/api"
)

type callGraphVisitor struct {
	fset        *token.FileSet
	pkgName     string
	adjacency   map[string][]string
	currentFunc string
}

func AnalyzeCallGraph(fset *token.FileSet, file *ast.File) api.CallGraph {
	v := &callGraphVisitor{
		fset:      fset,
		pkgName:   file.Name.Name,
		adjacency: make(map[string][]string),
	}
	ast.Walk(v, file)
	return api.CallGraph{Adjacency: v.adjacency}
}

func (v *callGraphVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			return nil
		}
		var funcName string
		if n.Recv != nil && len(n.Recv.List) > 0 {
			recvType := typeName(n.Recv.List[0].Type)
			funcName = fmt.Sprintf("%s.%s.%s", v.pkgName, recvType, n.Name.Name)
		} else {
			funcName = fmt.Sprintf("%s.%s", v.pkgName, n.Name.Name)
		}
		v.adjacency[funcName] = []string{}
		return &funcBodyVisitor{
			parent:    v,
			ownerFunc: funcName,
		}
	}
	return v
}

type funcBodyVisitor struct {
	parent    *callGraphVisitor
	ownerFunc string
}

func (fv *funcBodyVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.FuncDecl:
		return nil
	case *ast.FuncLit:
		pos := fv.parent.fset.Position(n.Pos())
		litName := fmt.Sprintf("%s.literal@%d:%d", fv.parent.pkgName, pos.Line, pos.Column)
		if _, ok := fv.parent.adjacency[litName]; !ok {
			fv.parent.adjacency[litName] = []string{}
		}
		return &funcBodyVisitor{
			parent:    fv.parent,
			ownerFunc: litName,
		}
	case *ast.CallExpr:
		callee := resolveCallee(fv.parent.pkgName, n.Fun)
		if callee != "" {
			if _, ok := fv.parent.adjacency[fv.ownerFunc]; !ok {
				fv.parent.adjacency[fv.ownerFunc] = []string{}
			}
			fv.parent.adjacency[fv.ownerFunc] = append(fv.parent.adjacency[fv.ownerFunc], callee)
		}
	}
	return fv
}

func resolveCallee(pkgName string, fun ast.Expr) string {
	switch e := fun.(type) {
	case *ast.Ident:
		return fmt.Sprintf("%s.%s", pkgName, e.Name)
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", ident.Name, e.Sel.Name)
		}
	case *ast.ParenExpr:
		return resolveCallee(pkgName, e.X)
	case *ast.IndexExpr:
		return resolveCallee(pkgName, e.X)
	case *ast.IndexListExpr:
		return resolveCallee(pkgName, e.X)
	case *ast.TypeAssertExpr:
		return resolveCallee(pkgName, e.X)
	}
	return ""
}
