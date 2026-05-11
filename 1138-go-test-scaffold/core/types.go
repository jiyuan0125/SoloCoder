package core

type FieldInfo struct {
	Name       string
	Type       string
	IsEmbedded bool
	JSONTag    string
	IsExported bool
}

type StructInfo struct {
	Name       string
	Fields     []*FieldInfo
	Methods    []*MethodInfo
	Interfaces []string
}

type InterfaceInfo struct {
	Name    string
	Methods []*MethodInfo
}

type MethodInfo struct {
	Name       string
	Params     []string
	Results    []string
	IsExported bool
}

type ParsedSource struct {
	PackageName string
	Structs     map[string]*StructInfo
	Interfaces  map[string]*InterfaceInfo
}

type GeneratedTest struct {
	FileName string
	Code     string
}

type TestCase struct {
	Name     string
	Input    string
	Expected string
}

type TestFunction struct {
	Name     string
	Cases    []*TestCase
	IsMethod bool
	Receiver string
}

type Preview struct {
	FileName       string
	TestFunctions  []*TestFunction
	Structs        []string
}
