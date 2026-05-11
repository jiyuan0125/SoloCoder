package mockgen

import "os"

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func GenerateMock(source string, ifaceName string) (string, error) {
	parser := NewParser()
	info, err := parser.ExtractInterface(source, ifaceName)
	if err != nil {
		return "", err
	}
	generator := NewGenerator()
	return generator.Generate(info)
}

func ListInterfaces(source string) ([]string, error) {
	parser := NewParser()
	return parser.ListInterfaces(source)
}

func GenerateMockFromFile(filePath string, ifaceName string) (string, error) {
	source, err := readFile(filePath)
	if err != nil {
		return "", err
	}
	return GenerateMock(string(source), ifaceName)
}

func ListInterfacesFromFile(filePath string) ([]string, error) {
	source, err := readFile(filePath)
	if err != nil {
		return nil, err
	}
	return ListInterfaces(string(source))
}
