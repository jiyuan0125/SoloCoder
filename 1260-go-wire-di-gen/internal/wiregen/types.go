package wiregen

type TypeKind int

const (
	TypeKindUnknown TypeKind = iota
	TypeKindStruct
	TypeKindInterface
)

type TypeInfo struct {
	Name    string
	Kind    TypeKind
	Methods []string
}

type ParamInfo struct {
	Name string
	Type string
}

type Provider struct {
	ID                string
	Name              string
	Params            []ParamInfo
	ReturnType        string
	ReturnsError      bool
	IsExplicitProvider bool
}

type PackageInfo struct {
	Name      string
	Providers map[string]*Provider
	Types     map[string]*TypeInfo
}

type ExternalParam struct {
	Name string
	Type string
}

type ResolvedProvider struct {
	Provider   *Provider
	VarName    string
	External   bool
	ExternalParam *ExternalParam
}
