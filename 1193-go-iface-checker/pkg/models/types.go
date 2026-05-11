package models

type MethodParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsVariadic bool `json:"is_variadic,omitempty"`
}

type Method struct {
	Name       string        `json:"name"`
	Params     []MethodParam `json:"params"`
	Returns    []string      `json:"returns"`
	IsPointerReceiver bool `json:"is_pointer_receiver,omitempty"`
}

type StructType struct {
	Name       string   `json:"name"`
	Methods    []Method `json:"methods"`
	EmbeddedFields []EmbeddedField `json:"embedded_fields,omitempty"`
}

type EmbeddedField struct {
	Type   string `json:"type"`
	IsPointer bool `json:"is_pointer,omitempty"`
	IsInterface bool `json:"is_interface,omitempty"`
}

type InterfaceType struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
}

type RegisterRequest struct {
	Structs    []StructType   `json:"structs,omitempty"`
	Interfaces []InterfaceType `json:"interfaces,omitempty"`
}

type CheckRequest struct {
	StructName    string `json:"struct_name"`
	InterfaceName string `json:"interface_name"`
	UseValueReceiver bool `json:"use_value_receiver,omitempty"`
}

type MismatchDetail struct {
	IssueType    string `json:"issue_type"`
	Expected     string `json:"expected"`
	Actual       string `json:"actual"`
}

type MethodCheckResult struct {
	MethodName string `json:"method_name"`
	IsMissing  bool   `json:"is_missing"`
	IsMismatch bool   `json:"is_mismatch"`
	Details    []MismatchDetail `json:"details,omitempty"`
}

type CheckResponse struct {
	Satisfies   bool               `json:"satisfies"`
	MissingMethods   []MethodCheckResult `json:"missing_methods"`
	MismatchedMethods []MethodCheckResult `json:"mismatched_methods"`
	Message        string             `json:"message"`
}

type CheckAllRequest struct {
	TargetStruct   string        `json:"target_struct"`
	Structs        []StructType  `json:"structs"`
	Interfaces     []InterfaceType `json:"interfaces"`
	UseValueReceiver bool        `json:"use_value_receiver,omitempty"`
}

type InterfaceCheckResult struct {
	InterfaceName string `json:"interface_name"`
	CheckResponse
}

type CheckAllResponse struct {
	Results []InterfaceCheckResult `json:"results"`
}
