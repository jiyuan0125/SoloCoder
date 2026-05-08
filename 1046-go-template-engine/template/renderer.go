package template

import (
	"fmt"
	"reflect"
	"strings"
)

type renderer struct {
	engine *Engine
}

func newRenderer(engine *Engine) *renderer {
	return &renderer{engine: engine}
}

func (r *renderer) render(root *RootNode, data interface{}) (string, error) {
	ctx := newRenderContext(data)
	return r.renderNodes(root.Children, ctx)
}

func (r *renderer) renderNodes(nodes []Node, ctx *renderContext) (string, error) {
	var sb strings.Builder
	for _, node := range nodes {
		s, err := r.renderNode(node, ctx)
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	return sb.String(), nil
}

func (r *renderer) renderNode(node Node, ctx *renderContext) (string, error) {
	switch n := node.(type) {
	case *TextNode:
		return n.Text, nil
	case *VariableNode:
		return r.renderVariable(n, ctx)
	case *FunctionNode:
		return r.renderFunction(n, ctx)
	case *IfNode:
		return r.renderIf(n, ctx)
	case *RangeNode:
		return r.renderRange(n, ctx)
	default:
		return "", NewErrorf(node.Pos(), "unknown node type: %T", node)
	}
}

func (r *renderer) renderVariable(n *VariableNode, ctx *renderContext) (string, error) {
	val, err := r.resolvePath(n.Path, ctx, n.Pos())
	if err != nil {
		return "", err
	}
	s, err := toString(val)
	if err != nil {
		return "", NewErrorf(n.Pos(), "failed to convert variable to string: %v", err)
	}
	return s, nil
}

func (r *renderer) renderFunction(n *FunctionNode, ctx *renderContext) (string, error) {
	fn, ok := r.engine.GetFunction(n.Name)
	if !ok {
		return "", NewErrorf(n.Pos(), "undefined function: %s", n.Name)
	}

	var args []interface{}
	for _, argNode := range n.Args {
		switch an := argNode.(type) {
		case *TextNode:
			args = append(args, an.Text)
		case *VariableNode:
			val, err := r.resolvePath(an.Path, ctx, an.Pos())
			if err != nil {
				return "", err
			}
			args = append(args, val)
		case *FunctionNode:
			res, err := r.renderFunction(an, ctx)
			if err != nil {
				return "", err
			}
			args = append(args, res)
		default:
			return "", NewErrorf(n.Pos(), "unsupported argument node type: %T", argNode)
		}
	}

	result, err := fn(args...)
	if err != nil {
		argStrs := make([]string, len(args))
		for i, a := range args {
			argStrs[i] = fmt.Sprintf("%v", a)
		}
		return "", NewErrorf(n.Pos(), "function '%s' failed with args (%s): %v", n.Name, strings.Join(argStrs, ", "), err)
	}

	return result, nil
}

func (r *renderer) renderIf(n *IfNode, ctx *renderContext) (string, error) {
	condVal, err := r.evaluateCondition(n.Condition, ctx)
	if err != nil {
		return "", err
	}

	if condVal {
		return r.renderNodes(n.Body, ctx)
	} else if n.Else != nil {
		return r.renderNodes(n.Else, ctx)
	}
	return "", nil
}

func (r *renderer) evaluateCondition(node Node, ctx *renderContext) (bool, error) {
	switch n := node.(type) {
	case *VariableNode:
		val, err := r.resolvePath(n.Path, ctx, n.Pos())
		if err != nil {
			return false, err
		}
		return r.isTruthy(val), nil
	case *FunctionNode:
		result, err := r.renderFunction(n, ctx)
		if err != nil {
			return false, err
		}
		return len(strings.TrimSpace(result)) > 0, nil
	case *TextNode:
		return len(strings.TrimSpace(n.Text)) > 0, nil
	default:
		return false, NewErrorf(node.Pos(), "unsupported condition node type: %T", node)
	}
}

func (r *renderer) isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() > 0
	case reflect.Bool:
		return rv.Bool()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return false
		}
		return r.isTruthy(rv.Elem().Interface())
	case reflect.Invalid:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	}
	return true
}

func (r *renderer) renderRange(n *RangeNode, ctx *renderContext) (string, error) {
	var val interface{}
	var err error

	switch v := n.Variable.(type) {
	case *VariableNode:
		val, err = r.resolvePath(v.Path, ctx, v.Pos())
		if err != nil {
			return "", err
		}
	default:
		return "", NewErrorf(n.Pos(), "range variable must be a variable reference")
	}

	if val == nil {
		return "", nil
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return "", nil
		}
		rv = rv.Elem()
	}

	var sb strings.Builder

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i).Interface()
			itemCtx := ctx.withIndex(i).withValue(item)
			s, err := r.renderNodes(n.Body, itemCtx)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
	case reflect.Map:
		keys := rv.MapKeys()
		for i, key := range keys {
			val := rv.MapIndex(key).Interface()
			itemCtx := ctx.withIndex(i).withValue(val)
			s, err := r.renderNodes(n.Body, itemCtx)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
	default:
		return "", NewErrorf(n.Pos(), "range requires array, slice, or map, got %T", val)
	}

	return sb.String(), nil
}

func (r *renderer) resolvePath(path []string, ctx *renderContext, pos Position) (interface{}, error) {
	if len(path) == 0 {
		return ctx.current, nil
	}

	var current interface{}
	startIdx := 0

	if len(path) > 0 && path[0] == "Index" {
		if ctx.index == nil {
			if r.engine.config.StrictMissing {
				return nil, NewErrorf(pos, "Index is not available in this context")
			}
			return "", nil
		}
		if len(path) == 1 {
			return *ctx.index, nil
		}
		return nil, NewErrorf(pos, "Index cannot have sub-fields")
	}

	current = ctx.current

	for i := startIdx; i < len(path); i++ {
		field := path[i]
		if current == nil {
			if r.engine.config.StrictMissing {
				return nil, NewErrorf(pos, "cannot access field '%s' on nil", field)
			}
			return "", nil
		}

		var err error
		current, err = r.accessField(current, field, pos)
		if err != nil {
			if r.engine.config.StrictMissing {
				return nil, err
			}
			return "", nil
		}
	}

	return current, nil
}

func (r *renderer) accessField(v interface{}, field string, pos Position) (interface{}, error) {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil, nil
		}
		return r.accessField(rv.Elem().Interface(), field, pos)
	case reflect.Map:
		key := reflect.ValueOf(field)
		val := rv.MapIndex(key)
		if !val.IsValid() {
			if r.engine.config.StrictMissing {
				return nil, NewErrorf(pos, "map key not found: %s", field)
			}
			return "", nil
		}
		return val.Interface(), nil
	case reflect.Struct:
		f := rv.FieldByName(field)
		if !f.IsValid() {
			if r.engine.config.StrictMissing {
				return nil, NewErrorf(pos, "struct field not found: %s", field)
			}
			return "", nil
		}
		if !f.CanInterface() {
			if r.engine.config.StrictMissing {
				return nil, NewErrorf(pos, "cannot access unexported field: %s", field)
			}
			return "", nil
		}
		return f.Interface(), nil
	default:
		if r.engine.config.StrictMissing {
			return nil, NewErrorf(pos, "cannot access field '%s' on type %T", field, v)
		}
		return "", nil
	}
}

type renderContext struct {
	current interface{}
	index   *int
	parent  *renderContext
}

func newRenderContext(data interface{}) *renderContext {
	return &renderContext{current: data}
}

func (c *renderContext) withValue(v interface{}) *renderContext {
	return &renderContext{
		current: v,
		index:   c.index,
		parent:  c,
	}
}

func (c *renderContext) withIndex(i int) *renderContext {
	return &renderContext{
		current: c.current,
		index:   &i,
		parent:  c,
	}
}
