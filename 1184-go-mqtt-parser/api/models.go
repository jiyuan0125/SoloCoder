package api

type ParseRequest struct {
	Format string `json:"format"`
	Data   string `json:"data"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Packet  PacketInfo  `json:"packet,omitempty"`
}

type PacketInfo struct {
	Type            string                 `json:"type"`
	Flags           FlagsInfo              `json:"flags"`
	RemainingLength uint32                 `json:"remaining_length"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

type FlagsInfo struct {
	Dup    bool   `json:"dup,omitempty"`
	QoS    int    `json:"qos"`
	Retain bool   `json:"retain,omitempty"`
	Raw    byte   `json:"raw"`
}

type BuildRequest struct {
	PacketType string                 `json:"packet_type"`
	Fields     map[string]interface{} `json:"fields"`
}

type BuildResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Hex     string `json:"hex,omitempty"`
	Base64  string `json:"base64,omitempty"`
	Length  int    `json:"length,omitempty"`
}

type WillConfig struct {
	Enabled bool   `json:"enabled"`
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
	QoS     int    `json:"qos"`
	Retain  bool   `json:"retain"`
}

type WillConfigResponse struct {
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Config  WillConfig `json:"config,omitempty"`
}
