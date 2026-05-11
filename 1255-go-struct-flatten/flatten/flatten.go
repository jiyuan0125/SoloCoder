package flatten

import (
	"fmt"
	"strings"
	"unicode"

	"go-struct-flatten/common"
)

type Flattener struct {
	structs map[string]*StructInfo
}

func NewFlattener(source string) (*Flattener, error) {
	structs, err := ParseSource(source)
	if err != nil {
		return nil, err
	}
	return &Flattener{structs: structs}, nil
}

func (f *Flattener) ListStructs() []string {
	names := make([]string, 0, len(f.structs))
	for name := range f.structs {
		names = append(names, name)
	}
	return names
}

func (f *Flattener) Flatten(structName string) ([]common.FieldInfo, error) {
	root, exists := f.structs[structName]
	if !exists {
		return nil, fmt.Errorf("struct %s not found", structName)
	}

	ctx := &flattenContext{
		flattener:  f,
		fieldInfos: []common.FieldInfo{},
		seenTypes:  map[string]bool{},
		duplicateFields: map[string][]string{},
	}

	ctx.flattenStruct(root, "", "", false)

	fields := ctx.fieldInfos
	for i := range fields {
		if paths, exists := ctx.duplicateFields[fields[i].FlatKey]; exists && len(paths) > 1 {
			fields[i].Notes = append(fields[i].Notes, fmt.Sprintf("ambiguous: multiple paths -> %s", strings.Join(paths, ", ")))
		}
	}

	return fields, nil
}

type flattenContext struct {
	flattener       *Flattener
	fieldInfos      []common.FieldInfo
	seenTypes       map[string]bool
	duplicateFields map[string][]string
}

func (ctx *flattenContext) flattenStruct(s *StructInfo, flatPrefix, pathPrefix string, parentPointer bool) {
	if ctx.seenTypes[s.Name] {
		return
	}
	ctx.seenTypes[s.Name] = true
	defer func() { delete(ctx.seenTypes, s.Name) }()

	for _, field := range s.Fields {
		if !field.Exported && !field.Embedded {
			continue
		}
		if field.JSONTag != nil && field.JSONTag.Ignored {
			continue
		}

		fieldName := field.Name
		jsonName := ""
		hasOmitempty := false
		if field.JSONTag != nil {
			if field.JSONTag.Name != "" {
				jsonName = field.JSONTag.Name
				fieldName = jsonName
			}
			hasOmitempty = field.JSONTag.Omitempty
		}

		keyPart := toSnakeCase(fieldName)

		flatKey := keyPart
		if flatPrefix != "" {
			flatKey = flatPrefix + "_" + keyPart
		}

		originalPath := fieldName
		if pathPrefix != "" {
			originalPath = pathPrefix + "." + fieldName
		}

		ctx.flattenType(field.Type, flatKey, originalPath, hasOmitempty, field.Embedded, parentPointer || field.Type.IsPointer)
	}
}

func (ctx *flattenContext) flattenType(t TypeInfo, flatKey, originalPath string, hasOmitempty, embedded, hasPointer bool) {
	switch t.Kind {
	case KindBasic, KindInterface:
		typeStr := t.Name
		if t.IsPointer {
			typeStr = "*" + typeStr
		}

		info := common.FieldInfo{
			FlatKey:         flatKey,
			OriginalPath:    originalPath,
			Type:            typeStr,
			HasOmitempty:    hasOmitempty,
			HasMultiplePaths: hasPointer,
		}
		if embedded {
			info.Notes = append(info.Notes, "embedded")
		}
		ctx.fieldInfos = append(ctx.fieldInfos, info)
		ctx.duplicateFields[flatKey] = append(ctx.duplicateFields[flatKey], originalPath)

	case KindStruct:
		typeStr := t.Name
		if t.IsPointer {
			typeStr = "*" + typeStr
		}

		if ctx.seenTypes[t.Name] {
			info := common.FieldInfo{
				FlatKey:         flatKey,
				OriginalPath:    originalPath,
				Type:            typeStr,
				IsCircular:      true,
				HasOmitempty:    hasOmitempty,
				HasMultiplePaths: hasPointer,
			}
			ctx.fieldInfos = append(ctx.fieldInfos, info)
			return
		}

		info := common.FieldInfo{
			FlatKey:         flatKey,
			OriginalPath:    originalPath,
			Type:            typeStr,
			HasOmitempty:    hasOmitempty,
			HasMultiplePaths: hasPointer,
		}
		ctx.fieldInfos = append(ctx.fieldInfos, info)
		ctx.duplicateFields[flatKey] = append(ctx.duplicateFields[flatKey], originalPath)

		if subStruct, exists := ctx.flattener.structs[t.Name]; exists {
			ctx.flattenStruct(subStruct, flatKey, originalPath, hasPointer || t.IsPointer)
		}

	case KindSlice, KindArray:
		typeStr := "[]"
		if t.IsPointer {
			typeStr = "*[]"
		}
		if t.Elem != nil {
			typeStr += typeString(*t.Elem)
		}

		ctx.fieldInfos = append(ctx.fieldInfos, common.FieldInfo{
			FlatKey:         flatKey,
			OriginalPath:    originalPath,
			Type:            typeStr,
			IsArrayIndex:    true,
			HasOmitempty:    hasOmitempty,
			HasMultiplePaths: hasPointer,
			Notes:           []string{"wildcard index: use _N or _* for array elements"},
		})
		ctx.duplicateFields[flatKey] = append(ctx.duplicateFields[flatKey], originalPath)

		if t.Elem != nil {
			indexFlatKey := flatKey + "_0"
			indexPath := originalPath + "[0]"
			ctx.flattenType(*t.Elem, indexFlatKey, indexPath, false, false, hasPointer || t.IsPointer)

			wildcardFlatKey := flatKey + "_*"
			wildcardPath := originalPath + "[*]"
			ctx.fieldInfos = append(ctx.fieldInfos, common.FieldInfo{
				FlatKey:         wildcardFlatKey,
				OriginalPath:    wildcardPath,
				Type:            typeString(*t.Elem),
				IsArrayIndex:    true,
				Notes:           []string{"wildcard pattern"},
			})
		}

	case KindMap:
		typeStr := "map["
		if t.Key != nil {
			typeStr += typeString(*t.Key)
		}
		typeStr += "]"
		if t.Elem != nil {
			typeStr += typeString(*t.Elem)
		}
		if t.IsPointer {
			typeStr = "*" + typeStr
		}

		ctx.fieldInfos = append(ctx.fieldInfos, common.FieldInfo{
			FlatKey:         flatKey,
			OriginalPath:    originalPath,
			Type:            typeStr,
			IsMapKey:        true,
			HasOmitempty:    hasOmitempty,
			HasMultiplePaths: hasPointer,
			Notes:           []string{"use map key as path segment"},
		})
		ctx.duplicateFields[flatKey] = append(ctx.duplicateFields[flatKey], originalPath)

		if t.Elem != nil {
			keyPlaceholder := "key"
			if t.Key != nil && t.Key.Kind == KindBasic && t.Key.Name == "string" {
				keyPlaceholder = "<string_key>"
			} else if t.Key != nil {
				keyPlaceholder = fmt.Sprintf("<%s_key>", t.Key.Name)
			}

			cleanKey := strings.ReplaceAll(strings.ReplaceAll(keyPlaceholder, "<", ""), ">", "")
			mapFlatKey := flatKey + "_" + cleanKey
			mapPath := originalPath + "[" + keyPlaceholder + "]"
			ctx.flattenType(*t.Elem, mapFlatKey, mapPath, false, false, hasPointer || t.IsPointer)
		}
	}
}

func typeString(t TypeInfo) string {
	ptr := ""
	if t.IsPointer {
		ptr = "*"
	}

	switch t.Kind {
	case KindBasic, KindStruct, KindInterface:
		return ptr + t.Name
	case KindSlice, KindArray:
		elem := ""
		if t.Elem != nil {
			elem = typeString(*t.Elem)
		}
		return ptr + "[]" + elem
	case KindMap:
		key := ""
		if t.Key != nil {
			key = typeString(*t.Key)
		}
		elem := ""
		if t.Elem != nil {
			elem = typeString(*t.Elem)
		}
		return ptr + "map[" + key + "]" + elem
	}
	return ptr + t.Name
}

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result []rune
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if !(unicode.IsUpper(prev) || prev == '_') {
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	return strings.ToLower(string(result))
}
