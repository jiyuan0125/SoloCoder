package api

type ValidateRequest struct {
	Plate string `json:"plate"`
}

type ValidateResponse struct {
	Plate  string `json:"plate"`
	Valid  bool   `json:"valid"`
	Type   string `json:"type"`
	Reason string `json:"reason,omitempty"`
}

type BatchValidateRequest struct {
	Plates []string `json:"plates"`
}

type BatchValidateResponse struct {
	Results []ValidateResponse `json:"results"`
}
