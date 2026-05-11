package executor

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"graphql-parser/internal/graphql/ast"
	"graphql-parser/internal/graphql/parser"
	"graphql-parser/internal/graphql/registry"
)

const (
	DefaultMaxDepth = 100
)

type ExecutionContext struct {
	Registry   *registry.Registry
	Variables  map[string]interface{}
	MaxDepth   int
	Warnings   []string
	Depth      int
	fragments  map[string]*ast.FragmentDefinition
}

func NewExecutionContext(reg *registry.Registry) *ExecutionContext {
	return &ExecutionContext{
		Registry:  reg,
		Variables: make(map[string]interface{}),
		MaxDepth:  DefaultMaxDepth,
		Warnings:  []string{},
		fragments: make(map[string]*ast.FragmentDefinition),
	}
}

func (ec *ExecutionContext) addWarning(msg string) {
	ec.Warnings = append(ec.Warnings, msg)
}

func (ec *ExecutionContext) enterField() error {
	ec.Depth++
	if ec.Depth > ec.MaxDepth {
		return fmt.Errorf("query depth exceeds maximum limit of %d", ec.MaxDepth)
	}
	return nil
}

func (ec *ExecutionContext) exitField() {
	ec.Depth--
}

type ExecutionResult struct {
	Data     interface{} `json:"data,omitempty"`
	Errors   []string    `json:"errors,omitempty"`
	Warnings []string    `json:"warnings,omitempty"`
}

func Parse(query string) (*ast.Document, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("empty query")
	}
	return parser.Parse(query)
}

func Execute(ctx context.Context, ec *ExecutionContext, query string, variables map[string]interface{}) (*ExecutionResult, error) {
	doc, err := Parse(query)
	if err != nil {
		return &ExecutionResult{Errors: []string{err.Error()}}, err
	}
	return ExecuteDocument(ctx, ec, doc, variables)
}

func ExecuteDocument(ctx context.Context, ec *ExecutionContext, doc *ast.Document, variables map[string]interface{}) (*ExecutionResult, error) {
	ec.Registry.ResetCallCounts()
	ec.Warnings = []string{}
	ec.Depth = 0
	ec.fragments = make(map[string]*ast.FragmentDefinition)

	for _, def := range doc.Definitions {
		if frag, ok := def.(*ast.FragmentDefinition); ok {
			ec.fragments[frag.Name] = frag
		}
	}

	if variables != nil {
		for k, v := range variables {
			ec.Variables[k] = v
		}
	}

	result := make(map[string]interface{})
	errors := []string{}

	for _, def := range doc.Definitions {
		if op, ok := def.(*ast.OperationDefinition); ok {
			opErrs := ec.validateVariables(op)
			if len(opErrs) > 0 {
				errors = append(errors, opErrs...)
				continue
			}

			var rootTypeName string
			switch op.OperationType {
			case ast.OperationTypeQuery:
				rootTypeName = ec.Registry.GetQueryType()
			case ast.OperationTypeMutation:
				rootTypeName = ec.Registry.GetMutationType()
				if rootTypeName == "" {
					rootTypeName = ec.Registry.GetQueryType()
				}
			default:
				rootTypeName = ec.Registry.GetQueryType()
			}

			opResult, err := ec.executeSelectionSet(ctx, nil, rootTypeName, op.SelectionSet)
			if err != nil {
				errors = append(errors, err.Error())
				continue
			}
			if opResultMap, ok := opResult.(map[string]interface{}); ok {
				for k, v := range opResultMap {
					result[k] = v
				}
			}
		}
	}

	warnings := ec.Registry.NPlus1Warnings()
	ec.Warnings = append(ec.Warnings, warnings...)

	execResult := &ExecutionResult{
		Data:     result,
		Errors:   errors,
		Warnings: ec.Warnings,
	}

	if len(errors) > 0 {
		return execResult, fmt.Errorf("execution errors: %v", errors)
	}

	return execResult, nil
}

func (ec *ExecutionContext) validateVariables(op *ast.OperationDefinition) []string {
	errs := []string{}
	for _, varDef := range op.Variables {
		val, provided := ec.Variables[varDef.Variable]
		if !provided {
			if varDef.DefaultValue != nil {
				ec.Variables[varDef.Variable] = varDef.DefaultValue
				continue
			}
			if varDef.Type.IsNonNull() {
				errs = append(errs, fmt.Sprintf("variable %s is required but not provided", varDef.Variable))
			}
			continue
		}

		typeErr := ec.validateVariableType(varDef.Type, val, varDef.Variable)
		if typeErr != nil {
			errs = append(errs, typeErr.Error())
		}
	}
	return errs
}

func (ec *ExecutionContext) validateVariableType(t ast.Type, val interface{}, varName string) error {
	switch tt := t.(type) {
	case *ast.NonNullType:
		if val == nil {
			return fmt.Errorf("variable %s: expected non-null value for type %s", varName, t)
		}
		return ec.validateVariableType(tt.OfType, val, varName)
	case *ast.ListType:
		if val == nil {
			return nil
		}
		rv := reflect.ValueOf(val)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return fmt.Errorf("variable %s: expected list for type %s, got %T", varName, t, val)
		}
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			if err := ec.validateVariableType(tt.OfType, elem, varName); err != nil {
				return err
			}
		}
		return nil
	case *ast.NamedType:
		if val == nil {
			return nil
		}
		switch tt.Name {
		case "Int":
			if _, ok := val.(int64); ok {
				return nil
			}
			if _, ok := val.(int); ok {
				return nil
			}
			if _, ok := val.(float64); ok {
				return fmt.Errorf("variable %s: expected Int, got Float", varName)
			}
			return fmt.Errorf("variable %s: expected Int, got %T", varName, val)
		case "Float":
			if _, ok := val.(float64); ok {
				return nil
			}
			if _, ok := val.(int64); ok {
				return nil
			}
			if _, ok := val.(int); ok {
				return nil
			}
			return fmt.Errorf("variable %s: expected Float, got %T", varName, val)
		case "String":
			if _, ok := val.(string); ok {
				return nil
			}
			return fmt.Errorf("variable %s: expected String, got %T", varName, val)
		case "Boolean":
			if _, ok := val.(bool); ok {
				return nil
			}
			return fmt.Errorf("variable %s: expected Boolean, got %T", varName, val)
		case "ID":
			if _, ok := val.(string); ok {
				return nil
			}
			if _, ok := val.(int64); ok {
				return nil
			}
			if _, ok := val.(int); ok {
				return nil
			}
			return fmt.Errorf("variable %s: expected ID, got %T", varName, val)
		default:
			return nil
		}
	}
	return nil
}

func (ec *ExecutionContext) executeSelectionSet(ctx context.Context, source interface{}, typeName string, selSet ast.SelectionSet) (interface{}, error) {
	if err := ec.enterField(); err != nil {
		return nil, err
	}
	defer ec.exitField()

	if selSet == nil {
		return source, nil
	}

	result := make(map[string]interface{})

	expanded := ec.expandSelections(typeName, selSet)

	for _, sel := range expanded {
		switch s := sel.(type) {
		case *ast.Field:
			fieldVal, err := ec.executeField(ctx, source, typeName, s)
			if err != nil {
				return nil, err
			}
			resultKey := s.Name
			if s.Alias != "" {
				resultKey = s.Alias
			}
			result[resultKey] = fieldVal
		}
	}

	return result, nil
}

func (ec *ExecutionContext) expandSelections(typeName string, selSet ast.SelectionSet) []ast.Selection {
	result := []ast.Selection{}
	for _, sel := range selSet {
		switch s := sel.(type) {
		case *ast.Field:
			result = append(result, s)
		case *ast.FragmentSpread:
			if frag, ok := ec.fragments[s.Name]; ok {
				if frag.TypeCondition == "" || frag.TypeCondition == typeName {
					result = append(result, ec.expandSelections(typeName, frag.SelectionSet)...)
				}
			}
		case *ast.InlineFragment:
			if s.TypeCondition == "" || s.TypeCondition == typeName {
				result = append(result, ec.expandSelections(typeName, s.SelectionSet)...)
			}
		}
	}
	return result
}

func (ec *ExecutionContext) executeField(ctx context.Context, source interface{}, parentType string, field *ast.Field) (interface{}, error) {
	resolver, hasResolver := ec.Registry.GetResolver(parentType, field.Name)

	if !hasResolver || resolver == nil {
		if source != nil {
			rv := reflect.ValueOf(source)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Map {
				if mv := rv.MapIndex(reflect.ValueOf(field.Name)); mv.IsValid() {
					val := mv.Interface()
					if field.SelectionSet != nil && len(field.SelectionSet) > 0 {
						fieldType, _ := ec.Registry.GetFieldType(parentType, field.Name)
						return ec.executeSelectionSet(ctx, val, cleanTypeForExecution(fieldType), field.SelectionSet)
					}
					return val, nil
				}
			}
			if rv.Kind() == reflect.Struct {
				fv := rv.FieldByName(field.Name)
				if fv.IsValid() {
					val := fv.Interface()
					if field.SelectionSet != nil && len(field.SelectionSet) > 0 {
						fieldType, _ := ec.Registry.GetFieldType(parentType, field.Name)
						return ec.executeSelectionSet(ctx, val, cleanTypeForExecution(fieldType), field.SelectionSet)
					}
					return val, nil
				}
			}
		}
		if !hasResolver {
			return nil, fmt.Errorf("field %s not found on type %s", field.Name, parentType)
		}
		return nil, nil
	}

	args, err := ec.buildArgs(field.Arguments)
	if err != nil {
		return nil, err
	}

	ec.Registry.RecordResolverCall(parentType, field.Name)

	val, err := resolver(ctx, source, args)
	if err != nil {
		return nil, err
	}

	if field.SelectionSet != nil && len(field.SelectionSet) > 0 {
		fieldType, _ := ec.Registry.GetFieldType(parentType, field.Name)
		return ec.executeValueWithSelections(ctx, val, cleanTypeForExecution(fieldType), field.SelectionSet)
	}

	return val, nil
}

func (ec *ExecutionContext) executeValueWithSelections(ctx context.Context, val interface{}, typeName string, selSet ast.SelectionSet) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		val = rv.Interface()
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		results := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			elemResult, err := ec.executeSelectionSet(ctx, elem, typeName, selSet)
			if err != nil {
				return nil, err
			}
			results[i] = elemResult
		}
		return results, nil
	}

	return ec.executeSelectionSet(ctx, val, typeName, selSet)
}

func (ec *ExecutionContext) buildArgs(args []ast.Argument) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, arg := range args {
		val, err := ec.resolveValue(arg.Value)
		if err != nil {
			return nil, err
		}
		result[arg.Name] = val
	}
	return result, nil
}

func (ec *ExecutionContext) resolveValue(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case *ast.Variable:
		if resolved, ok := ec.Variables[val.Name]; ok {
			return resolved, nil
		}
		return nil, fmt.Errorf("variable $%s not provided", val.Name)
	case []ast.Value:
		list := make([]interface{}, len(val))
		for i, item := range val {
			resolved, err := ec.resolveValue(item)
			if err != nil {
				return nil, err
			}
			list[i] = resolved
		}
		return list, nil
	case map[string]ast.Value:
		obj := make(map[string]interface{})
		for k, item := range val {
			resolved, err := ec.resolveValue(item)
			if err != nil {
				return nil, err
			}
			obj[k] = resolved
		}
		return obj, nil
	default:
		return v, nil
	}
}

func cleanTypeForExecution(t string) string {
	t = strings.TrimSuffix(t, "!")
	t = strings.TrimPrefix(t, "[")
	t = strings.TrimSuffix(t, "!")
	t = strings.TrimSuffix(t, "]")
	t = strings.TrimSuffix(t, "!")
	return t
}

func Format(query string) (string, error) {
	doc, err := Parse(query)
	if err != nil {
		return "", err
	}
	return formatDocument(doc), nil
}

func formatDocument(doc *ast.Document) string {
	var sb strings.Builder
	for i, def := range doc.Definitions {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		switch d := def.(type) {
		case *ast.OperationDefinition:
			sb.WriteString(formatOperation(d))
		case *ast.FragmentDefinition:
			sb.WriteString(formatFragment(d))
		}
	}
	return sb.String()
}

func formatOperation(op *ast.OperationDefinition) string {
	var sb strings.Builder
	sb.WriteString(string(op.OperationType))
	if op.Name != "" {
		sb.WriteString(" " + op.Name)
	}
	if len(op.Variables) > 0 {
		sb.WriteString("(")
		for i, v := range op.Variables {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("$" + v.Variable + ": " + v.Type.String())
			if v.DefaultValue != nil {
				sb.WriteString(" = " + formatValue(v.DefaultValue))
			}
		}
		sb.WriteString(")")
	}
	sb.WriteString(" ")
	sb.WriteString(formatSelectionSet(op.SelectionSet, 0))
	return sb.String()
}

func formatFragment(frag *ast.FragmentDefinition) string {
	var sb strings.Builder
	sb.WriteString("fragment " + frag.Name + " on " + frag.TypeCondition + " ")
	sb.WriteString(formatSelectionSet(frag.SelectionSet, 0))
	return sb.String()
}

func formatSelectionSet(selSet ast.SelectionSet, indent int) string {
	if len(selSet) == 0 {
		return "{}"
	}
	var sb strings.Builder
	indentStr := strings.Repeat("  ", indent)
	sb.WriteString("{\n")
	for _, sel := range selSet {
		switch s := sel.(type) {
		case *ast.Field:
			sb.WriteString(indentStr + "  " + formatField(s, indent+1))
		case *ast.FragmentSpread:
			sb.WriteString(indentStr + "  ..." + s.Name)
		case *ast.InlineFragment:
			sb.WriteString(indentStr + "  ...")
			if s.TypeCondition != "" {
				sb.WriteString(" on " + s.TypeCondition)
			}
			sb.WriteString(" " + formatSelectionSet(s.SelectionSet, indent+1))
		}
		sb.WriteString("\n")
	}
	sb.WriteString(indentStr + "}")
	return sb.String()
}

func formatField(field *ast.Field, indent int) string {
	var sb strings.Builder
	if field.Alias != "" {
		sb.WriteString(field.Alias + ": ")
	}
	sb.WriteString(field.Name)
	if len(field.Arguments) > 0 {
		sb.WriteString("(")
		for i, arg := range field.Arguments {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(arg.Name + ": " + formatValue(arg.Value))
		}
		sb.WriteString(")")
	}
	if field.SelectionSet != nil && len(field.SelectionSet) > 0 {
		sb.WriteString(" ")
		sb.WriteString(formatSelectionSet(field.SelectionSet, indent))
	}
	return sb.String()
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return "\"" + val + "\""
	case int64, int, float64, bool:
		return fmt.Sprintf("%v", val)
	case nil:
		return "null"
	case *ast.Variable:
		return "$" + val.Name
	case []ast.Value:
		var sb strings.Builder
		sb.WriteString("[")
		for i, item := range val {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(formatValue(item))
		}
		sb.WriteString("]")
		return sb.String()
	case map[string]ast.Value:
		var sb strings.Builder
		sb.WriteString("{")
		first := true
		for k, item := range val {
			if !first {
				sb.WriteString(", ")
			}
			first = false
			sb.WriteString(k + ": " + formatValue(item))
		}
		sb.WriteString("}")
		return sb.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func Validate(query string) ([]string, error) {
	_, err := Parse(query)
	if err != nil {
		return nil, err
	}
	return []string{}, nil
}
