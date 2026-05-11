package api

type ParseRequest struct {
	VCardText string `json:"vcardText,omitempty"`
}

type ParseResponse struct {
	Success bool      `json:"success"`
	Contacts []Contact `json:"contacts,omitempty"`
	Error   string    `json:"error,omitempty"`
}
