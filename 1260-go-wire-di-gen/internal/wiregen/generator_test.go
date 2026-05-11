package wiregen

import (
	"strings"
	"testing"
)

func TestGenerateNormal(t *testing.T) {
	code, err := Generate("testdata/normal")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(code, "package normal") {
		t.Errorf("Generated code should contain 'package normal', got:\n%s", code)
	}

	if !strings.Contains(code, "func Initialize(") {
		t.Errorf("Generated code should contain 'func Initialize', got:\n%s", code)
	}

	if !strings.Contains(code, "dsn string") {
		t.Errorf("Generated code should have dsn parameter, got:\n%s", code)
	}

	if !strings.Contains(code, "NewDatabase") {
		t.Errorf("Generated code should call NewDatabase, got:\n%s", code)
	}

	if !strings.Contains(code, "NewRepository") {
		t.Errorf("Generated code should call NewRepository, got:\n%s", code)
	}

	if !strings.Contains(code, "NewService") {
		t.Errorf("Generated code should call NewService, got:\n%s", code)
	}

	t.Logf("Generated code:\n%s", code)
}

func TestGenerateCircular(t *testing.T) {
	_, err := Generate("testdata/circular")
	if err == nil {
		t.Fatal("Expected error for circular dependency, got nil")
	}

	if !IsCircularDependencyError(err) {
		t.Errorf("Expected CircularDependencyError, got: %T: %v", err, err)
	}

	path := GetCircularDependencyPath(err)
	if len(path) < 2 {
		t.Errorf("Expected cycle path with at least 2 elements, got: %v", path)
	}

	t.Logf("Circular dependency detected: %v", path)
}

func TestGenerateWithError(t *testing.T) {
	code, err := Generate("testdata/error")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(code, "error") {
		t.Errorf("Generated code should handle error, got:\n%s", code)
	}

	if !strings.Contains(code, "return nil, err") {
		t.Errorf("Generated code should return error, got:\n%s", code)
	}

	t.Logf("Generated code:\n%s", code)
}

func TestGenerateMultiProvider(t *testing.T) {
	code, err := Generate("testdata/multi")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.Contains(code, "NewMemCache") && !strings.Contains(code, "NewRedisCache") {
		t.Errorf("Should prefer explicit provider (NewRedisCache with +wire:provider), got:\n%s", code)
	}

	t.Logf("Generated code:\n%s", code)
}

func TestIsBasicType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"string", "string", true},
		{"int", "int", true},
		{"int64", "int64", true},
		{"bool", "bool", true},
		{"float64", "float64", true},
		{"custom", "MyType", false},
		{"pointer", "*string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBasicType(tt.typeName); got != tt.want {
				t.Errorf("IsBasicType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestIsContextType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"context", "context.Context", true},
		{"string", "string", false},
		{"other", "other.Context", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContextType(tt.typeName); got != tt.want {
				t.Errorf("IsContextType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}
