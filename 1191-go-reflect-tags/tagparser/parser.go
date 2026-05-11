package tagparser

import (
	"reflect"
	"strings"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseStruct(s interface{}) []FieldTags {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	return p.parseStructType(t, "")
}

func (p *Parser) parseStructType(t reflect.Type, prefix string) []FieldTags {
	var result []FieldTags

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		path := fieldName
		if prefix != "" {
			path = prefix + "." + fieldName
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		currentTags := parseFieldTags(field, path)

		if field.Anonymous && fieldType.Kind() == reflect.Struct {
			nestedTags := p.parseStructType(fieldType, path)
			mergedTags := p.mergeAnonymousTags(currentTags, nestedTags)
			result = append(result, mergedTags...)
		} else if fieldType.Kind() == reflect.Struct {
			result = append(result, currentTags)
			nestedTags := p.parseStructType(fieldType, path)
			result = append(result, nestedTags...)
		} else {
			result = append(result, currentTags)
		}
	}

	return result
}

func (p *Parser) mergeAnonymousTags(current FieldTags, nested []FieldTags) []FieldTags {
	var result []FieldTags

	for _, nestedTag := range nested {
		merged := FieldTags{
			Path: nestedTag.Path,
		}

		if current.JSON != nil {
			merged.JSON = current.JSON
		} else {
			merged.JSON = nestedTag.JSON
		}

		if current.DB != nil {
			merged.DB = current.DB
		} else {
			merged.DB = nestedTag.DB
		}

		if current.Validate != nil {
			merged.Validate = current.Validate
		} else {
			merged.Validate = nestedTag.Validate
		}

		result = append(result, merged)
	}

	return result
}

func (p *Parser) ParseFromDef(def StructDef) []FieldTags {
	return p.parseStructDef(def, "")
}

func (p *Parser) parseStructDef(def StructDef, prefix string) []FieldTags {
	var result []FieldTags

	for _, field := range def.Fields {
		path := field.Name
		if prefix != "" {
			path = prefix + "." + field.Name
		}

		currentTags := parseFieldTagsFromDef(field, path)

		if field.IsAnonymous && len(field.Fields) > 0 {
			nestedTags := p.parseStructDef(StructDef{Fields: field.Fields}, path)
			mergedTags := p.mergeAnonymousTags(currentTags, nestedTags)
			result = append(result, mergedTags...)
		} else if len(field.Fields) > 0 {
			result = append(result, currentTags)
			nestedTags := p.parseStructDef(StructDef{Fields: field.Fields}, path)
			result = append(result, nestedTags...)
		} else {
			result = append(result, currentTags)
		}
	}

	return result
}

func parseFieldTags(field reflect.StructField, path string) FieldTags {
	tags := FieldTags{Path: path}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		tags.JSON = ParseJSONTag(jsonTag)
	}

	dbTag := field.Tag.Get("db")
	if dbTag != "" {
		tags.DB = ParseDBTag(dbTag)
	}

	validateTag := field.Tag.Get("validate")
	if validateTag != "" {
		tags.Validate = ParseValidateTag(validateTag)
	}

	return tags
}

func parseFieldTagsFromDef(field StructFieldDef, path string) FieldTags {
	tags := FieldTags{Path: path}

	if field.Tags != nil {
		if jsonTag, ok := field.Tags["json"]; ok {
			tags.JSON = ParseJSONTag(jsonTag)
		}
		if dbTag, ok := field.Tags["db"]; ok {
			tags.DB = ParseDBTag(dbTag)
		}
		if validateTag, ok := field.Tags["validate"]; ok {
			tags.Validate = ParseValidateTag(validateTag)
		}
	}

	return tags
}

func (p *Parser) QueryField(allTags []FieldTags, fieldPath string) *FieldTags {
	for _, tag := range allTags {
		if strings.EqualFold(tag.Path, fieldPath) {
			return &tag
		}
	}
	return nil
}
