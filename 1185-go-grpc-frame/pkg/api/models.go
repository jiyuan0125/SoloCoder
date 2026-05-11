package api

type ParseRequest struct {
	Bytes []byte `json:"bytes"`
	Hex   string `json:"hex,omitempty"`
}

type ParseResponse struct {
	Success    bool              `json:"success"`
	HTTP2Frames []HTTP2FrameInfo  `json:"http2Frames,omitempty"`
	TotalStreams int              `json:"totalStreams"`
	Errors     []string          `json:"errors,omitempty"`
}

type HTTP2FrameInfo struct {
	Type          string           `json:"type"`
	StreamID      uint32           `json:"streamId"`
	Flags         []string         `json:"flags"`
	Length        uint32           `json:"length"`
	HasEndStream  bool             `json:"hasEndStream"`
	HasEndHeaders bool             `json:"hasEndHeaders"`
	Headers       *HeadersInfo     `json:"headers,omitempty"`
	GRPCFrames    []GRPCFrameInfo  `json:"grpcFrames,omitempty"`
	ParseError    string           `json:"parseError,omitempty"`
}

type HeadersInfo struct {
	Headers   map[string]string `json:"headers"`
	IsTrailer bool              `json:"isTrailer"`
	IsGRPC    bool              `json:"isGRPC"`
}

type GRPCFrameInfo struct {
	Compressed bool                 `json:"compressed"`
	Length     uint32               `json:"length"`
	Protobuf   []ProtobufFieldInfo  `json:"protobuf,omitempty"`
	RawMessage string               `json:"rawMessage"`
	ParseError string               `json:"parseError,omitempty"`
}

type ProtobufFieldInfo struct {
	FieldNumber uint64        `json:"fieldNumber"`
	WireType    string        `json:"wireType"`
	Value       interface{}   `json:"value"`
}

type EncodeRequest struct {
	HTTP2Frames []HTTP2FrameEncode `json:"http2Frames"`
}

type HTTP2FrameEncode struct {
	Type        uint8  `json:"type"`
	StreamID    uint32 `json:"streamId"`
	Flags       uint8  `json:"flags"`
	Payload     []byte `json:"payload,omitempty"`
	PayloadHex  string `json:"payloadHex,omitempty"`
}

type EncodeResponse struct {
	Success bool     `json:"success"`
	Bytes   []byte   `json:"bytes"`
	Hex     string   `json:"hex"`
	Errors  []string `json:"errors,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
