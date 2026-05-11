package frameparser

const (
	FrameTypeData         = 0x00
	FrameTypeHeaders    = 0x01
	FrameTypePriority = 0x02
	FrameTypeRSTStream = 0x03
	FrameTypeSettings = 0x04
	FrameTypePushPromise = 0x05
	FrameTypePing     = 0x06
	FrameTypeGoAway   = 0x07
	FrameTypeWindowUpdate = 0x08
	FrameTypeContinuation = 0x09
)

const (
	FlagEndStream     = 0x01
	FlagEndHeaders    = 0x04
	FlagPadded        = 0x08
	FlagPriority      = 0x20
	FlagAck           = 0x01
)

const (
	DefaultMaxFrameSize    = 16384
	DefaultHeaderTableSize = 4096
	MagicString       = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

const (
	SettingsHeaderTableSize      = 0x01
	SettingsEnablePush       = 0x02
	SettingsMaxConcurrentStreams = 0x03
	SettingsInitialWindowSize  = 0x04
	SettingsMaxFrameSize     = 0x05
	SettingsMaxHeaderListSize = 0x06
)

var FrameTypeNames = map[uint8]string{
	FrameTypeData:         "DATA",
	FrameTypeHeaders:    "HEADERS",
	FrameTypePriority:     "PRIORITY",
	FrameTypeRSTStream: "RST_STREAM",
	FrameTypeSettings:     "SETTINGS",
	FrameTypePushPromise:  "PUSH_PROMISE",
	FrameTypePing:         "PING",
	FrameTypeGoAway:      "GOAWAY",
	FrameTypeWindowUpdate: "WINDOW_UPDATE",
	FrameTypeContinuation: "CONTINUATION",
}

var RSTStreamErrorCodes = map[uint32]string{
	0x00: "NO_ERROR",
	0x01: "PROTOCOL_ERROR",
	0x02: "INTERNAL_ERROR",
	0x03: "FLOW_CONTROL_ERROR",
	0x04: "SETTINGS_TIMEOUT",
	0x05: "STREAM_CLOSED",
	0x06: "FRAME_SIZE_ERROR",
	0x07: "REFUSED_STREAM",
	0x08: "CANCEL",
	0x09: "COMPRESSION_ERROR",
	0x0a: "CONNECT_ERROR",
	0x0b: "ENHANCE_YOUR_CALM",
	0x0c: "INADEQUATE_SECURITY",
	0x0d: "HTTP_1_1_REQUIRED",
}
