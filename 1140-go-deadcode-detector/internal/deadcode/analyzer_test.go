package deadcode

import (
	"testing"
)

func TestDeadCodeDetection(t *testing.T) {
	testCode := `
package testdata

import "fmt"

type MyInterface interface {
	DoSomething() string
}

type MyStruct struct {
	Name        string
	UnusedField int
}

func (m *MyStruct) DoSomething() string {
	return m.Name
}

type UnusedType struct {
	Value int
}

type UsedAlias = MyStruct
type UnusedAlias = UnusedType

const (
	UsedConst   = "used"
	UnusedConst = "unused"
)

var (
	UsedVar     = "used"
	UnusedVar   = "unused"
	AssignedVar = ""
)

func init() {
	fmt.Println("init")
}

func UsedFunc() {
	fmt.Println(UsedConst, UsedVar)
	AssignedVar = "assigned"
}

func unusedFunc() {
	fmt.Println("unused")
}

func ExportedUnusedFunc() {
	fmt.Println("exported but unused in package")
}

func main() {
	UsedFunc()
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "testdata")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	t.Logf("Package: %s", report.PackageName)
	t.Logf("Summary: %s", report.Summary.String())

	for _, entry := range report.Entries {
		t.Logf("  - %s [%s] (%s) (%s)", entry.Declaration.Name, entry.Declaration.Type, entry.Confidence, entry.Reason)
	}

	foundUnusedFunc := false
	foundMain := false
	for _, entry := range report.Entries {
		if entry.Declaration.Name == "unusedFunc" && entry.Declaration.Type == "function" {
			foundUnusedFunc = true
			if entry.Confidence != "certain" {
				t.Errorf("Expected 'certain' confidence for unusedFunc, got %s", entry.Confidence)
			}
		}
		if entry.Declaration.Name == "main" && entry.Declaration.Type == "function" {
			foundMain = true
		}
	}

	if !foundUnusedFunc {
		t.Error("Expected unusedFunc to be detected as dead code")
	}
	if !foundMain {
		t.Error("Expected main to be detected as unused in non-main package")
	}
}

func TestSpecialCases(t *testing.T) {
	testCode := `
package main

func init() {
}

func main() {
}

//go:linkname externalFunc syscall.externalFunc
func externalFunc() {
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "main")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	if len(report.Entries) != 0 {
		t.Errorf("Expected 0 dead code entries for special cases, got %d", len(report.Entries))
		for _, entry := range report.Entries {
			t.Logf("Unexpected: %s [%s]", entry.Declaration.Name, entry.Declaration.Type)
		}
	}
}

func TestBlankIdentifier(t *testing.T) {
	testCode := `
package test

var _ = "blank"
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "test")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()

	if len(result.Report.Entries) != 0 {
		t.Errorf("Blank identifier should not be reported")
	}
}

func TestUnusedFields(t *testing.T) {
	testCode := `
package test

type MyStruct struct {
	Used   string
	Unused int
}

func NewMyStruct() *MyStruct {
	return &MyStruct{Used: "test"}
}

func (m *MyStruct) GetUsed() string {
	return m.Used
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "test")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	t.Logf("Summary: %s", report.Summary.String())

	foundUnusedField := false
	foundUsedFieldAsDead := false
	for _, entry := range report.Entries {
		t.Logf("  - %s [%s]", entry.Declaration.Name, entry.Declaration.Type)
		if entry.Declaration.Name == "Unused" && entry.Declaration.Type == "field" {
			foundUnusedField = true
		}
		if entry.Declaration.Name == "Used" && entry.Declaration.Type == "field" {
			foundUsedFieldAsDead = true
		}
	}

	if !foundUnusedField {
		t.Error("Expected 'Unused' field to be detected as dead code")
	}
	if foundUsedFieldAsDead {
		t.Error("'Used' field should NOT be marked as dead code")
	}
}

func TestFieldInStructLiteral(t *testing.T) {
	testCode := `
package test

type MyStruct struct {
	UsedField   int
	UnusedField string
}

func useThings() { 
	_ = MyStruct{UsedField: 1} 
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "test")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	t.Logf("Summary: %s", report.Summary.String())

	foundUnusedField := false
	for _, entry := range report.Entries {
		t.Logf("  - %s [%s]", entry.Declaration.Name, entry.Declaration.Type)
		if entry.Declaration.Name == "UnusedField" && entry.Declaration.Type == "field" {
			foundUnusedField = true
		}
		if entry.Declaration.Name == "UsedField" && entry.Declaration.Type == "field" {
			t.Error("UsedField should NOT be marked as dead code (used in struct literal)")
		}
	}

	if !foundUnusedField {
		t.Error("Expected 'UnusedField' to be detected as dead code")
	}
}

func TestConfidenceLevels(t *testing.T) {
	testCode := `
package test

func unexportedUnused() {}

func ExportedUnused() {}

func used() {
}

func main() {
	used()
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "test")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	for _, entry := range report.Entries {
		t.Logf("  - %s [%s]", entry.Declaration.Name, entry.Confidence)
		if entry.Declaration.Name == "unexportedUnused" {
			if entry.Confidence != "certain" {
				t.Errorf("Unexported should have 'certain' confidence")
			}
		}
		if entry.Declaration.Name == "ExportedUnused" {
			if entry.Confidence != "possible" {
				t.Errorf("Exported should have 'possible' confidence")
			}
		}
	}
}

func TestInterfaceMethodNotMarkedAsDead(t *testing.T) {
	testCode := `
package test

type MyInterface interface {
	DoSomething() string
}

type MyStruct struct {
	Name string
}

func (m *MyStruct) DoSomething() string {
	return m.Name
}

func main() {
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "test")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	t.Logf("Summary: %s", report.Summary.String())
	for _, entry := range report.Entries {
		t.Logf("  - %s [%s] (%s)", entry.Declaration.Name, entry.Declaration.Type, entry.Reason)
	}

	for _, entry := range report.Entries {
		if entry.Declaration.Name == "DoSomething" && entry.Declaration.Type == "function" {
			t.Error("DoSomething (interface method) should NOT be marked as dead code")
		}
	}
}

func TestInterfaceMethodMultipleParams(t *testing.T) {
	testCode := `
package test

type Processor interface {
	Process(a, b int, name string) (string, error)
}

type MyProcessor struct{}

func (p *MyProcessor) Process(x, y int, n string) (string, error) {
	return "", nil
}

func main() {
}
`

	analyzer, err := NewAnalyzer(map[string]string{"test.go": testCode}, "test")
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	result := analyzer.Analyze()
	report := result.Report

	t.Logf("Summary: %s", report.Summary.String())
	for _, entry := range report.Entries {
		t.Logf("  - %s [%s] (%s)", entry.Declaration.Name, entry.Declaration.Type, entry.Reason)
	}

	for _, entry := range report.Entries {
		if entry.Declaration.Name == "Process" && entry.Declaration.Type == "function" {
			t.Error("Process (interface method with multiple params) should NOT be marked as dead code")
		}
	}
}
