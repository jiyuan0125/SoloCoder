package common

type JsonTagInfo struct {
	Name      string `json:"name"`
	OmitEmpty bool   `json:"omitempty"`
	String    bool   `json:"string"`
	Ignored   bool   `json:"ignored"`
}

type DbIndexType string

const (
	DbIndexNone   DbIndexType = ""
	DbIndexPK     DbIndexType = "pk"
	DbIndexUnique DbIndexType = "unique"
	DbIndexIndex  DbIndexType = "index"
)

type DbTagInfo struct {
	Name      string      `json:"name"`
	IndexType DbIndexType `json:"indexType"`
	Ignored   bool        `json:"ignored"`
}

type ValidateRule struct {
	Name   string `json:"name"`
	Value  string `json:"value,omitempty"`
	Number int64  `json:"number,omitempty"`
}

type ValidateTagInfo struct {
	Rules []ValidateRule `json:"rules"`
}

type FieldTagInfo struct {
	FieldPath string           `json:"fieldPath"`
	FieldType string           `json:"fieldType"`
	Json      *JsonTagInfo     `json:"json,omitempty"`
	Db        *DbTagInfo       `json:"db,omitempty"`
	Validate  *ValidateTagInfo `json:"validate,omitempty"`
}

type FieldDesc struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	IsAnonymous bool      `json:"isAnonymous"`
	IsNested    bool      `json:"isNested"`
	Tags       map[string]string `json:"tags"`
	Fields     []FieldDesc  `json:"fields,omitempty"`
}

type RegisterRequest struct {
	Name   string      `json:"name"`
	Fields []FieldDesc `json:"fields"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type QueryRequest struct {
	StructName string `json:"structName"`
	FieldPath  string `json:"fieldPath"`
}

type QueryResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    *FieldTagInfo `json:"data,omitempty"`
}
