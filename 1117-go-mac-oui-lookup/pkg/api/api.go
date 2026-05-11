package api

type ParseRequest struct {
	MAC string `json:"mac"`
}

type ParseResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	MAC     MACInfo `json:"mac,omitempty"`
}

type MACInfo struct {
	Address          string   `json:"address"`
	OUI              string   `json:"oui"`
	Vendor           string   `json:"vendor,omitempty"`
	Types            []string `json:"types"`
	IsMulticast      bool     `json:"is_multicast"`
	IsLocallyAdmin   bool     `json:"is_locally_admin"`
	IsGloballyUnique bool     `json:"is_globally_unique"`
	IsZero           bool     `json:"is_zero"`
	IsBroadcast      bool     `json:"is_broadcast"`
}

type FormatRequest struct {
	MAC    string `json:"mac"`
	Format string `json:"format"`
}

type FormatResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Result  string `json:"result,omitempty"`
}

type OUIRequest struct {
	MAC string `json:"mac"`
}

type OUIResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	OUI     string `json:"oui,omitempty"`
	Vendor  string `json:"vendor,omitempty"`
}

type TypesRequest struct {
	MAC string `json:"mac"`
}

type TypesResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Types   []string `json:"types,omitempty"`
}
