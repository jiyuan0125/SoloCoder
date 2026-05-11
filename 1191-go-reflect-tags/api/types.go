package api

import "reflect-tags/tagparser"

type RegisterRequest struct {
	Struct tagparser.StructDef `json:"struct"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type QueryRequest struct {
	StructName string `json:"struct_name"`
	FieldPath  string `json:"field_path"`
}

type QueryResponse struct {
	Success bool                  `json:"success"`
	Field   *tagparser.FieldTags  `json:"field,omitempty"`
	All     []tagparser.FieldTags `json:"all,omitempty"`
	Message string                `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
