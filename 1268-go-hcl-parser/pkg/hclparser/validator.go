package hclparser

import (
	"fmt"
)

type Schema struct {
	BlockTypes map[string]*BlockSchema
}

type BlockSchema struct {
	Labels     int
	MinLabels  int
	MaxLabels  int
	Attributes map[string]*AttributeSchema
	NestedBlocks map[string]*BlockSchema
}

type AttributeSchema struct {
	Required bool
	Type     string
}

func NewSchema() *Schema {
	return &Schema{
		BlockTypes: make(map[string]*BlockSchema),
	}
}

type ValidationError struct {
	Path    string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

func Validate(body *Body, schema *Schema) ([]*ValidationError, error) {
	errors := []*ValidationError{}

	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *Block:
			blockSchema, ok := schema.BlockTypes[s.Type]
			if !ok {
				errors = append(errors, &ValidationError{
					Path:    s.Type,
					Message: fmt.Sprintf("unknown block type: %s", s.Type),
				})
				continue
			}
			errors = append(errors, validateBlock(s, blockSchema, s.Type)...)
		}
	}

	return errors, nil
}

func validateBlock(block *Block, schema *BlockSchema, path string) []*ValidationError {
	errors := []*ValidationError{}

	labelCount := len(block.Labels)
	if schema.MinLabels > 0 && labelCount < schema.MinLabels {
		errors = append(errors, &ValidationError{
			Path:    path,
			Message: fmt.Sprintf("block %s requires at least %d labels, got %d", block.Type, schema.MinLabels, labelCount),
		})
	}
	if schema.MaxLabels > 0 && labelCount > schema.MaxLabels {
		errors = append(errors, &ValidationError{
			Path:    path,
			Message: fmt.Sprintf("block %s requires at most %d labels, got %d", block.Type, schema.MaxLabels, labelCount),
		})
	}

	foundAttrs := make(map[string]bool)
	for _, stmt := range block.Body.Statements {
		switch s := stmt.(type) {
		case *Attribute:
			foundAttrs[s.Key] = true
			attrSchema, ok := schema.Attributes[s.Key]
			if !ok {
				errors = append(errors, &ValidationError{
					Path:    fmt.Sprintf("%s.%s", path, s.Key),
					Message: fmt.Sprintf("unknown attribute: %s", s.Key),
				})
			} else if attrSchema.Type != "" {
				if !isType(s.Value, attrSchema.Type) {
					errors = append(errors, &ValidationError{
						Path:    fmt.Sprintf("%s.%s", path, s.Key),
						Message: fmt.Sprintf("attribute %s expected type %s", s.Key, attrSchema.Type),
					})
				}
			}
		case *Block:
			nestedSchema, ok := schema.NestedBlocks[s.Type]
			if !ok {
				errors = append(errors, &ValidationError{
					Path:    fmt.Sprintf("%s.%s", path, s.Type),
					Message: fmt.Sprintf("unknown nested block: %s", s.Type),
				})
				continue
			}
			nestedPath := path
			for _, label := range s.Labels {
				nestedPath = fmt.Sprintf("%s.%s", nestedPath, label)
			}
			errors = append(errors, validateBlock(s, nestedSchema, nestedPath)...)
		}
	}

	for key, attrSchema := range schema.Attributes {
		if attrSchema.Required && !foundAttrs[key] {
			errors = append(errors, &ValidationError{
				Path:    fmt.Sprintf("%s.%s", path, key),
				Message: fmt.Sprintf("required attribute missing: %s", key),
			})
		}
	}

	return errors
}

func isType(expr Expression, typ string) bool {
	switch e := expr.(type) {
	case *LiteralValue:
		switch typ {
		case "string":
			_, ok := e.Value.(string)
			return ok
		case "number":
			switch e.Value.(type) {
			case int64, float64:
				return true
			}
			return false
		case "bool":
			_, ok := e.Value.(bool)
			return ok
		case "null":
			return e.Value == nil
		}
	case *ListExpr:
		return typ == "list"
	case *MapExpr:
		return typ == "map"
	case *VariableReference:
		return true
	case *FunctionCall:
		return true
	case *ConditionalExpr:
		return true
	case *HeredocValue:
		return typ == "string"
	}
	return false
}
