package ical

import (
	"io"
)

type ICal struct {
	parser    *Parser
	generator *Generator
}

func New() *ICal {
	return &ICal{
		parser:    NewParser(),
		generator: NewGenerator(),
	}
}

func (ic *ICal) Parse(r io.Reader) (*Calendar, error) {
	return ic.parser.Parse(r)
}

func (ic *ICal) Generate(cal *Calendar, w io.Writer) error {
	return ic.generator.Generate(cal, w)
}
