package tagparser

import (
	"strings"

	"go-reflect-tags/pkg/common"
)

func (p *Parser) QueryField(structName, fieldPath string) (*common.FieldTagInfo, error) {
	s, ok := p.structs[structName]
	if !ok {
		return nil, &notFoundError{structName: structName}
	}

	if fieldPath == "" {
		info := &common.FieldTagInfo{
			FieldPath: structName,
			FieldType: s.Type,
		}
		if jsonTag, ok := s.Tags["json"]; ok {
			info.Json = ParseJsonTag(jsonTag)
		}
		if dbTag, ok := s.Tags["db"]; ok {
			info.Db = ParseDbTag(dbTag)
		}
		if validateTag, ok := s.Tags["validate"]; ok {
			info.Validate = ParseValidateTag(validateTag, s.Type)
		}
		return info, nil
	}

	parts := strings.Split(fieldPath, ".")
	current := s
	currentPath := structName
	mergedTags := make(map[string]string)

	for i, part := range parts {
		found := false
		for _, field := range current.Fields {
			if field.Name == part {
				currentPath = currentPath + "." + part

				if field.IsAnonymous && len(field.Fields) > 0 {
					mergedTags = mergeTags(field.Tags, mergedTags)
					current = &field
					found = true
					continue
				}

				if field.IsNested && len(field.Fields) > 0 {
					current = &field
					mergedTags = mergeTags(field.Tags, mergedTags)
					found = true

					if i == len(parts)-1 {
						info := &common.FieldTagInfo{
							FieldPath: currentPath,
							FieldType: field.Type,
						}
						finalTags := mergeTags(field.Tags, mergedTags)
						if jsonTag, ok := finalTags["json"]; ok {
							info.Json = ParseJsonTag(jsonTag)
						}
						if dbTag, ok := finalTags["db"]; ok {
							info.Db = ParseDbTag(dbTag)
						}
						if validateTag, ok := finalTags["validate"]; ok {
							info.Validate = ParseValidateTag(validateTag, field.Type)
						}
						return info, nil
					}

					continue
				}

				finalTags := mergeTags(field.Tags, mergedTags)
				info := &common.FieldTagInfo{
					FieldPath: currentPath,
					FieldType: field.Type,
				}

				if jsonTag, ok := finalTags["json"]; ok {
					info.Json = ParseJsonTag(jsonTag)
				}
				if dbTag, ok := finalTags["db"]; ok {
					info.Db = ParseDbTag(dbTag)
				}
				if validateTag, ok := finalTags["validate"]; ok {
					info.Validate = ParseValidateTag(validateTag, field.Type)
				}

				return info, nil
			}
		}

		if !found {
			return nil, &notFoundError{
				structName: structName,
				fieldPath:  fieldPath,
			}
		}
	}

	return nil, &notFoundError{
		structName: structName,
		fieldPath:  fieldPath,
	}
}

func (p *Parser) ListFields(structName string) ([]*common.FieldTagInfo, error) {
	s, ok := p.structs[structName]
	if !ok {
		return nil, &notFoundError{structName: structName}
	}

	var result []*common.FieldTagInfo
	p.collectFields(s, structName, make(map[string]string), &result)
	return result, nil
}

func (p *Parser) collectFields(field *common.FieldDesc, path string, parentTags map[string]string, result *[]*common.FieldTagInfo) {
	mergedTags := mergeTags(field.Tags, parentTags)

	if len(field.Fields) == 0 {
		info := &common.FieldTagInfo{
			FieldPath: path,
			FieldType: field.Type,
		}

		if jsonTag, ok := mergedTags["json"]; ok {
			info.Json = ParseJsonTag(jsonTag)
		}
		if dbTag, ok := mergedTags["db"]; ok {
			info.Db = ParseDbTag(dbTag)
		}
		if validateTag, ok := mergedTags["validate"]; ok {
			info.Validate = ParseValidateTag(validateTag, field.Type)
		}

		*result = append(*result, info)
		return
	}

	for _, f := range field.Fields {
		newPath := path + "." + f.Name
		p.collectFields(&f, newPath, mergedTags, result)
	}
}

type notFoundError struct {
	structName string
	fieldPath  string
}

func (e *notFoundError) Error() string {
	if e.fieldPath != "" {
		return "field not found: " + e.structName + "." + e.fieldPath
	}
	return "struct not found: " + e.structName
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*notFoundError)
	return ok
}
