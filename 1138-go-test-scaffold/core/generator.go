package core

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

const testTemplate = `package {{.PackageName}}

import (
	"testing"
)
{{range .TestFunctions}}
func {{.Name}}(t *testing.T) {
	// TODO: Add more test cases as needed
	testCases := []struct {
		name  string
		input {{.Receiver}}
	}{
{{range .Cases}}		{
			name:  "{{.Name}}",
			input: {{.Input}},
		},
{{end}}	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: Implement test logic and assertions
			_ = tc.input
			// TODO: Define expected values and add assertions
		})
	}
}
{{end}}`

type TemplateData struct {
	PackageName   string
	TestFunctions []*TestFunction
}

type Generator struct {
	testCaseGen *TestCaseGenerator
}

func NewGenerator(testCaseGen *TestCaseGenerator) *Generator {
	return &Generator{testCaseGen: testCaseGen}
}

func (g *Generator) GenerateTestFile(parsed *ParsedSource, originalFileName string) (*GeneratedTest, error) {
	var testFunctions []*TestFunction

	structNames := make([]string, 0, len(parsed.Structs))
	for name := range parsed.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	for _, name := range structNames {
		structInfo := parsed.Structs[name]
		if name[0] < 'A' || name[0] > 'Z' {
			continue
		}

		testFunctions = append(testFunctions, g.generateStructTest(structInfo))

		for _, method := range structInfo.Methods {
			if method.IsExported {
				testFunctions = append(testFunctions, g.generateMethodTest(structInfo, method))
			}
		}

		for _, ifaceName := range structInfo.Interfaces {
			testFunctions = append(testFunctions, g.generateInterfaceTest(structInfo, ifaceName))
		}
	}

	data := TemplateData{
		PackageName:   parsed.PackageName,
		TestFunctions: testFunctions,
	}

	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	testFileName := strings.TrimSuffix(originalFileName, ".go") + "_test.go"

	return &GeneratedTest{
		FileName: testFileName,
		Code:     buf.String(),
	}, nil
}

func (g *Generator) generateStructTest(structInfo *StructInfo) *TestFunction {
	cases := g.testCaseGen.GenerateStructTestCases(structInfo)

	return &TestFunction{
		Name:     fmt.Sprintf("Test%s", structInfo.Name),
		Cases:    cases,
		IsMethod: false,
		Receiver: structInfo.Name,
	}
}

func (g *Generator) generateMethodTest(structInfo *StructInfo, method *MethodInfo) *TestFunction {
	cases := g.testCaseGen.GenerateMethodTestCases(structInfo, method)

	return &TestFunction{
		Name:     fmt.Sprintf("Test%s_%s", structInfo.Name, method.Name),
		Cases:    cases,
		IsMethod: true,
		Receiver: structInfo.Name,
	}
}

func (g *Generator) generateInterfaceTest(structInfo *StructInfo, ifaceName string) *TestFunction {
	cases := g.testCaseGen.GenerateInterfaceTestCases(structInfo, ifaceName)

	return &TestFunction{
		Name:     fmt.Sprintf("Test%s_Implements%s", structInfo.Name, ifaceName),
		Cases:    cases,
		IsMethod: false,
		Receiver: structInfo.Name,
	}
}

func (g *Generator) GeneratePreview(parsed *ParsedSource, originalFileName string) (*Preview, error) {
	var testFunctions []*TestFunction

	structNames := make([]string, 0, len(parsed.Structs))
	for name := range parsed.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	for _, name := range structNames {
		structInfo := parsed.Structs[name]
		if name[0] < 'A' || name[0] > 'Z' {
			continue
		}

		testFunctions = append(testFunctions, g.generateStructTest(structInfo))

		for _, method := range structInfo.Methods {
			if method.IsExported {
				testFunctions = append(testFunctions, g.generateMethodTest(structInfo, method))
			}
		}

		for _, ifaceName := range structInfo.Interfaces {
			testFunctions = append(testFunctions, g.generateInterfaceTest(structInfo, ifaceName))
		}
	}

	testFileName := strings.TrimSuffix(originalFileName, ".go") + "_test.go"

	return &Preview{
		FileName:      testFileName,
		TestFunctions: testFunctions,
		Structs:       structNames,
	}, nil
}
