package template

import (
	"fmt"
)

type TemplateError struct {
	Message  string
	Position Position
	Cause    error
}

func (e *TemplateError) Error() string {
	loc := fmt.Sprintf("line %d, column %d", e.Position.Line, e.Position.Column)
	if e.Position.Source != "" {
		loc = e.Position.Source + ": " + loc
	}
	if e.Cause != nil {
		return fmt.Sprintf("template error at %s: %s (%v)", loc, e.Message, e.Cause)
	}
	return fmt.Sprintf("template error at %s: %s", loc, e.Message)
}

func NewError(pos Position, message string) *TemplateError {
	return &TemplateError{Message: message, Position: pos}
}

func NewErrorf(pos Position, format string, args ...interface{}) *TemplateError {
	return &TemplateError{Message: fmt.Sprintf(format, args...), Position: pos}
}

type Config struct {
	StrictMissing bool
	Delims        Delims
}

type Delims struct {
	Left  string
	Right string
}

func DefaultConfig() *Config {
	return &Config{
		StrictMissing: false,
		Delims: Delims{
			Left:  "{{",
			Right: "}}",
		},
	}
}

type Engine struct {
	config    *Config
	functions map[string]Function
}

type Function func(args ...interface{}) (string, error)

func New(config *Config) *Engine {
	if config == nil {
		config = DefaultConfig()
	}
	e := &Engine{
		config:    config,
		functions: make(map[string]Function),
	}
	registerBuiltins(e)
	return e
}

func (e *Engine) RegisterFunction(name string, fn Function) error {
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}
	e.functions[name] = fn
	return nil
}

func (e *Engine) HasFunction(name string) bool {
	_, ok := e.functions[name]
	return ok
}

func (e *Engine) GetFunction(name string) (Function, bool) {
	fn, ok := e.functions[name]
	return fn, ok
}

func (e *Engine) Config() *Config {
	return e.config
}

func (e *Engine) Parse(template string) (*ParsedTemplate, error) {
	parser := newParser(e, template)
	return parser.parse()
}

func (e *Engine) Render(template string, data interface{}) (string, error) {
	parsed, err := e.Parse(template)
	if err != nil {
		return "", err
	}
	return parsed.Render(data)
}

type ParsedTemplate struct {
	engine *Engine
	root   *RootNode
	source string
}

func newParsedTemplate(engine *Engine, root *RootNode, source string) *ParsedTemplate {
	return &ParsedTemplate{
		engine: engine,
		root:   root,
		source: source,
	}
}

func (p *ParsedTemplate) Render(data interface{}) (string, error) {
	renderer := newRenderer(p.engine)
	return renderer.render(p.root, data)
}
