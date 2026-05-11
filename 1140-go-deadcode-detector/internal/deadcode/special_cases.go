package deadcode

import (
	"go/ast"
	"strings"

	"deadcode-detector/internal/common"
)

type SpecialCaseHandler struct {
	isMainPackage  bool
	interfaceInfos []InterfaceInfo
	hasLinkname    map[string]bool
}

func NewSpecialCaseHandler(files []*ParsedFile) *SpecialCaseHandler {
	handler := &SpecialCaseHandler{
		isMainPackage:  IsMainPackage(files),
		interfaceInfos: ExtractInterfaces(files),
		hasLinkname:    make(map[string]bool),
	}

	handler.collectLinknameDirectives(files)

	return handler
}

func (h *SpecialCaseHandler) collectLinknameDirectives(files []*ParsedFile) {
	for _, file := range files {
		for _, decl := range file.AstFile.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				if hasGoLinknameComment(fd, file) {
					h.hasLinkname[fd.Name.Name] = true
				}
			}
		}
	}
}

func hasGoLinknameComment(fd *ast.FuncDecl, file *ParsedFile) bool {
	if fd.Doc != nil {
		for _, comment := range fd.Doc.List {
			if strings.Contains(comment.Text, "//go:linkname") {
				return true
			}
		}
	}
	return false
}

func (h *SpecialCaseHandler) IsExcluded(decl common.Declaration, files []*ParsedFile) bool {
	if decl.Type != common.TypeFunction {
		return false
	}

	if decl.Name == "init" {
		return true
	}

	if h.isMainPackage && decl.Name == "main" {
		return true
	}

	if h.hasLinkname[decl.Name] {
		return true
	}

	if h.ImplementsInterfaceMethod(decl, files) {
		return true
	}

	return false
}

func (h *SpecialCaseHandler) ImplementsInterfaceMethod(decl common.Declaration, files []*ParsedFile) bool {
	if decl.Type != common.TypeFunction {
		return false
	}

	var funcDecl *ast.FuncDecl
	for _, file := range files {
		for _, d := range file.AstFile.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if fd.Name.Name == decl.Name && fd.Pos() != 0 {
					pos := file.FileSet.Position(fd.Pos())
					if pos.Line == decl.Location.Line && pos.Column == decl.Location.Column {
						funcDecl = fd
						break
					}
				}
			}
		}
		if funcDecl != nil {
			break
		}
	}

	if funcDecl == nil {
		return false
	}

	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}

	methodName := funcDecl.Name.Name
	numParams := countParams(funcDecl.Type.Params)
	numResults := countResults(funcDecl.Type.Results)

	for _, iface := range h.interfaceInfos {
		for _, method := range iface.Methods {
			if method.Name == methodName &&
				method.NumParams == numParams &&
				method.NumResults == numResults {
				return true
			}
		}
	}

	return false
}

func countParams(params *ast.FieldList) int {
	if params == nil {
		return 0
	}
	count := 0
	for _, p := range params.List {
		if len(p.Names) > 0 {
			count += len(p.Names)
		} else {
			count += 1
		}
	}
	return count
}

func countResults(results *ast.FieldList) int {
	if results == nil {
		return 0
	}
	count := 0
	for _, r := range results.List {
		if len(r.Names) > 0 {
			count += len(r.Names)
		} else {
			count += 1
		}
	}
	return count
}

func (h *SpecialCaseHandler) IsInterfaceMethodName(name string) bool {
	for _, iface := range h.interfaceInfos {
		for _, method := range iface.Methods {
			if method.Name == name {
				return true
			}
		}
	}
	return false
}

func (h *SpecialCaseHandler) GetInterfaceMethodNames() map[string]bool {
	names := make(map[string]bool)
	for _, iface := range h.interfaceInfos {
		for _, method := range iface.Methods {
			names[method.Name] = true
		}
	}
	return names
}

func IsBlankIdentifier(name string) bool {
	return name == "_"
}

func IsTestFunction(name string) bool {
	return strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Example")
}
