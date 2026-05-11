package openapi

import "encoding/json"

type OpenAPI struct {
	OpenAPI    string                 `json:"openapi" yaml:"openapi"`
	Info       Info                   `json:"info" yaml:"info"`
	Paths      map[string]PathItem    `json:"paths" yaml:"paths"`
	Components *Components            `json:"components,omitempty" yaml:"components,omitempty"`
	Raw        map[string]interface{} `json:"-" yaml:"-"`
}

type Info struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

type PathItem struct {
	Get        *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post       *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put        *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete     *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch      *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Options    *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	Head       *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	Trace      *Operation `json:"trace,omitempty" yaml:"trace,omitempty"`
	Parameters []RefOrParam `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Raw        map[string]interface{} `json:"-" yaml:"-"`
}

type Operation struct {
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []RefOrParam        `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RefOrRequestBody   `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]RefOrResponse `json:"responses" yaml:"responses"`
}

type Components struct {
	Schemas         map[string]Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Parameters      map[string]Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBodies   map[string]RequestBody `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Responses       map[string]Response `json:"responses,omitempty" yaml:"responses,omitempty"`
	Headers         map[string]Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

type Reference struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

type RefOrParam struct {
	Ref    string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Param  *Parameter  `json:"-" yaml:"-"`
}

func (r *RefOrParam) UnmarshalJSON(data []byte) error {
	var ref Reference
	if err := json.Unmarshal(data, &ref); err == nil && ref.Ref != "" {
		r.Ref = ref.Ref
		return nil
	}
	var param Parameter
	if err := json.Unmarshal(data, &param); err != nil {
		return err
	}
	r.Param = &param
	return nil
}

func (r RefOrParam) MarshalJSON() ([]byte, error) {
	if r.Ref != "" {
		return json.Marshal(Reference{Ref: r.Ref})
	}
	return json.Marshal(r.Param)
}

type RefOrRequestBody struct {
	Ref   string        `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Body  *RequestBody  `json:"-" yaml:"-"`
}

func (r *RefOrRequestBody) UnmarshalJSON(data []byte) error {
	var ref Reference
	if err := json.Unmarshal(data, &ref); err == nil && ref.Ref != "" {
		r.Ref = ref.Ref
		return nil
	}
	var body RequestBody
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}
	r.Body = &body
	return nil
}

func (r RefOrRequestBody) MarshalJSON() ([]byte, error) {
	if r.Ref != "" {
		return json.Marshal(Reference{Ref: r.Ref})
	}
	return json.Marshal(r.Body)
}

type RefOrResponse struct {
	Ref     string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Resp    *Response   `json:"-" yaml:"-"`
}

func (r *RefOrResponse) UnmarshalJSON(data []byte) error {
	var ref Reference
	if err := json.Unmarshal(data, &ref); err == nil && ref.Ref != "" {
		r.Ref = ref.Ref
		return nil
	}
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	r.Resp = &resp
	return nil
}

func (r RefOrResponse) MarshalJSON() ([]byte, error) {
	if r.Ref != "" {
		return json.Marshal(Reference{Ref: r.Ref})
	}
	return json.Marshal(r.Resp)
}

type Parameter struct {
	Name        string  `json:"name" yaml:"name"`
	In          string  `json:"in" yaml:"in"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type RequestBody struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType  `json:"content" yaml:"content"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
}

type Response struct {
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]RefOrHeader `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType  `json:"content,omitempty" yaml:"content,omitempty"`
}

type RefOrHeader struct {
	Ref     string   `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Header  *Header  `json:"-" yaml:"-"`
}

func (r *RefOrHeader) UnmarshalJSON(data []byte) error {
	var ref Reference
	if err := json.Unmarshal(data, &ref); err == nil && ref.Ref != "" {
		r.Ref = ref.Ref
		return nil
	}
	var header Header
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	r.Header = &header
	return nil
}

func (r RefOrHeader) MarshalJSON() ([]byte, error) {
	if r.Ref != "" {
		return json.Marshal(Reference{Ref: r.Ref})
	}
	return json.Marshal(r.Header)
}

type Header struct {
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type MediaType struct {
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type Schema struct {
	Type        string            `json:"type,omitempty" yaml:"type,omitempty"`
	Format      string            `json:"format,omitempty" yaml:"format,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required    []string          `json:"required,omitempty" yaml:"required,omitempty"`
	Items       *Schema           `json:"items,omitempty" yaml:"items,omitempty"`
	Ref         string            `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	AllOf       []Schema          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf       []Schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf       []Schema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Raw         map[string]interface{} `json:"-" yaml:"-"`
}

type SecurityScheme struct {
	Type         string `json:"type" yaml:"type"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Scheme       string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
}
