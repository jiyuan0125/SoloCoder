package api

type MacParseRequest struct {
	Mac string `json:"mac"`
}

type MacFormatRequest struct {
	Mac    string `json:"mac"`
	Format string `json:"format"`
}

type MacOUIRequest struct {
	Mac string `json:"mac"`
}

type MacTypeRequest struct {
	Mac string `json:"mac"`
}

type MacParseResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	MacHex    string `json:"mac_hex,omitempty"`
	MacBytes  []byte `json:"mac_bytes,omitempty"`
}

type MacFormatResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Formatted string `json:"formatted,omitempty"`
}

type MacOUIResponse struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	OUI        string `json:"oui,omitempty"`
	Vendor     string `json:"vendor,omitempty"`
	Found      bool   `json:"found,omitempty"`
}

type MacTypeResponse struct {
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
	IsUnicast        bool   `json:"is_unicast,omitempty"`
	IsMulticast      bool   `json:"is_multicast,omitempty"`
	IsBroadcast      bool   `json:"is_broadcast,omitempty"`
	IsLocal          bool   `json:"is_local,omitempty"`
	IsGlobal         bool   `json:"is_global,omitempty"`
	IsUninitialized  bool   `json:"is_uninitialized,omitempty"`
	AddressType      string `json:"address_type,omitempty"`
}

type MacInfoResponse struct {
	Success   bool           `json:"success"`
	Error     string         `json:"error,omitempty"`
	Mac       string         `json:"mac,omitempty"`
	MacHex    string         `json:"mac_hex,omitempty"`
	MacBytes  []byte         `json:"mac_bytes,omitempty"`
	OUI       string         `json:"oui,omitempty"`
	Vendor    string         `json:"vendor,omitempty"`
	OUIFound  bool           `json:"oui_found,omitempty"`
	Types     MacTypeResponse `json:"types,omitempty"`
}
