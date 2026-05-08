package main

import (
	"fmt"
	"template-engine/template"
)

func main() {
	testSimpleVariable()
	testFunction()
	testIf()
}

func testSimpleVariable() {
	fmt.Println("=== Testing simple variable ===")
	engine := template.New(nil)
	result, err := engine.Render("Hello, {{.Name}}!", map[string]interface{}{"Name": "World"})
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("Result:", result)
	}
}

func testFunction() {
	fmt.Println("\n=== Testing function ===")
	engine := template.New(nil)
	result, err := engine.Render("{{upper .Name}}", map[string]interface{}{"Name": "hello"})
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("Result:", result)
	}
}

func testIf() {
	fmt.Println("\n=== Testing if ===")
	engine := template.New(nil)
	result, err := engine.Render("{{if .Show}}yes{{end}}", map[string]interface{}{"Show": true})
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("Result:", result)
	}
}
