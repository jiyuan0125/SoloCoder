package template

import (
	"strings"
	"testing"
)

func TestVariableReplacement(t *testing.T) {
	template := "Hello, {{name}}!"
	data := map[string]interface{}{
		"name": "World",
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "Hello, World!"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestDefaultValue(t *testing.T) {
	template := "Hello, {{name|Guest}}!"
	data := map[string]interface{}{}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "Hello, Guest!"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestMissingVariable(t *testing.T) {
	template := "Hello, {{name}}!"
	data := map[string]interface{}{}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "Hello, {{name}}!"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
	
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestNestedObject(t *testing.T) {
	template := "City: {{user.address.city}}"
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"address": map[string]interface{}{
				"city": "Beijing",
			},
		},
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "City: Beijing"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestIfConditionTrue(t *testing.T) {
	template := "{{if name}}Name exists: {{name}}{{end}}"
	data := map[string]interface{}{
		"name": "John",
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "Name exists: John"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestIfConditionFalse(t *testing.T) {
	template := "{{if name}}Name exists: {{name}}{{end}}"
	data := map[string]interface{}{}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := ""
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestIfElseTrue(t *testing.T) {
	template := "{{if name}}有名字{{else}}没名字{{end}}"
	data := map[string]interface{}{
		"name": "John",
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "有名字"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestIfElseFalse(t *testing.T) {
	template := "{{if name}}有名字{{else}}没名字{{end}}"
	data := map[string]interface{}{}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "没名字"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestEachLoop(t *testing.T) {
	template := "{{each items}}- {{.name}}: {{.value}}\n{{end}}"
	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{"name": "A", "value": 1},
			{"name": "B", "value": 2},
		},
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "- A: 1\n- B: 2\n"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestCurlyBraceInValue(t *testing.T) {
	template := "JSON: {{json}}"
	data := map[string]interface{}{
		"json": `{"key": "value"}`,
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "JSON: {\"key\": \"value\"}"
	if result.Content != expected {
		t.Errorf("Expected %q, got %q", expected, result.Content)
	}
}

func TestSyntaxErrorUnclosedIf(t *testing.T) {
	template := "{{if name}}Hello"
	data := map[string]interface{}{
		"name": "John",
	}
	
	_, err := RenderString(template, data)
	if err == nil {
		t.Error("Expected syntax error, got none")
	}
}

func TestSyntaxErrorUnexpectedEnd(t *testing.T) {
	template := "Hello{{end}}"
	data := map[string]interface{}{}
	
	_, err := RenderString(template, data)
	if err == nil {
		t.Error("Expected syntax error, got none")
	}
}

func TestComplexTemplate(t *testing.T) {
	template := `Dear {{name|Customer}},
{{if orders}}
Your orders:
{{each orders}}  - {{.product}}: ${{.price}}
{{end}}{{else}}No orders found.
{{end}}Thank you!`
	
	data := map[string]interface{}{
		"name": "Alice",
		"orders": []map[string]interface{}{
			{"product": "Apple", "price": 5},
			{"product": "Banana", "price": 3},
		},
	}
	
	result, err := RenderString(template, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(result.Content, "Dear Alice") {
		t.Errorf("Expected 'Dear Alice' in result, got:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Your orders:") {
		t.Errorf("Expected 'Your orders:' in result, got:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Apple: $5") {
		t.Errorf("Expected 'Apple: $5' in result, got:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Banana: $3") {
		t.Errorf("Expected 'Banana: $3' in result, got:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Thank you!") {
		t.Errorf("Expected 'Thank you!' in result, got:\n%s", result.Content)
	}
}
