package protocol

type FramerMode string

const (
	ModeLengthPrefix FramerMode = "length_prefix"
	ModeDelimiter    FramerMode = "delimiter"
)

type ByteOrder string

const (
	ByteOrderBigEndian    ByteOrder = "big"
	ByteOrderLittleEndian ByteOrder = "little"
)

type ConfigRequest struct {
	Mode         FramerMode `json:"mode"`
	HeaderSize   int        `json:"header_size"`
	ByteOrder    ByteOrder  `json:"byte_order"`
	MaxFrameSize int        `json:"max_frame_size"`
	Delimiter    string     `json:"delimiter"`
}

type ConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetConfigResponse struct {
	Mode         FramerMode `json:"mode"`
	HeaderSize   int        `json:"header_size"`
	ByteOrder    ByteOrder  `json:"byte_order"`
	MaxFrameSize int        `json:"max_frame_size"`
	Delimiter    string     `json:"delimiter"`
}

type EchoRequest struct {
	Messages [][]byte `json:"messages"`
	Mode     FramerMode `json:"mode"`
}

type EchoResponse struct {
	Messages [][]byte `json:"messages"`
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
}
