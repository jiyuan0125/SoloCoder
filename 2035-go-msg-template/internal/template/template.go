package template

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type TemplateManager struct {
	templatesDir string
}

func NewTemplateManager(templatesDir string) *TemplateManager {
	return &TemplateManager{templatesDir: templatesDir}
}

func (m *TemplateManager) TemplatesDir() string {
	return m.templatesDir
}

func (m *TemplateManager) EnsureDir() error {
	return os.MkdirAll(m.templatesDir, 0755)
}

func (m *TemplateManager) List() ([]string, error) {
	if err := m.EnsureDir(); err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(m.templatesDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".tpl") {
			names = append(names, strings.TrimSuffix(f.Name(), ".tpl"))
		}
	}
	return names, nil
}

func (m *TemplateManager) Exists(name string) (bool, error) {
	path := m.templatePath(name)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (m *TemplateManager) Create(name, content string) error {
	exists, err := m.Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmtError("template already exists: %s", name)
	}
	if err := m.EnsureDir(); err != nil {
		return err
	}
	path := m.templatePath(name)
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func (m *TemplateManager) Update(name, content string) error {
	exists, err := m.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmtError("template not found: %s", name)
	}
	path := m.templatePath(name)
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func (m *TemplateManager) Delete(name string) error {
	exists, err := m.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmtError("template not found: %s", name)
	}
	path := m.templatePath(name)
	return os.Remove(path)
}

func (m *TemplateManager) Read(name string) (string, error) {
	exists, err := m.Exists(name)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmtError("template not found: %s", name)
	}
	path := m.templatePath(name)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (m *TemplateManager) Render(name string, data map[string]interface{}, w io.Writer) error {
	content, err := m.Read(name)
	if err != nil {
		return err
	}

	filename := name + ".tpl"
	parser := NewParser(content, filename)
	nodes, err := parser.Parse()
	if err != nil {
		return err
	}

	renderer := NewRenderer()
	partials, err := m.loadPartials()
	if err != nil {
		return err
	}
	for pName, pContent := range partials {
		if err := renderer.AddPartial(pName, pContent, pName+".tpl"); err != nil {
			return err
		}
	}

	return renderer.Render(nodes, data, w)
}

func (m *TemplateManager) loadPartials() (map[string]string, error) {
	partials := make(map[string]string)
	names, err := m.List()
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		if strings.HasPrefix(name, "_") {
			content, err := m.Read(name)
			if err != nil {
				return nil, err
			}
			partials[name] = content
		}
	}
	return partials, nil
}

func (m *TemplateManager) templatePath(name string) string {
	return filepath.Join(m.templatesDir, name+".tpl")
}

func ReadJSONData(dataFile string) (map[string]interface{}, error) {
	var data []byte
	var err error

	if dataFile == "-" || dataFile == "" {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = ioutil.ReadFile(dataFile)
		if err != nil {
			return nil, err
		}
	}

	var result map[string]interface{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
	}
	if result == nil {
		result = make(map[string]interface{})
	}
	return result, nil
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}

func fmtError(format string, args ...interface{}) error {
	return &simpleError{msg: sprintf(format, args...)}
}

func sprintf(format string, args ...interface{}) string {
	return fmtSprintf(format, args...)
}
