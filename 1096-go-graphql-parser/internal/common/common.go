package common

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string              `json:"operationName,omitempty"`
}

type GraphQLResponse struct {
	Data     interface{} `json:"data,omitempty"`
	Errors   []string  `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type SchemaResponse struct {
	Types []TypeInfo `json:"types"`
}

type TypeInfo struct {
	Name   string      `json:"name"`
	Fields []FieldInfo `json:"fields"`
}

type FieldInfo struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Args map[string]string `json:"args,omitempty"`
}

type VariablesSetRequest struct {
	Variables map[string]interface{} `json:"variables"`
}

type VariablesResponse struct {
	Variables map[string]interface{} `json:"variables"`
}

type ValidateResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type FormatResponse struct {
	Formatted string `json:"formatted"`
	Errors    []string `json:"errors,omitempty"`
}
