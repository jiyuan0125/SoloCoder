package main

import (
	"fmt"
	"template-engine/template"
)

func main() {
	delims := template.Delims{Left: "{{", Right: "}}"}
	input := "Hello, {{.Name}}!"
	fmt.Println("Input:", input)
	
	l := template.NewLexerDebug(input, delims)
	tokens := l.Lex()
	for i, tok := range tokens {
		fmt.Printf("%d: Type=%v, Value=%q, Line=%d, Col=%d\n", 
			i, tok.Type, tok.Value, tok.Position.Line, tok.Position.Column)
	}
	
	engine := template.New(nil)
	result, err := engine.Render("Hello, {{.Name}}!", map[string]interface{}{"Name": "World"})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result)
	}
}
