package common

type HandshakeRequest struct {
	ClientCipherSuites []string `json:"client_cipher_suites,omitempty"`
	ServerCipherSuites []string `json:"server_cipher_suites,omitempty"`
	SNI                string   `json:"sni,omitempty"`
	ALPNProtocols      []string `json:"alpn_protocols,omitempty"`
	ResumeSessionID    string   `json:"resume_session_id,omitempty"`
	Verbose            bool     `json:"verbose,omitempty"`
}

type HandshakeResponse struct {
	Success          bool             `json:"success"`
	IsSessionResume  bool             `json:"is_session_resume"`
	Error            string           `json:"error,omitempty"`
	NegotiatedSuite  *CipherSuiteInfo `json:"negotiated_suite,omitempty"`
	ClientRandom     string           `json:"client_random,omitempty"`
	ServerRandom     string           `json:"server_random,omitempty"`
	SessionID        string           `json:"session_id,omitempty"`
	ClientExtensions []string         `json:"client_extensions,omitempty"`
	ServerExtensions []string         `json:"server_extensions,omitempty"`
	Steps            []string         `json:"steps"`
	FullMessageDump  string           `json:"full_message_dump,omitempty"`
}

type CipherSuiteInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	KeyExchange string `json:"key_exchange"`
	Encryption  string `json:"encryption"`
	MAC         string `json:"mac"`
}

type CipherSuitesResponse struct {
	CipherSuites []CipherSuiteInfo `json:"cipher_suites"`
}

type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
