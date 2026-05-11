package core

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type ValueGenerator struct {
	rng          *rand.Rand
	structs      map[string]*StructInfo
	depthLimit   int
	currentDepth int
}

func NewValueGenerator(structs map[string]*StructInfo) *ValueGenerator {
	return &ValueGenerator{
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		structs:    structs,
		depthLimit: 3,
	}
}

func (g *ValueGenerator) GenerateTestValues(structInfo *StructInfo) (normalValue string, requiredValue string, zeroValue string) {
	normalValue = g.generateStructValue(structInfo, false, 0)
	requiredValue = g.generateStructValue(structInfo, true, 0)
	zeroValue = fmt.Sprintf("%s{}", structInfo.Name)
	return
}

func (g *ValueGenerator) generateStructValue(structInfo *StructInfo, requiredOnly bool, depth int) string {
	if depth >= g.depthLimit {
		return fmt.Sprintf("%s{}", structInfo.Name)
	}

	var fields []string
	for i, field := range structInfo.Fields {
		if !field.IsExported {
			continue
		}

		if requiredOnly && i > 0 {
			continue
		}

		fieldValue := g.generateFieldValue(field, depth+1)
		if field.IsEmbedded {
			fields = append(fields, fieldValue)
		} else {
			fieldName := field.Name
			fields = append(fields, fmt.Sprintf("%s: %s", fieldName, fieldValue))
		}
	}

	return fmt.Sprintf("%s{%s}", structInfo.Name, strings.Join(fields, ", "))
}

func (g *ValueGenerator) generateFieldValue(field *FieldInfo, depth int) string {
	fieldType := field.Type

	if strings.HasPrefix(fieldType, "*") {
		baseType := strings.TrimPrefix(fieldType, "*")
		if structInfo, exists := g.structs[baseType]; exists {
			return fmt.Sprintf("&%s", g.generateStructValue(structInfo, false, depth))
		}
		return fmt.Sprintf("&%s", g.generatePrimitiveValue(baseType, field.Name))
	}

	if structInfo, exists := g.structs[fieldType]; exists {
		return g.generateStructValue(structInfo, false, depth)
	}

	if strings.HasPrefix(fieldType, "[]") {
		elemType := strings.TrimPrefix(fieldType, "[]")
		var elem1, elem2 string
		if structInfo, exists := g.structs[elemType]; exists {
			elem1 = g.generateStructValue(structInfo, false, depth+1)
			elem2 = g.generateStructValue(structInfo, false, depth+1)
		} else {
			elem1 = g.generatePrimitiveValue(elemType, field.Name)
			elem2 = g.generatePrimitiveValue(elemType, field.Name)
		}
		return fmt.Sprintf("%s{%s, %s}", fieldType, elem1, elem2)
	}

	if strings.HasPrefix(fieldType, "map[") {
		keyType, valueType := parseMapType(fieldType)
		var key, value string
		if structInfo, exists := g.structs[valueType]; exists {
			key = g.generatePrimitiveValue(keyType, field.Name)
			value = g.generateStructValue(structInfo, false, depth+1)
		} else {
			key = g.generatePrimitiveValue(keyType, field.Name)
			value = g.generatePrimitiveValue(valueType, field.Name)
		}
		return fmt.Sprintf("%s{%s: %s}", fieldType, key, value)
	}

	return g.generatePrimitiveValue(fieldType, field.Name)
}

func (g *ValueGenerator) generatePrimitiveValue(fieldType string, fieldName string) string {
	switch fieldType {
	case "string":
		return fmt.Sprintf(`"test_%s"`, fieldName)
	case "int", "int8", "int16", "int32", "int64":
		return fmt.Sprintf("%d", g.rng.Intn(1000)+1)
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return fmt.Sprintf("%d", g.rng.Intn(1000)+1)
	case "float32", "float64":
		return fmt.Sprintf("%.2f", float64(g.rng.Intn(1000))+0.5)
	case "bool":
		return "true"
	case "byte":
		return fmt.Sprintf("%d", g.rng.Intn(256))
	case "rune":
		return fmt.Sprintf("'%c'", g.rng.Intn(26)+97)
	case "complex64", "complex128":
		return fmt.Sprintf("%d + %di", g.rng.Intn(100), g.rng.Intn(100))
	default:
		return fmt.Sprintf("%s{}", fieldType)
	}
}

func parseMapType(mapType string) (keyType string, valueType string) {
	mapType = strings.TrimPrefix(mapType, "map[")
	bracketIndex := findMatchingBracket(mapType)
	if bracketIndex == -1 {
		return "", ""
	}
	keyType = mapType[:bracketIndex]
	valueType = mapType[bracketIndex+1:]
	return
}

func findMatchingBracket(s string) int {
	depth := 1
	for i, c := range s {
		switch c {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
