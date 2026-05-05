package template

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type RenderResult struct {
	Content  string
	Warnings []string
}

type Engine struct {
	templatePath string
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) LoadFromFile(path string) (*Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("加载模板文件失败: %w", err)
	}
	
	tmpl, err := e.Compile(string(content))
	if err != nil {
		return nil, err
	}
	
	tmpl.path = path
	return tmpl, nil
}

func (e *Engine) Compile(source string) (*Template, error) {
	lexer := NewLexer(source)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, err
	}
	
	parser := NewParser(tokens)
	root, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	
	return &Template{
		source: source,
		root:   root,
		path:   "",
	}, nil
}

type Template struct {
	source string
	root   *RootNode
	path   string
}

func (t *Template) Render(data interface{}) (*RenderResult, error) {
	context := &RenderContext{
		data:     data,
		warnings: []string{},
		loopItem: nil,
	}
	
	content, err := t.renderNode(t.root, context)
	if err != nil {
		return nil, err
	}
	
	return &RenderResult{
		Content:  content,
		Warnings: context.warnings,
	}, nil
}

func (t *Template) renderNode(node Node, context *RenderContext) (string, error) {
	switch n := node.(type) {
	case *RootNode:
		return t.renderNodes(n.Children, context)
		
	case *TextNode:
		return n.Content, nil
		
	case *VariableNode:
		return t.renderVariable(n, context)
		
	case *IfNode:
		return t.renderIf(n, context)
		
	case *EachNode:
		return t.renderEach(n, context)
		
	default:
		return "", nil
	}
}

func (t *Template) renderNodes(nodes []Node, context *RenderContext) (string, error) {
	var result strings.Builder
	
	for _, node := range nodes {
		content, err := t.renderNode(node, context)
		if err != nil {
			return "", err
		}
		result.WriteString(content)
	}
	
	return result.String(), nil
}

func (t *Template) renderVariable(node *VariableNode, context *RenderContext) (string, error) {
	varName := node.Name
	
	if strings.HasPrefix(varName, ".") && context.loopItem != nil {
		varName = varName[1:]
		if varName == "" {
			return fmt.Sprintf("%v", context.loopItem), nil
		}
		
		value, found := getProperty(context.loopItem, varName)
		if found {
			return fmt.Sprintf("%v", value), nil
		}
		return t.handleMissingVariable(node, context), nil
	}
	
	value, found := getProperty(context.data, varName)
	if found {
		return fmt.Sprintf("%v", value), nil
	}
	
	return t.handleMissingVariable(node, context), nil
}

func (t *Template) handleMissingVariable(node *VariableNode, context *RenderContext) string {
	if node.DefaultValue != "" {
		return node.DefaultValue
	}
	
	context.addWarning(fmt.Sprintf("变量 '%s' 未找到，保留原始占位符", node.Name))
	return node.Raw
}

func (t *Template) renderIf(node *IfNode, context *RenderContext) (string, error) {
	condition := node.Condition
	
	var value interface{}
	var found bool
	
	if strings.HasPrefix(condition, ".") && context.loopItem != nil {
		propName := condition[1:]
		if propName == "" {
			value = context.loopItem
			found = true
		} else {
			value, found = getProperty(context.loopItem, propName)
		}
	} else {
		value, found = getProperty(context.data, condition)
	}
	
	conditionTrue := found && isTruthy(value)
	
	if conditionTrue {
		return t.renderNodes(node.Then, context)
	} else if len(node.Else) > 0 {
		return t.renderNodes(node.Else, context)
	}
	
	return "", nil
}

func (t *Template) renderEach(node *EachNode, context *RenderContext) (string, error) {
	arrayName := node.ArrayName
	
	var value interface{}
	var found bool
	
	if strings.HasPrefix(arrayName, ".") && context.loopItem != nil {
		propName := arrayName[1:]
		if propName == "" {
			value = context.loopItem
			found = true
		} else {
			value, found = getProperty(context.loopItem, propName)
		}
	} else {
		value, found = getProperty(context.data, arrayName)
	}
	
	if !found {
		context.addWarning(fmt.Sprintf("each循环: 数组 '%s' 未找到", arrayName))
		return "", nil
	}
	
	reflectValue := reflect.ValueOf(value)
	if !isIterable(reflectValue) {
		context.addWarning(fmt.Sprintf("each循环: '%s' 不是可迭代类型", arrayName))
		return "", nil
	}
	
	var result strings.Builder
	
	for i := 0; i < reflectValue.Len(); i++ {
		item := reflectValue.Index(i).Interface()
		
		itemContext := &RenderContext{
			data:     context.data,
			warnings: context.warnings,
			loopItem: item,
		}
		
		content, err := t.renderNodes(node.Body, itemContext)
		if err != nil {
			return "", err
		}
		result.WriteString(content)
	}
	
	return result.String(), nil
}

type RenderContext struct {
	data     interface{}
	warnings []string
	loopItem interface{}
}

func (c *RenderContext) addWarning(msg string) {
	c.warnings = append(c.warnings, msg)
}

func getProperty(data interface{}, path string) (interface{}, bool) {
	if data == nil {
		return nil, false
	}
	
	parts := strings.Split(path, ".")
	current := data
	
	for i, part := range parts {
		if current == nil {
			return nil, false
		}
		
		v := reflect.ValueOf(current)
		
		switch v.Kind() {
		case reflect.Map:
			keyVal := reflect.ValueOf(part)
			if v.Type().Key().Kind() == reflect.String {
				mapValue := v.MapIndex(keyVal)
				if mapValue.IsValid() {
					current = mapValue.Interface()
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
			
		case reflect.Struct:
			fieldVal := v.FieldByName(part)
			if fieldVal.IsValid() {
				if fieldVal.CanInterface() {
					current = fieldVal.Interface()
				} else {
					return nil, false
				}
			} else {
				methodVal := v.MethodByName(part)
				if methodVal.IsValid() && methodVal.Type().NumIn() == 0 && methodVal.Type().NumOut() == 1 {
					results := methodVal.Call(nil)
					current = results[0].Interface()
				} else {
					return nil, false
				}
			}
			
		case reflect.Ptr, reflect.Interface:
			if v.IsNil() {
				return nil, false
			}
			current = v.Elem().Interface()
			remainingParts := parts[i:]
			return getProperty(current, strings.Join(remainingParts, "."))
			
		default:
			return nil, false
		}
	}
	
	return current, true
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0.0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !v.IsNil()
	default:
		return true
	}
}

func isIterable(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

func (t *Template) Path() string {
	return t.path
}

func RenderString(source string, data interface{}) (*RenderResult, error) {
	engine := NewEngine()
	tmpl, err := engine.Compile(source)
	if err != nil {
		return nil, err
	}
	return tmpl.Render(data)
}

func RenderFile(path string, data interface{}) (*RenderResult, error) {
	engine := NewEngine()
	tmpl, err := engine.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	return tmpl.Render(data)
}
