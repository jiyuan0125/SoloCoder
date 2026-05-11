package generator

import (
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/toml"
)

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true,
	"continue": true, "default": true, "defer": true, "else": true,
	"fallthrough": true, "for": true, "func": true, "go": true,
	"goto": true, "if": true, "import": true, "interface": true,
	"map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true,
	"var": true,
}

type structDef struct {
	name   string
	fields []fieldDef
}

type fieldDef struct {
	goName       string
	originalName string
	goType       string
	isArray      bool
}

type Generator struct {
	structs     []*structDef
	structMap   map[string]*structDef
	packageName string
}

func NewGenerator(packageName string) *Generator {
	if packageName == "" {
		packageName = "main"
	}
	return &Generator{
		structs:     []*structDef{},
		structMap:   make(map[string]*structDef),
		packageName: packageName,
	}
}

func (g *Generator) Generate(tomlContent string) (string, error) {
	var data map[string]interface{}
	if _, err := toml.Decode(tomlContent, &data); err != nil {
		return "", err
	}

	rootStruct := &structDef{
		name:   "Config",
		fields: []fieldDef{},
	}
	g.structs = append(g.structs, rootStruct)
	g.structMap["Config"] = rootStruct

	for _, key := range sortedKeys(data) {
		value := data[key]
		field := g.processValue(key, value, rootStruct)
		rootStruct.fields = append(rootStruct.fields, field)
	}

	return g.render(), nil
}

func (g *Generator) processValue(key string, value interface{}, parent *structDef) fieldDef {
	goName := toGoIdentifier(key, true)
	originalName := key

	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Int64:
		return fieldDef{goName: goName, originalName: originalName, goType: "int64", isArray: false}
	case reflect.Float64:
		return fieldDef{goName: goName, originalName: originalName, goType: "float64", isArray: false}
	case reflect.Bool:
		return fieldDef{goName: goName, originalName: originalName, goType: "bool", isArray: false}
	case reflect.String:
		return fieldDef{goName: goName, originalName: originalName, goType: "string", isArray: false}
	case reflect.Struct:
		if _, ok := value.(time.Time); ok {
			return fieldDef{goName: goName, originalName: originalName, goType: "time.Time", isArray: false}
		}
		return fieldDef{goName: goName, originalName: originalName, goType: "interface{}", isArray: false}
	case reflect.Map:
		if m, ok := value.(map[string]interface{}); ok {
			structName := toGoIdentifier(key, true)
			if parent.name != "Config" {
				structName = parent.name + structName
			}
			newStruct := g.getOrCreateStruct(structName)
			for _, k := range sortedKeys(m) {
				field := g.processValue(k, m[k], newStruct)
				newStruct.fields = append(newStruct.fields, field)
			}
			return fieldDef{goName: goName, originalName: originalName, goType: structName, isArray: false}
		}
		return fieldDef{goName: goName, originalName: originalName, goType: "interface{}", isArray: false}
	case reflect.Slice:
		if rv.Len() == 0 {
			return fieldDef{goName: goName, originalName: originalName, goType: "[]interface{}", isArray: true}
		}

		firstElem := rv.Index(0).Interface()
		elemType := g.inferArrayElementType(value)

		if elemType == "interface{}" {
			return fieldDef{goName: goName, originalName: originalName, goType: "[]interface{}", isArray: true}
		}

		if _, ok := firstElem.(map[string]interface{}); ok {
			structName := toGoIdentifier(key, true)
			if parent.name != "Config" {
				structName = parent.name + structName
			}
			newStruct := g.getOrCreateStruct(structName)

			for i := 0; i < rv.Len(); i++ {
				elem := rv.Index(i).Interface()
				if elemMap, ok := elem.(map[string]interface{}); ok {
					if i == 0 {
						for _, k := range sortedKeys(elemMap) {
							field := g.processValue(k, elemMap[k], newStruct)
							newStruct.fields = append(newStruct.fields, field)
						}
					} else {
						for _, k := range sortedKeys(elemMap) {
							field := g.processValue(k, elemMap[k], newStruct)
							found := false
							for _, f := range newStruct.fields {
								if f.goName == field.goName {
									found = true
									break
								}
							}
							if !found {
								newStruct.fields = append(newStruct.fields, field)
							}
						}
					}
				}
			}
			return fieldDef{goName: goName, originalName: originalName, goType: "[]" + structName, isArray: true}
		}

		return fieldDef{goName: goName, originalName: originalName, goType: "[]" + elemType, isArray: true}
	default:
		return fieldDef{goName: goName, originalName: originalName, goType: "interface{}", isArray: false}
	}
}

func (g *Generator) inferArrayElementType(arr interface{}) string {
	rv := reflect.ValueOf(arr)
	if rv.Len() == 0 {
		return "interface{}"
	}

	firstElem := rv.Index(0).Interface()
	firstType := reflect.TypeOf(firstElem)
	allSame := true

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if reflect.TypeOf(elem) != firstType {
			allSame = false
			break
		}
	}

	if !allSame {
		return "interface{}"
	}

	switch firstElem.(type) {
	case int64:
		return "int64"
	case float64:
		return "float64"
	case bool:
		return "bool"
	case string:
		return "string"
	case time.Time:
		return "time.Time"
	case map[string]interface{}:
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

func (g *Generator) getOrCreateStruct(name string) *structDef {
	if s, ok := g.structMap[name]; ok {
		return s
	}
	s := &structDef{
		name:   name,
		fields: []fieldDef{},
	}
	g.structs = append(g.structs, s)
	g.structMap[name] = s
	return s
}

func (g *Generator) render() string {
	var sb strings.Builder

	sb.WriteString("package ")
	sb.WriteString(g.packageName)
	sb.WriteString("\n\n")

	needsTimeImport := false
	for _, s := range g.structs {
		for _, f := range s.fields {
			if f.goType == "time.Time" || strings.Contains(f.goType, "time.Time") {
				needsTimeImport = true
				break
			}
		}
		if needsTimeImport {
			break
		}
	}

	if needsTimeImport {
		sb.WriteString("import \"time\"\n\n")
	}

	for _, s := range g.structs {
		sb.WriteString("type ")
		sb.WriteString(s.name)
		sb.WriteString(" struct {\n")
		for _, f := range s.fields {
			sb.WriteString("\t")
			sb.WriteString(f.goName)
			sb.WriteString(" ")
			sb.WriteString(f.goType)
			sb.WriteString(" `toml:\"")
			sb.WriteString(f.originalName)
			sb.WriteString("\"`\n")
		}
		sb.WriteString("}\n\n")
	}

	return strings.TrimSpace(sb.String()) + "\n"
}

func toGoIdentifier(key string, exported bool) string {
	if key == "" {
		return "X"
	}

	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '-' || r == '.' || r == '_'
	})

	var result string
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		if i == 0 {
			if exported {
				runes[0] = unicode.ToUpper(runes[0])
			} else {
				runes[0] = unicode.ToLower(runes[0])
			}
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		result += string(runes)
	}

	if result == "" {
		result = "X"
	}

	if goKeywords[result] {
		result += "_"
	}

	return result
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
