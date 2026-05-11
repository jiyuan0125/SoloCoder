package tagparser

type JSONTag struct {
	Name      string `json:"name"`
	OmitEmpty bool   `json:"omitempty"`
	String    bool   `json:"string"`
}

type DBTag struct {
	Column    string   `json:"column"`
	IndexType string   `json:"index_type"`
}

type ValidateTag struct {
	Required bool    `json:"required"`
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	Email    bool    `json:"email"`
	Regex    string  `json:"regex,omitempty"`
}

type FieldTags struct {
	Path     string      `json:"path"`
	JSON     *JSONTag    `json:"json,omitempty"`
	DB       *DBTag      `json:"db,omitempty"`
	Validate *ValidateTag `json:"validate,omitempty"`
}

type StructFieldDef struct {
	Name     string           `json:"name"`
	Type     string           `json:"type"`
	Tags     map[string]string `json:"tags"`
	IsAnonymous bool          `json:"is_anonymous,omitempty"`
	Fields   []StructFieldDef `json:"fields,omitempty"`
}

type StructDef struct {
	Name   string           `json:"name"`
	Fields []StructFieldDef `json:"fields"`
}
