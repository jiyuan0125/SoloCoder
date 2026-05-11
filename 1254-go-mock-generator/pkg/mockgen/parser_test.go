package mockgen

import "testing"

const testSource = `
package testpkg

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

type ReadCloser interface {
	Reader
	Closer
}

type Logger interface {
	Log(format string, args ...interface{})
	Debug(msg string)
}
`

func TestListInterfaces(t *testing.T) {
	parser := NewParser()
	ifaces, err := parser.ListInterfaces(testSource)
	if err != nil {
		t.Fatalf("ListInterfaces failed: %v", err)
	}

	expected := map[string]bool{
		"Reader":     true,
		"Closer":     true,
		"ReadCloser": true,
		"Logger":     true,
	}

	if len(ifaces) != len(expected) {
		t.Fatalf("expected %d interfaces, got %d", len(expected), len(ifaces))
	}

	for _, name := range ifaces {
		if !expected[name] {
			t.Errorf("unexpected interface: %s", name)
		}
	}
}

func TestExtractInterface(t *testing.T) {
	parser := NewParser()
	info, err := parser.ExtractInterface(testSource, "ReadCloser")
	if err != nil {
		t.Fatalf("ExtractInterface failed: %v", err)
	}

	if info.Package != "testpkg" {
		t.Errorf("expected package 'testpkg', got '%s'", info.Package)
	}

	if info.Name != "ReadCloser" {
		t.Errorf("expected name 'ReadCloser', got '%s'", info.Name)
	}

	if len(info.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(info.Methods))
	}

	methodNames := map[string]bool{}
	for _, m := range info.Methods {
		methodNames[m.Name] = true
	}

	if !methodNames["Read"] {
		t.Error("expected Read method")
	}
	if !methodNames["Close"] {
		t.Error("expected Close method")
	}
}

func TestExtractInterfaceWithVariadic(t *testing.T) {
	parser := NewParser()
	info, err := parser.ExtractInterface(testSource, "Logger")
	if err != nil {
		t.Fatalf("ExtractInterface failed: %v", err)
	}

	var logMethod *Method
	for i := range info.Methods {
		if info.Methods[i].Name == "Log" {
			logMethod = &info.Methods[i]
			break
		}
	}

	if logMethod == nil {
		t.Fatal("Log method not found")
	}

	if !logMethod.IsVariadic {
		t.Error("expected Log method to be variadic")
	}

	if len(logMethod.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(logMethod.Params))
	}
}
