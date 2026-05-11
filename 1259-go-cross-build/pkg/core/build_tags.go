package core

import (
	"fmt"
	"regexp"
	"strings"
)

type TagExpr interface {
	Evaluate(tags map[string]bool) bool
}

type TagAnd struct {
	Left, Right TagExpr
}

func (e *TagAnd) Evaluate(tags map[string]bool) bool {
	return e.Left.Evaluate(tags) && e.Right.Evaluate(tags)
}

type TagOr struct {
	Left, Right TagExpr
}

func (e *TagOr) Evaluate(tags map[string]bool) bool {
	return e.Left.Evaluate(tags) || e.Right.Evaluate(tags)
}

type TagNot struct {
	Inner TagExpr
}

func (e *TagNot) Evaluate(tags map[string]bool) bool {
	return !e.Inner.Evaluate(tags)
}

type TagName struct {
	Name string
}

func (e *TagName) Evaluate(tags map[string]bool) bool {
	return tags[e.Name]
}

func getPlatformTags(goos, goarch string) map[string]bool {
	tags := make(map[string]bool)
	tags[goos] = true
	tags[goarch] = true

	osTags := map[string][]string{
		"aix":       {"aix"},
		"android":   {"android", "unix"},
		"darwin":    {"darwin", "unix", "bsd"},
		"dragonfly": {"dragonfly", "unix", "bsd"},
		"freebsd":   {"freebsd", "unix", "bsd"},
		"illumos":   {"illumos", "unix"},
		"ios":       {"ios"},
		"js":        {"js"},
		"linux":     {"linux", "unix"},
		"netbsd":    {"netbsd", "unix", "bsd"},
		"openbsd":   {"openbsd", "unix", "bsd"},
		"plan9":     {"plan9"},
		"solaris":   {"solaris", "unix"},
		"windows":   {"windows"},
	}

	for _, tag := range osTags[goos] {
		tags[tag] = true
	}

	archTags := map[string][]string{
		"386":          {"386", "i386"},
		"amd64":        {"amd64", "x86_64"},
		"amd64p32":     {"amd64p32"},
		"arm":          {"arm", "arm32"},
		"armbe":        {"armbe", "arm32"},
		"arm64":        {"arm64", "arm64"},
		"arm64be":      {"arm64be", "arm64"},
		"mips":         {"mips", "mips32"},
		"mipsle":       {"mipsle", "mips32"},
		"mips64":       {"mips64", "mips64"},
		"mips64le":     {"mips64le", "mips64"},
		"ppc64":        {"ppc64"},
		"ppc64le":      {"ppc64le"},
		"riscv64":      {"riscv64"},
		"s390x":        {"s390x"},
		"wasm":         {"wasm"},
	}

	for _, tag := range archTags[goarch] {
		tags[tag] = true
	}

	return tags
}

func ParseTagExpr(expr string) (TagExpr, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty tag expression")
	}
	parser := &tagParser{input: expr, pos: 0}
	return parser.parseOr()
}

type tagParser struct {
	input string
	pos   int
}

func (p *tagParser) skipWhitespace() {
	for p.pos < len(p.input) && p.input[p.pos] == ' ' {
		p.pos++
	}
}

func (p *tagParser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *tagParser) consume() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	b := p.input[p.pos]
	p.pos++
	return b
}

func (p *tagParser) parseOr() (TagExpr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.peek() == ',' {
		p.consume()
		p.skipWhitespace()
		right, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		return &TagOr{Left: left, Right: right}, nil
	}
	return left, nil
}

func (p *tagParser) parseAnd() (TagExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.peek() == ' ' {
		p.consume()
		p.skipWhitespace()
		if p.peek() != 0 && p.peek() != ',' && p.peek() != ')' {
			right, err := p.parseAnd()
			if err != nil {
				return nil, err
			}
			return &TagAnd{Left: left, Right: right}, nil
		}
	}
	return left, nil
}

func (p *tagParser) parseUnary() (TagExpr, error) {
	p.skipWhitespace()
	if p.peek() == '!' {
		p.consume()
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &TagNot{Inner: inner}, nil
	}
	if p.peek() == '(' {
		p.consume()
		inner, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.consume() != ')' {
			return nil, fmt.Errorf("expected ')'")
		}
		return inner, nil
	}
	return p.parseName()
}

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+`)

func (p *tagParser) parseName() (TagExpr, error) {
	p.skipWhitespace()
	remaining := p.input[p.pos:]
	match := nameRegex.FindString(remaining)
	if match == "" {
		return nil, fmt.Errorf("expected tag name at position %d", p.pos)
	}
	p.pos += len(match)
	return &TagName{Name: match}, nil
}

func EvaluateBuildTag(tagLine, goos, goarch string) (bool, error) {
	expr, err := ParseTagExpr(tagLine)
	if err != nil {
		return false, err
	}
	tags := getPlatformTags(goos, goarch)
	return expr.Evaluate(tags), nil
}
