package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"

	"github.com/example/go-ast-analyzer/internal/api"
)

type varInfo struct {
	name   string
	line   int
	used   bool
	assigned bool
}

type unusedVisitor struct {
	filename string
	fset     *token.FileSet
	vars     map[string]*varInfo
	imports  map[string]string
	importLines map[string]int
	usedImports map[string]bool
	file     *ast.File
	scope    []string
}

func AnalyzeUnused(filename string, fset *token.FileSet, file *ast.File) ([]api.UnusedVariable, []api.UnusedImport) {
	v := &unusedVisitor{
		filename:    filepath.Base(filename),
		fset:        fset,
		vars:        make(map[string]*varInfo),
		imports:     make(map[string]string),
		importLines: make(map[string]int),
		usedImports: make(map[string]bool),
		file:        file,
	}
	
	for _, imp := range file.Imports {
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			name = filepath.Base(imp.Path.Value[1 : len(imp.Path.Value)-1])
		}
		path := imp.Path.Value[1 : len(imp.Path.Value)-1]
		line := fset.Position(imp.Pos()).Line
		if name != "_" {
			v.imports[name] = path
			v.importLines[name] = line
		}
	}
	
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.GenDecl:
			if n.Tok == token.VAR {
				for _, spec := range n.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range vs.Names {
							if id.Name != "_" {
								v.vars[id.Name] = &varInfo{
									name: id.Name,
									line: fset.Position(id.Pos()).Line,
								}
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			return v.visitFuncDecl(n)
		case *ast.AssignStmt:
			v.visitAssign(n)
		case *ast.Ident:
			v.visitIdent(n, false)
		case *ast.SelectorExpr:
			v.visitSelector(n)
		}
		return true
	})
	
	var unusedVars []api.UnusedVariable
	for name, info := range v.vars {
		if !info.used {
			unusedVars = append(unusedVars, api.UnusedVariable{
				Filename: v.filename,
				Line:     info.line,
				Name:     name,
			})
		}
	}
	
	var unusedImports []api.UnusedImport
	for name, path := range v.imports {
		if !v.usedImports[name] {
			unusedImports = append(unusedImports, api.UnusedImport{
				Filename: v.filename,
				Line:     v.importLines[name],
				Path:     path,
			})
		}
	}
	
	return unusedVars, unusedImports
}

func (v *unusedVisitor) visitFuncDecl(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	localVars := make(map[string]*varInfo)
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.FuncLit:
			for _, field := range n.Type.Params.List {
				for _, id := range field.Names {
					if id.Name != "_" {
						localVars[id.Name] = &varInfo{
							name: id.Name,
							line: v.fset.Position(id.Pos()).Line,
						}
					}
				}
			}
			ast.Inspect(n.Body, func(litNode ast.Node) bool {
				if litNode == nil {
					return true
				}
				switch litN := litNode.(type) {
				case *ast.AssignStmt:
					v.visitLocalAssign(litN, localVars)
				case *ast.Ident:
					v.visitLocalIdent(litN, false, localVars)
				}
				return true
			})
			for name, info := range localVars {
				if !info.used {
					v.vars[name] = info
				}
			}
			return false
		case *ast.AssignStmt:
			v.visitLocalAssign(n, localVars)
		case *ast.Ident:
			v.visitLocalIdent(n, false, localVars)
		case *ast.SelectorExpr:
			v.visitSelector(n)
		}
		return true
	})
	for name, info := range localVars {
		if !info.used {
			v.vars[name] = info
		}
	}
	return false
}

func (v *unusedVisitor) visitAssign(n *ast.AssignStmt) {
	for _, lhs := range n.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
			if n.Tok == token.DEFINE {
				if _, exists := v.vars[id.Name]; !exists {
					v.vars[id.Name] = &varInfo{
						name: id.Name,
						line: v.fset.Position(id.Pos()).Line,
					}
				}
			}
			if info, ok := v.vars[id.Name]; ok {
				info.assigned = true
			}
		}
	}
	for _, rhs := range n.Rhs {
		ast.Inspect(rhs, func(node ast.Node) bool {
			if id, ok := node.(*ast.Ident); ok && id.Name != "_" {
				v.visitIdent(id, true)
			}
			return true
		})
	}
}

func (v *unusedVisitor) visitLocalAssign(n *ast.AssignStmt, locals map[string]*varInfo) {
	for _, lhs := range n.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
			if n.Tok == token.DEFINE {
				if _, exists := locals[id.Name]; !exists {
					locals[id.Name] = &varInfo{
						name: id.Name,
						line: v.fset.Position(id.Pos()).Line,
					}
				}
			}
			if info, ok := locals[id.Name]; ok {
				info.assigned = true
			} else if info, ok := v.vars[id.Name]; ok {
				info.assigned = true
			}
		}
	}
	for _, rhs := range n.Rhs {
		ast.Inspect(rhs, func(node ast.Node) bool {
			if id, ok := node.(*ast.Ident); ok && id.Name != "_" {
				v.visitLocalIdent(id, true, locals)
			}
			return true
		})
	}
}

func (v *unusedVisitor) visitIdent(n *ast.Ident, forceUse bool) {
	if n.Name == "_" {
		return
	}
	if info, ok := v.vars[n.Name]; ok {
		info.used = true
	}
	if _, ok := v.imports[n.Name]; ok {
		v.usedImports[n.Name] = true
	}
}

func (v *unusedVisitor) visitLocalIdent(n *ast.Ident, forceUse bool, locals map[string]*varInfo) {
	if n.Name == "_" {
		return
	}
	if info, ok := locals[n.Name]; ok {
		info.used = true
	} else if info, ok := v.vars[n.Name]; ok {
		info.used = true
	}
	if _, ok := v.imports[n.Name]; ok {
		v.usedImports[n.Name] = true
	}
}

func (v *unusedVisitor) visitSelector(n *ast.SelectorExpr) {
	if id, ok := n.X.(*ast.Ident); ok {
		if _, ok := v.imports[id.Name]; ok {
			v.usedImports[id.Name] = true
		}
	}
}
