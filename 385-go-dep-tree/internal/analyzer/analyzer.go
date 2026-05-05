package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	
	"dep-tree/protocol"
)

func Analyze(projectPath string, filterOpts protocol.FilterOptions) (*protocol.ProjectInfo, error) {
	if err := validateProjectPath(projectPath); err != nil {
		return nil, err
	}

	moduleName, modFile, err := parseGoMod(projectPath)
	if err != nil {
		return nil, err
	}

	pkgImports, err := scanProjectImports(projectPath, moduleName)
	if err != nil {
		return nil, err
	}

	standardLibs := getStandardLibraries()

	depMap := buildDependencyMap(moduleName, pkgImports, modFile, standardLibs)

	root, circularDeps := buildDependencyTree(moduleName, depMap, filterOpts)

	allDeps := collectAllDependencies(root, standardLibs)
	totalDeps := len(allDeps)
	thirdPartyDeps := countThirdParty(allDeps, standardLibs)

	return &protocol.ProjectInfo{
		ModuleName:     moduleName,
		Root:           root,
		CircularDeps:   circularDeps,
		TotalDeps:      totalDeps,
		ThirdPartyDeps: thirdPartyDeps,
	}, nil
}

func validateProjectPath(projectPath string) error {
	info, err := os.Stat(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("error: project path does not exist: %s", projectPath)
		}
		return fmt.Errorf("error: cannot access project path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("error: path is not a directory: %s", projectPath)
	}

	modPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return fmt.Errorf("error: no go.mod file found in project path: %s", projectPath)
	}
	return nil
}

type GoModInfo struct {
	moduleName  string
	require     map[string]string
	indirect    map[string]bool
}

func parseGoMod(projectPath string) (string, *GoModInfo, error) {
	modPath := filepath.Join(projectPath, "go.mod")
	file, err := os.Open(modPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read go.mod: %w", err)
	}
	defer file.Close()

	info := &GoModInfo{
		require:  make(map[string]string),
		indirect: make(map[string]bool),
	}

	var moduleName string
	inRequireBlock := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimPrefix(line, "module ")
			continue
		}

		if line == "require (" {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		if (inRequireBlock || strings.HasPrefix(line, "require ")) && strings.Contains(line, " ") {
			var parts []string
			if inRequireBlock {
				parts = strings.Fields(line)
			} else {
				parts = strings.Fields(strings.TrimPrefix(line, "require "))
			}
			if len(parts) >= 2 {
				pkg := parts[0]
				version := parts[1]
				info.require[pkg] = version
				if len(parts) >= 3 && parts[2] == "//indirect" {
					info.indirect[pkg] = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	return moduleName, info, nil
}

func scanProjectImports(projectPath string, moduleName string) (map[string][]string, error) {
	pkgImports := make(map[string][]string)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		base := filepath.Base(path)
		if strings.HasPrefix(base, "_") {
			return nil
		}
		if strings.HasPrefix(base, ".") {
			return nil
		}

		relPath, _ := filepath.Rel(projectPath, path)
		pkgPath := filepath.Join(moduleName, filepath.Dir(relPath))
		pkgPath = filepath.ToSlash(pkgPath)

		imports, err := extractImports(path)
		if err != nil {
			return err
		}

		for _, imp := range imports {
			if _, exists := pkgImports[pkgPath]; !exists {
				pkgImports[pkgPath] = make([]string, 0)
			}
			pkgImports[pkgPath] = append(pkgImports[pkgPath], imp)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	return pkgImports, nil
}

var importPattern = regexp.MustCompile(`import\s+(?:\(\s*([\s\S]*?)\s*\)|"([^"]+)")`)

func extractImports(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imports []string
	scanner := bufio.NewScanner(file)
	inImportBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(line, "import") {
			if strings.Contains(line, "(") {
				inImportBlock = true
				continue
			}
			matches := importPattern.FindStringSubmatch(line)
			if len(matches) > 2 && matches[2] != "" {
				imp := matches[2]
				if !strings.HasPrefix(imp, "_") && !strings.HasPrefix(imp, ".") {
					imports = append(imports, imp)
				}
			}
			continue
		}

		if inImportBlock {
			if strings.Contains(line, ")") {
				inImportBlock = false
				continue
			}
			
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			
			if strings.HasPrefix(line, "_") || strings.HasPrefix(line, ".") {
				continue
			}
			
			parts := strings.Fields(line)
			if len(parts) > 0 {
				imp := strings.Trim(parts[len(parts)-1], `"`)
				if !strings.HasPrefix(imp, "_") && !strings.HasPrefix(imp, ".") {
					imports = append(imports, imp)
				}
			}
		}
	}

	return imports, scanner.Err()
}

func buildDependencyMap(moduleName string, pkgImports map[string][]string, modFile *GoModInfo, standardLibs map[string]bool) map[string]*PackageInfo {
	depMap := make(map[string]*PackageInfo)

	for pkg, imports := range pkgImports {
		if _, exists := depMap[pkg]; !exists {
			depMap[pkg] = &PackageInfo{
				Name:       pkg,
				IsStandard: false,
				IsLocal:    true,
			}
		}
		
		for _, imp := range imports {
			if _, exists := depMap[imp]; !exists {
				isStandard := standardLibs[imp]
				depMap[imp] = &PackageInfo{
					Name:       imp,
					IsStandard: isStandard,
					IsLocal:    !isStandard,
				}
				
				if !isStandard {
					if version, ok := modFile.require[imp]; ok {
						depMap[imp].Version = version
						depMap[imp].IsIndirect = modFile.indirect[imp]
					}
				}
			}
			
			depMap[pkg].Dependencies = append(depMap[pkg].Dependencies, imp)
		}
	}

	internalPkgs := collectInternalPackages(moduleName, pkgImports, depMap)
	entryPkgs := findEntryPackages(moduleName, internalPkgs, depMap)
	
	depMap[moduleName] = &PackageInfo{
		Name:         moduleName,
		IsStandard:   false,
		IsLocal:      true,
		Dependencies: entryPkgs,
	}

	return depMap
}

func collectInternalPackages(moduleName string, pkgImports map[string][]string, depMap map[string]*PackageInfo) map[string]bool {
	internalPkgs := make(map[string]bool)
	
	for pkg := range pkgImports {
		if pkg == moduleName || strings.HasPrefix(pkg, moduleName+"/") {
			internalPkgs[pkg] = true
		}
	}
	
	for pkg, info := range depMap {
		if info.IsLocal && (pkg == moduleName || strings.HasPrefix(pkg, moduleName+"/")) {
			internalPkgs[pkg] = true
		}
	}
	
	return internalPkgs
}

func findEntryPackages(moduleName string, internalPkgs map[string]bool, depMap map[string]*PackageInfo) []string {
	importedByOthers := make(map[string]bool)
	
	for _, info := range depMap {
		for _, dep := range info.Dependencies {
			if internalPkgs[dep] && dep != moduleName {
				importedByOthers[dep] = true
			}
		}
	}
	
	entryPkgs := make([]string, 0)
	for pkg := range internalPkgs {
		if pkg != moduleName && !importedByOthers[pkg] {
			entryPkgs = append(entryPkgs, pkg)
		}
	}
	
	return entryPkgs
}

type PackageInfo struct {
	Name         string
	Version      string
	IsStandard   bool
	IsLocal      bool
	IsIndirect   bool
	Dependencies []string
}

func buildDependencyTree(moduleName string, depMap map[string]*PackageInfo, filterOpts protocol.FilterOptions) (*protocol.DependencyNode, []protocol.CircularDep) {
	visited := make(map[string]bool)
	seenNodes := make(map[string]bool)
	circularDeps := make([]protocol.CircularDep, 0)
	visitingPath := make([]string, 0)

	root := &protocol.DependencyNode{
		Name:       moduleName,
		IsStandard: false,
		IsRoot:     true,
	}

	var buildTree func(*protocol.DependencyNode, string)
	buildTree = func(node *protocol.DependencyNode, pkg string) {
		if visited[pkg] {
			if seenNodes[pkg] {
				node.IsReused = true
				node.ReusedRef = pkg
			}
			return
		}
		
		for _, pathPkg := range visitingPath {
			if pathPkg == pkg {
				circularPath := make([]string, 0)
				found := false
				for _, p := range visitingPath {
					if p == pkg {
						found = true
					}
					if found {
						circularPath = append(circularPath, p)
					}
				}
				circularPath = append(circularPath, pkg)
				circularDeps = append(circularDeps, protocol.CircularDep{Path: circularPath})
				return
			}
		}

		visited[pkg] = true
		visitingPath = append(visitingPath, pkg)
		seenNodes[pkg] = true

		if pkgInfo, ok := depMap[pkg]; ok {
			node.Version = pkgInfo.Version
			node.IsStandard = pkgInfo.IsStandard
			node.IsIndirect = pkgInfo.IsIndirect

			for _, dep := range pkgInfo.Dependencies {
				if shouldInclude(dep, filterOpts, node) {
					child := &protocol.DependencyNode{
						Name:       dep,
						IsStandard: standardLibraries()[dep],
					}
					buildTree(child, dep)
					node.Children = append(node.Children, child)
				}
			}
		}

		visitingPath = visitingPath[:len(visitingPath)-1]
	}

	buildTree(root, moduleName)
	return root, circularDeps
}

func shouldInclude(dep string, filterOpts protocol.FilterOptions, parent *protocol.DependencyNode) bool {
	isStandard := standardLibraries()[dep]
	
	if filterOpts.FilterStandard && filterOpts.FilterThirdParty {
		return false
	}
	
	if filterOpts.FilterStandard {
		return !isStandard
	}
	if filterOpts.FilterThirdParty {
		return isStandard
	}
	
	if filterOpts.SearchPattern != "" {
		return strings.Contains(dep, filterOpts.SearchPattern)
	}
	
	return true
}

func collectAllDependencies(node *protocol.DependencyNode, standardLibs map[string]bool) map[string]bool {
	allDeps := make(map[string]bool)
	var collect func(*protocol.DependencyNode)
	collect = func(n *protocol.DependencyNode) {
		if !n.IsRoot {
			allDeps[n.Name] = true
		}
		if !n.IsStandard {
			for _, child := range n.Children {
				collect(child)
			}
		}
	}
	collect(node)
	return allDeps
}

func countThirdParty(deps map[string]bool, standardLibs map[string]bool) int {
	count := 0
	for dep := range deps {
		if !standardLibs[dep] {
			count++
		}
	}
	return count
}

var standardLibCache map[string]bool

func standardLibraries() map[string]bool {
	if standardLibCache != nil {
		return standardLibCache
	}
	standardLibCache = getStandardLibraries()
	return standardLibCache
}

func getStandardLibraries() map[string]bool {
	return map[string]bool{
		"archive/tar":      true,
		"archive/zip":      true,
		"bufio":            true,
		"bytes":            true,
		"compress/bzip2":   true,
		"compress/flate":   true,
		"compress/gzip":    true,
		"compress/lzw":     true,
		"compress/zlib":    true,
		"container/heap":   true,
		"container/list":   true,
		"container/ring":   true,
		"context":          true,
		"crypto":           true,
		"crypto/aes":       true,
		"crypto/cipher":    true,
		"crypto/des":       true,
		"crypto/dsa":       true,
		"crypto/ecdsa":     true,
		"crypto/ed25519":   true,
		"crypto/elliptic":  true,
		"crypto/hmac":      true,
		"crypto/md5":       true,
		"crypto/rand":      true,
		"crypto/rc4":       true,
		"crypto/rsa":       true,
		"crypto/sha1":      true,
		"crypto/sha256":    true,
		"crypto/sha512":    true,
		"crypto/subtle":    true,
		"crypto/tls":       true,
		"crypto/x509":      true,
		"crypto/x509/pkix": true,
		"database/sql":     true,
		"database/sql/driver": true,
		"debug/buildinfo":  true,
		"debug/dwarf":      true,
		"debug/elf":        true,
		"debug/gosym":      true,
		"debug/macho":      true,
		"debug/pe":         true,
		"debug/plan9obj":   true,
		"embed":            true,
		"encoding":         true,
		"encoding/ascii85": true,
		"encoding/asn1":    true,
		"encoding/base32":  true,
		"encoding/base64":  true,
		"encoding/binary":  true,
		"encoding/csv":     true,
		"encoding/gob":     true,
		"encoding/hex":     true,
		"encoding/json":    true,
		"encoding/pem":     true,
		"encoding/xml":     true,
		"errors":           true,
		"expvar":           true,
		"flag":             true,
		"fmt":              true,
		"go/ast":           true,
		"go/build":         true,
		"go/build/constraint": true,
		"go/constant":      true,
		"go/doc":           true,
		"go/doc/comment":   true,
		"go/format":        true,
		"go/importer":      true,
		"go/parser":        true,
		"go/printer":       true,
		"go/scanner":       true,
		"go/token":         true,
		"go/types":         true,
		"hash":             true,
		"hash/adler32":     true,
		"hash/crc32":       true,
		"hash/crc64":       true,
		"hash/fnv":         true,
		"hash/maphash":     true,
		"html":             true,
		"html/template":    true,
		"image":            true,
		"image/color":      true,
		"image/color/palette": true,
		"image/draw":       true,
		"image/gif":        true,
		"image/jpeg":       true,
		"image/png":        true,
		"index/suffixarray": true,
		"io":               true,
		"io/fs":            true,
		"io/ioutil":        true,
		"log":              true,
		"log/slog":         true,
		"log/syslog":       true,
		"math":             true,
		"math/big":         true,
		"math/bits":        true,
		"math/cmplx":        true,
		"math/rand":        true,
		"mime":             true,
		"mime/multipart":   true,
		"mime/quotedprintable": true,
		"net":              true,
		"net/http":         true,
		"net/http/cgi":     true,
		"net/http/cookiejar": true,
		"net/http/fcgi":    true,
		"net/http/httptest": true,
		"net/http/httptrace": true,
		"net/http/httputil": true,
		"net/http/pprof":   true,
		"net/mail":         true,
		"net/netip":        true,
		"net/rpc":          true,
		"net/rpc/jsonrpc":  true,
		"net/smtp":         true,
		"net/textproto":    true,
		"net/url":          true,
		"os":               true,
		"os/exec":          true,
		"os/signal":        true,
		"os/user":          true,
		"path":             true,
		"path/filepath":    true,
		"plugin":           true,
		"reflect":          true,
		"regexp":           true,
		"regexp/syntax":    true,
		"runtime":          true,
		"runtime/cgo":      true,
		"runtime/debug":    true,
		"runtime/metrics":  true,
		"runtime/pprof":    true,
		"runtime/race":     true,
		"runtime/trace":    true,
		"sort":             true,
		"strconv":          true,
		"strings":          true,
		"sync":             true,
		"sync/atomic":      true,
		"syscall":          true,
		"testing":          true,
		"testing/fstest":   true,
		"testing/iotest":   true,
		"testing/quick":    true,
		"testing/slogtest": true,
		"text/scanner":     true,
		"text/tabwriter":   true,
		"text/template":    true,
		"text/template/parse": true,
		"time":             true,
		"unicode":          true,
		"unicode/utf16":    true,
		"unicode/utf8":     true,
		"unsafe":           true,
	}
}
