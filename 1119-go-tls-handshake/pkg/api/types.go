package api

type HandshakeRequest struct {
	ClientCipherSuites    []string `json:"client_cipher_suites"`
	ServerCipherSuites    []string `json:"server_cipher_suites"`
	ClientSNI              string   `json:"client_sni,omitempty"`
	ClientSupportedGroups  []string `json:"client_supported_groups,omitempty"`
	ClientSignatureAlgs  []string `json:"client_signature_algs,omitempty"`
	ClientALPN           []string `json:"client_alpn,omitempty"`
	SessionID             string   `json:"session_id,omitempty"`
}

type HandshakeResponse struct {
	Success        bool                   `json:"success"`
	ErrorMessage string                   `json:"error_message,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Steps        []HandshakeStep        `json:"steps"`
	NegotiatedCipherSuite string        `json:"negotiated_cipher_suite,omitempty"`
	FinalState   string                 `json:"final_state"`
}

type HandshakeStep struct {
	Sender      string       `json:"sender"`
	MessageType string       `json:"message_type"`
	Content     interface{}  `json:"content"`
	Description string       `json:"description,omitempty"`
}

type CipherSuiteInfo struct {
	ID       uint16 `json:"id"`
	HexID    string `json:"hex_id"`
	Name     string `json:"name"`
}

type ExtensionInfo struct {
	Type    uint16 `json:"type"`
	HexType string `json:"hex_type"`
	Name    string `json:"name"`
}
