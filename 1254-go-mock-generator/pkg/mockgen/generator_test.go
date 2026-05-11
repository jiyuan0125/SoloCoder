package mockgen

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const testIface = `
package testpkg

type UserStore interface {
	Get(id int) (string, error)
	Save(id int, name string) error
	List() ([]string, error)
	Close()
}
`

func TestGenerateMock(t *testing.T) {
	parser := NewParser()
	info, err := parser.ExtractInterface(testIface, "UserStore")
	if err != nil {
		t.Fatalf("ExtractInterface failed: %v", err)
	}

	generator := NewGenerator()
	code, err := generator.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	tmpDir, err := ioutil.TempDir("", "mockgen-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mockFile := tmpDir + "/mock_UserStore.go"
	ifaceFile := tmpDir + "/userstore.go"

	if err := ioutil.WriteFile(mockFile, []byte(code), 0644); err != nil {
		t.Fatalf("write mock file: %v", err)
	}

	ifaceCode := testIface + `

func UseStore(store UserStore) {
	store.Get(1)
	store.Save(1, "test")
	store.List()
	store.Close()
}
`
	if err := ioutil.WriteFile(ifaceFile, []byte(ifaceCode), 0644); err != nil {
		t.Fatalf("write iface file: %v", err)
	}

	cmd := exec.Command("go", "build", ".")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GO111MODULE=off")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Generated code:\n%s", code)
		t.Fatalf("build generated code failed: %v\n%s", err, string(output))
	}
}

const testIfaceVariadic = `
package testpkg

type Logger interface {
	Log(format string, args ...interface{})
	Debug(msg string)
	PrintItems(items ...string) error
}
`

func TestGenerateMockWithVariadic(t *testing.T) {
	parser := NewParser()
	info, err := parser.ExtractInterface(testIfaceVariadic, "Logger")
	if err != nil {
		t.Fatalf("ExtractInterface failed: %v", err)
	}

	generator := NewGenerator()
	code, err := generator.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(code, "func (m *MockLogger) Log(format string, args ...interface{})") {
		t.Errorf("expected variadic signature 'Log(format string, args ...interface{})' not found")
		t.Logf("Generated code:\n%s", code)
	}

	if !strings.Contains(code, "func (m *MockLogger) PrintItems(items ...string) error") {
		t.Errorf("expected variadic signature 'PrintItems(items ...string) error' not found")
		t.Logf("Generated code:\n%s", code)
	}

	tmpDir, err := ioutil.TempDir("", "mockgen-variadic-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mockFile := tmpDir + "/mock_Logger.go"
	ifaceFile := tmpDir + "/logger.go"

	if err := ioutil.WriteFile(mockFile, []byte(code), 0644); err != nil {
		t.Fatalf("write mock file: %v", err)
	}

	ifaceCode := testIfaceVariadic + `

func UseLogger(logger Logger) {
	logger.Log("test %d", 1)
	logger.Debug("msg")
	logger.PrintItems("a", "b", "c")
}
`
	if err := ioutil.WriteFile(ifaceFile, []byte(ifaceCode), 0644); err != nil {
		t.Fatalf("write iface file: %v", err)
	}

	cmd := exec.Command("go", "build", ".")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GO111MODULE=off")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Generated code:\n%s", code)
		t.Fatalf("build generated code failed: %v\n%s", err, string(output))
	}
}
