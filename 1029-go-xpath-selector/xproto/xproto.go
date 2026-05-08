package xproto

type NamespaceMapping struct {
	Prefix string `json:"prefix"`
	URI    string `json:"uri"`
}

type QueryRequest struct {
	XML        string             `json:"xml"`
	XPath      string             `json:"xpath"`
	Namespaces []NamespaceMapping `json:"namespaces,omitempty"`
	ReturnType string             `json:"return_type,omitempty"`
}

type NodeInfo struct {
	Type       string            `json:"type"`
	LocalName  string            `json:"local_name,omitempty"`
	Prefix     string            `json:"prefix,omitempty"`
	Namespace  string            `json:"namespace,omitempty"`
	Text       string            `json:"text,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type QueryResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	
	ResultType string      `json:"result_type,omitempty"`
	Nodes      []NodeInfo  `json:"nodes,omitempty"`
	String     string      `json:"string,omitempty"`
	Number     *float64    `json:"number,omitempty"`
	Boolean    *bool       `json:"boolean,omitempty"`
}

const (
	ReturnTypeDefault = ""
	ReturnTypeNodes   = "nodes"
	ReturnTypeString  = "string"
	ReturnTypeNumber  = "number"
	ReturnTypeBoolean = "boolean"
)

const (
	ResultTypeNodeSet = "nodeset"
	ResultTypeString  = "string"
	ResultTypeNumber  = "number"
	ResultTypeBoolean = "boolean"
)

const (
	NodeTypeElement = "element"
	NodeTypeText    = "text"
	NodeTypeAttribute = "attribute"
	NodeTypeDocument = "document"
)

func (nr *QueryResult) AddNode(node NodeInfo) {
	nr.Nodes = append(nr.Nodes, node)
}

func (nr *QueryResult) SetString(s string) {
	nr.ResultType = ResultTypeString
	nr.String = s
}

func (nr *QueryResult) SetNumber(n float64) {
	nr.ResultType = ResultTypeNumber
	nr.Number = &n
}

func (nr *QueryResult) SetBoolean(b bool) {
	nr.ResultType = ResultTypeBoolean
	nr.Boolean = &b
}
