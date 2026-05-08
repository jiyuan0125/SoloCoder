package template

import "testing"

func TestVariableReplacement(t *testing.T) {
	engine := New(nil)

	result, err := engine.Render("Hello, {{.Name}}!", map[string]interface{}{"Name": "World"})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %q", result)
	}
}

func TestNestedAccess(t *testing.T) {
	engine := New(nil)
	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name": "Alice",
			"Age":  30,
		},
	}

	result, err := engine.Render("{{.User.Name}} is {{.User.Age}}", data)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Alice is 30" {
		t.Fatalf("expected 'Alice is 30', got %q", result)
	}
}

func TestBuiltinUpper(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("{{upper .Name}}", map[string]interface{}{"Name": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if result != "HELLO" {
		t.Fatalf("expected 'HELLO', got %q", result)
	}
}

func TestBuiltinDefault(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("{{default .Value 'default'}}", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if result != "default" {
		t.Fatalf("expected 'default', got %q", result)
	}
}

func TestIfCondition(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("{{if .Show}}yes{{end}}", map[string]interface{}{"Show": true})
	if err != nil {
		t.Fatal(err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}
}

func TestIfElse(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("{{if .Show}}yes{{else}}no{{end}}", map[string]interface{}{"Show": false})
	if err != nil {
		t.Fatal(err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}
}

func TestRangeLoop(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("{{range .Items}}{{.}} {{end}}", map[string]interface{}{
		"Items": []string{"a", "b", "c"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "a b c " {
		t.Fatalf("expected 'a b c ', got %q", result)
	}
}

func TestRangeIndex(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("{{range .Items}}{{.Index}}:{{.}};{{end}}", map[string]interface{}{
		"Items": []string{"a", "b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "0:a;1:b;" {
		t.Fatalf("expected '0:a;1:b;', got %q", result)
	}
}

func TestUnclosedIf(t *testing.T) {
	engine := New(nil)
	_, err := engine.Render("{{if .X}}hello", nil)
	if err == nil {
		t.Fatal("expected error for unclosed if")
	}
}

func TestCustomFunction(t *testing.T) {
	engine := New(nil)
	engine.RegisterFunction("greet", func(args ...interface{}) (string, error) {
		if len(args) != 1 {
			return "", nil
		}
		return "Hello, " + args[0].(string), nil
	})

	result, err := engine.Render("{{greet .Name}}", map[string]interface{}{"Name": "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello, Alice" {
		t.Fatalf("expected 'Hello, Alice', got %q", result)
	}
}

func TestEmptyRange(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("Items:{{range .Items}}{{.}}{{end}}", map[string]interface{}{
		"Items": []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Items:" {
		t.Fatalf("expected 'Items:', got %q", result)
	}
}

func TestNilValue(t *testing.T) {
	engine := New(nil)
	result, err := engine.Render("Value: {{.X}}", map[string]interface{}{"X": nil})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Value: " {
		t.Fatalf("expected 'Value: ', got %q", result)
	}
}
