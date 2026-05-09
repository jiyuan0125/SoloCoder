package api

import "encoding/base64"

type DecodeRequest struct {
	Data string `json:"data"`
}

type DecodeResponse struct {
	Success bool         `json:"success"`
	Error   string       `json:"error,omitempty"`
	Result  *NodeJSON    `json:"result,omitempty"`
	Dump    string       `json:"dump,omitempty"`
}

type NodeJSON struct {
	Tag       TagJSON      `json:"tag"`
	Length    int          `json:"length"`
	ValueHex  string       `json:"valueHex,omitempty"`
	ValueStr  string       `json:"valueStr,omitempty"`
	OID       string       `json:"oid,omitempty"`
	Children  []*NodeJSON  `json:"children,omitempty"`
}

type TagJSON struct {
	Class       string `json:"class"`
	Constructed bool   `json:"constructed"`
	Number      int    `json:"number"`
	Name        string `json:"name"`
}

type EncodeRequest struct {
	Node *NodeJSON `json:"node"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    string `json:"data,omitempty"`
}

func (r *DecodeRequest) GetBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(r.Data)
}

func (r *EncodeResponse) SetBytes(data []byte) {
	r.Data = base64.StdEncoding.EncodeToString(data)
}
