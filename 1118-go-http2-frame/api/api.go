package api

type ParseRequest struct {
	Data     string `json:"data"`
	DataKind string `json:"data_kind"`
	Settings *FrameSettings `json:"settings,omitempty"`
}

type FrameSettings struct {
	MaxFrameSize       uint32 `json:"max_frame_size,omitempty"`
	HeaderTableSize    uint32 `json:"header_table_size,omitempty"`
}

type ParseResponse struct {
	Frames []ParsedFrame `json:"frames"`
	Error  string        `json:"error,omitempty"`
}

type ParsedFrame struct {
	Offset       int         `json:"offset"`
	Type         string      `json:"type"`
	TypeCode     uint8       `json:"type_code"`
	Flags        []string    `json:"flags"`
	FlagsCode    uint8       `json:"flags_code"`
	StreamID     uint32      `json:"stream_id"`
	Length       uint32      `json:"length"`
	Payload      interface{} `json:"payload"`
	RawHex       string      `json:"raw_hex"`
}
