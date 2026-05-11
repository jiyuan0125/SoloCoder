package buildtag

import (
	"path/filepath"
	"regexp"
	"strings"

	"buildtag/api"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

type FileConstraints struct {
	NameConstraint     Expr
	NewConstraint      Expr
	OldConstraint      Expr
	GenerateDirectives []api.GenerateDirective
	Warnings           []string
}

var fileNamePattern = regexp.MustCompile(`^([a-zA-Z0-9_]+?)_([a-z]+)(?:_([a-z0-9]+))?\.go$`)

func (a *Analyzer) extractFileNameConstraint(fileName string) (Expr, string) {
	base := filepath.Base(fileName)
	if strings.HasSuffix(base, "_test.go") {
		return nil, ""
	}
	if !strings.HasSuffix(base, ".go") {
		return nil, ""
	}
	
	parts := strings.Split(strings.TrimSuffix(base, ".go"), "_")
	if len(parts) < 2 {
		return nil, ""
	}
	
	var result Expr
	constraintDesc := ""
	
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if part == "test" {
			continue
		}
		
		expr := &TagExpr{Tag: strings.ToLower(part)}
		if constraintDesc == "" {
			constraintDesc = part
		} else {
			constraintDesc += " AND " + part
		}
		
		if result == nil {
			result = expr
		} else {
			result = &AndExpr{Left: result, Right: expr}
		}
	}
	
	return result, constraintDesc
}

func (a *Analyzer) extractFromContent(fileName string, content string) *FileConstraints {
	constraints := &FileConstraints{}
	lines := strings.Split(content, "\n")
	
	var newBuildLines []string
	var oldBuildLines []string
	
	inOldSection := false
	inNewSection := false
	seenBlankAfterNew := false
	pastPackage := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "//go:generate") {
			cmd := strings.TrimSpace(strings.TrimPrefix(trimmed, "//go:generate"))
			constraints.GenerateDirectives = append(constraints.GenerateDirectives, api.GenerateDirective{
				FileName: fileName,
				Command:  cmd,
				Position: i + 1,
			})
			continue
		}
		
		if strings.HasPrefix(trimmed, "package") {
			pastPackage = true
			continue
		}
		
		if pastPackage {
			continue
		}
		
		if strings.HasPrefix(trimmed, "//go:build") {
			exprStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "//go:build"))
			newBuildLines = append(newBuildLines, exprStr)
			inNewSection = true
			inOldSection = false
			seenBlankAfterNew = false
			continue
		}
		
		isOldBuild := strings.HasPrefix(trimmed, "// +build") ||
			(trimmed != "" && strings.HasPrefix(trimmed, "//+build"))
		if isOldBuild {
			prefix := "// +build"
			if strings.HasPrefix(trimmed, "//+build") {
				prefix = "//+build"
			}
			exprStr := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			oldBuildLines = append(oldBuildLines, exprStr)
			inOldSection = true
			inNewSection = false
			seenBlankAfterNew = true
			continue
		}
		
		if inNewSection && !seenBlankAfterNew {
			if trimmed == "" {
				seenBlankAfterNew = true
				inNewSection = false
			} else if !strings.HasPrefix(trimmed, "//") {
				inNewSection = false
			}
			continue
		}
		
		if inOldSection {
			if strings.HasPrefix(trimmed, "//") || trimmed == "" {
				continue
			}
			inOldSection = false
		}
	}
	
	if len(newBuildLines) > 0 {
		var result Expr
		for _, line := range newBuildLines {
			p := NewParser(line)
			expr, err := p.Parse()
			if err != nil {
				constraints.Warnings = append(constraints.Warnings, 
					"Failed to parse //go:build constraint: "+err.Error())
				continue
			}
			if expr == nil {
				continue
			}
			if result == nil {
				result = expr
			} else {
				constraints.Warnings = append(constraints.Warnings, 
					"Multiple //go:build lines found, only first used")
				break
			}
		}
		constraints.NewConstraint = result
	}
	
	if len(oldBuildLines) > 0 {
		expr, err := ParseOldLines(oldBuildLines)
		if err != nil {
			constraints.Warnings = append(constraints.Warnings, 
				"Failed to parse // +build constraint: "+err.Error())
		}
		constraints.OldConstraint = expr
	}
	
	return constraints
}

func (a *Analyzer) buildTagMap(target api.TargetPlatform) map[string]bool {
	tags := make(map[string]bool)
	if target.GOOS != "" {
		tags[strings.ToLower(target.GOOS)] = true
	}
	if target.GOARCH != "" {
		tags[strings.ToLower(target.GOARCH)] = true
	}
	for _, tag := range target.Tags {
		tags[strings.ToLower(tag)] = true
	}
	return tags
}

func (a *Analyzer) constraintsEquivalent(e1, e2 Expr) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	
	testSets := generateTestTagSets()
	for _, tags := range testSets {
		if e1.Evaluate(tags) != e2.Evaluate(tags) {
			return false
		}
	}
	return true
}

func generateTestTagSets() []map[string]bool {
	tags := []string{"linux", "windows", "amd64", "386", "cgo"}
	n := len(tags)
	result := make([]map[string]bool, 0, 1<<n)
	for mask := 0; mask < (1 << n); mask++ {
		set := make(map[string]bool)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				set[tags[i]] = true
			}
		}
		result = append(result, set)
	}
	return result
}

func (a *Analyzer) AnalyzeFile(file api.File, target api.TargetPlatform) (*api.FileAnalysis, []api.GenerateDirective, []string) {
	constraints := a.extractFromContent(file.Name, file.Content)
	nameConstraint, nameDesc := a.extractFileNameConstraint(file.Name)
	constraints.NameConstraint = nameConstraint
	
	var allWarnings []string
	allWarnings = append(allWarnings, constraints.Warnings...)
	
	if constraints.NewConstraint != nil && constraints.OldConstraint != nil {
		if !a.constraintsEquivalent(constraints.NewConstraint, constraints.OldConstraint) {
			allWarnings = append(allWarnings, 
				file.Name+": //go:build and // +build constraints are not equivalent")
		}
	}
	
	tagMap := a.buildTagMap(target)
	
	analysis := &api.FileAnalysis{
		FileName: file.Name,
		Included:   true,
		Reason:     "No constraints",
	}
	
	activeConstraint := constraints.NewConstraint
	if activeConstraint == nil {
		activeConstraint = constraints.OldConstraint
	}
	
	if constraints.NameConstraint != nil {
		if !constraints.NameConstraint.Evaluate(tagMap) {
			analysis.Included = false
			analysis.Reason = "File name constraint not satisfied: " + nameDesc
		} else {
			analysis.Reason = "File name constraint satisfied: " + nameDesc
		}
	}
	
	if analysis.Included && activeConstraint != nil {
		if !activeConstraint.Evaluate(tagMap) {
			analysis.Included = false
			if analysis.Reason != "No constraints" {
				analysis.Reason += " AND header constraint not satisfied"
			} else {
				analysis.Reason = "Header constraint not satisfied"
			}
		} else {
			if analysis.Reason == "No constraints" {
				analysis.Reason = "Header constraint satisfied"
			} else {
				analysis.Reason += " AND header constraint satisfied"
			}
		}
	}
	
	return analysis, constraints.GenerateDirectives, allWarnings
}

func (a *Analyzer) Analyze(request api.AnalyzeRequest) api.AnalyzeResponse {
	response := api.AnalyzeResponse{
		IncludedFiles:      make([]api.FileAnalysis, 0),
		ExcludedFiles:      make([]api.FileAnalysis, 0),
		Warnings:           make([]string, 0),
		GenerateDirectives: make([]api.GenerateDirective, 0),
	}
	
	for _, file := range request.Files {
		analysis, directives, warnings := a.AnalyzeFile(file, request.Target)
		response.Warnings = append(response.Warnings, warnings...)
		response.GenerateDirectives = append(response.GenerateDirectives, directives...)
		
		if analysis.Included {
			response.IncludedFiles = append(response.IncludedFiles, *analysis)
		} else {
			response.ExcludedFiles = append(response.ExcludedFiles, *analysis)
		}
	}
	
	return response
}
