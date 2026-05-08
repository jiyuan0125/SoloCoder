package common

import "encoding/json"

type ProtoToJSONRequest struct {
	ProtoName string          `json:"proto_name"`
	ProtoData json.RawMessage `json:"proto_data"`
}

type ProtoToJSONResponse struct {
	JSONData json.RawMessage `json:"json_data"`
	Error    string          `json:"error,omitempty"`
}

type JSONToProtoRequest struct {
	ProtoName string          `json:"proto_name"`
	JSONData  json.RawMessage `json:"json_data"`
}

type JSONToProtoResponse struct {
	ProtoData json.RawMessage `json:"proto_data"`
	Error     string          `json:"error,omitempty"`
}
