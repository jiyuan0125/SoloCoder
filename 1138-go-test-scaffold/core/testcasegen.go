package core

import (
	"fmt"
)

type TestCaseGenerator struct {
	valueGen *ValueGenerator
}

func NewTestCaseGenerator(valueGen *ValueGenerator) *TestCaseGenerator {
	return &TestCaseGenerator{valueGen: valueGen}
}

func (g *TestCaseGenerator) GenerateStructTestCases(structInfo *StructInfo) []*TestCase {
	normalValue, requiredValue, zeroValue := g.valueGen.GenerateTestValues(structInfo)

	return []*TestCase{
		{
			Name:     "normal",
			Input:    normalValue,
			Expected: "// TODO: fill expected value",
		},
		{
			Name:     "required_only",
			Input:    requiredValue,
			Expected: "// TODO: fill expected value",
		},
		{
			Name:     "zero_value",
			Input:    zeroValue,
			Expected: "// TODO: fill expected value",
		},
	}
}

func (g *TestCaseGenerator) GenerateMethodTestCases(structInfo *StructInfo, method *MethodInfo) []*TestCase {
	normalValue, _, _ := g.valueGen.GenerateTestValues(structInfo)

	return []*TestCase{
		{
			Name:     "normal",
			Input:    normalValue,
			Expected: "// TODO: fill expected result and assertions",
		},
		{
			Name:     "zero_value",
			Input:    fmt.Sprintf("%s{}", structInfo.Name),
			Expected: "// TODO: fill expected result and assertions",
		},
	}
}

func (g *TestCaseGenerator) GenerateInterfaceTestCases(structInfo *StructInfo, ifaceName string) []*TestCase {
	normalValue, _, _ := g.valueGen.GenerateTestValues(structInfo)

	return []*TestCase{
		{
			Name:     "implements_interface",
			Input:    normalValue,
			Expected: "// Interface contract test - should compile if implemented correctly",
		},
	}
}
