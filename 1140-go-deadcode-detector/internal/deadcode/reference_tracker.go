package deadcode

import (
	"go/ast"

	"deadcode-detector/internal/common"
)

type ReferenceMap struct {
	DeclarationLocations map[string][]common.Location
	ReferenceLocations   map[string][]common.ReferenceLocation
	FieldDeclarations    map[string]map[string]common.Location
	FieldReferences      map[string]map[string][]common.ReferenceLocation
	AssignmentLocations  map[string][]common.ReferenceLocation
	ReadLocations        map[string][]common.ReferenceLocation
	AllDeclPositions     map[string]map[string]bool
}

func NewReferenceMap() *ReferenceMap {
	return &ReferenceMap{
		DeclarationLocations: make(map[string][]common.Location),
		ReferenceLocations:   make(map[string][]common.ReferenceLocation),
		FieldDeclarations:    make(map[string]map[string]common.Location),
		FieldReferences:      make(map[string]map[string][]common.ReferenceLocation),
		AssignmentLocations:  make(map[string][]common.ReferenceLocation),
		ReadLocations:        make(map[string][]common.ReferenceLocation),
		AllDeclPositions:     make(map[string]map[string]bool),
	}
}

func TrackReferences(files []*ParsedFile, declarations []common.Declaration) *ReferenceMap {
	rm := NewReferenceMap()

	for _, decl := range declarations {
		if decl.Type == common.TypeField {
			if _, ok := rm.FieldDeclarations[decl.Location.File]; !ok {
				rm.FieldDeclarations[decl.Location.File] = make(map[string]common.Location)
			}
			rm.FieldDeclarations[decl.Location.File][decl.Name] = decl.Location
		} else {
			rm.DeclarationLocations[decl.Name] = append(rm.DeclarationLocations[decl.Name], decl.Location)
		}
		key := fmtKey(decl.Location.File, decl.Location.Line, decl.Name)
		if _, ok := rm.AllDeclPositions[key]; !ok {
			rm.AllDeclPositions[key] = make(map[string]bool)
		}
		rm.AllDeclPositions[key][decl.Name] = true
	}

	for _, file := range files {
		collectAllIdentifiers(file, rm)
	}

	return rm
}

func fmtKey(file string, line int, name string) string {
	return file + ":" + string(rune(line)) + ":" + name
}

func collectAllIdentifiers(file *ParsedFile, rm *ReferenceMap) {
	ast.Inspect(file.AstFile, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch node := n.(type) {
		case *ast.CallExpr:
			switch fun := node.Fun.(type) {
			case *ast.Ident:
				pos := file.FileSet.Position(fun.Pos())
				if !isDeclarationPosition(file, pos.Line, fun.Name, rm) {
					refLoc := common.ReferenceLocation{
						File:     file.Filename,
						Line:     pos.Line,
						Column:   pos.Column,
						LineText: getLineText(file, pos.Line),
					}
					rm.ReferenceLocations[fun.Name] = append(rm.ReferenceLocations[fun.Name], refLoc)
					rm.ReadLocations[fun.Name] = append(rm.ReadLocations[fun.Name], refLoc)
				}
			}

		case *ast.SelectorExpr:
			if node.Sel != nil {
				ident := node.Sel
				pos := file.FileSet.Position(ident.Pos())
				refLoc := common.ReferenceLocation{
					File:     file.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					LineText: getLineText(file, pos.Line),
				}

				if _, ok := rm.FieldReferences[file.Filename]; !ok {
					rm.FieldReferences[file.Filename] = make(map[string][]common.ReferenceLocation)
				}
				rm.FieldReferences[file.Filename][ident.Name] = append(rm.FieldReferences[file.Filename][ident.Name], refLoc)
				rm.ReferenceLocations[ident.Name] = append(rm.ReferenceLocations[ident.Name], refLoc)
				rm.ReadLocations[ident.Name] = append(rm.ReadLocations[ident.Name], refLoc)
			}

		case *ast.Ident:
			if node.Name == "_" {
				return true
			}

			pos := file.FileSet.Position(node.Pos())

			if isDeclarationPosition(file, pos.Line, node.Name, rm) {
				return true
			}

			if isFunctionNameDecl(node) {
				return true
			}

			refLoc := common.ReferenceLocation{
				File:     file.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				LineText: getLineText(file, pos.Line),
			}

			rm.ReferenceLocations[node.Name] = append(rm.ReferenceLocations[node.Name], refLoc)
			rm.ReadLocations[node.Name] = append(rm.ReadLocations[node.Name], refLoc)
		}

		return true
	})
}

func isDeclarationPosition(file *ParsedFile, line int, name string, rm *ReferenceMap) bool {
	key := fmtKey(file.Filename, line, name)
	if names, ok := rm.AllDeclPositions[key]; ok {
		if names[name] {
			return true
		}
	}

	declLocs := rm.DeclarationLocations[name]
	for _, dl := range declLocs {
		if dl.File == file.Filename && dl.Line == line {
			return true
		}
	}
	return false
}

func isFunctionNameDecl(ident *ast.Ident) bool {
	if ident.Obj == nil {
		return false
	}
	return ident.Obj.Kind == ast.Fun && ident.Obj.Decl != nil
}

func getLineText(file *ParsedFile, line int) string {
	if line > 0 && line <= len(file.Lines) {
		return file.Lines[line-1]
	}
	return ""
}

func IsReferenced(name string, rm *ReferenceMap) bool {
	refs := rm.ReferenceLocations[name]
	return len(refs) > 0
}

func IsVariableAssigned(name string, rm *ReferenceMap) bool {
	return len(rm.AssignmentLocations[name]) > 0
}

func IsVariableRead(name string, rm *ReferenceMap) bool {
	return len(rm.ReadLocations[name]) > 0
}

func IsFieldReferenced(filename, fieldName string, rm *ReferenceMap) bool {
	if refs, ok := rm.FieldReferences[filename]; ok {
		if len(refs[fieldName]) > 0 {
			return true
		}
	}

	var fieldDeclLoc common.Location
	var hasFieldDecl bool
	if decls, ok := rm.FieldDeclarations[filename]; ok {
		fieldDeclLoc, hasFieldDecl = decls[fieldName]
	}

	refs := rm.ReferenceLocations[fieldName]
	if len(refs) == 0 {
		return false
	}

	for _, ref := range refs {
		if !hasFieldDecl {
			return true
		}
		if ref.File != fieldDeclLoc.File || ref.Line != fieldDeclLoc.Line {
			return true
		}
	}
	return false
}

func GetReferences(name string, rm *ReferenceMap) []common.ReferenceLocation {
	return rm.ReferenceLocations[name]
}

func GetAssignmentLocations(name string, rm *ReferenceMap) []common.ReferenceLocation {
	return rm.AssignmentLocations[name]
}

func GetReadLocations(name string, rm *ReferenceMap) []common.ReferenceLocation {
	return rm.ReadLocations[name]
}

func GetDeclarationLocation(name string, rm *ReferenceMap) (common.Location, bool) {
	locs, ok := rm.DeclarationLocations[name]
	if !ok || len(locs) == 0 {
		return common.Location{}, false
	}
	return locs[0], true
}
