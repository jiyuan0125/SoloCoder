package common

type ParseRequest struct {
	IDNumber string `json:"id_number"`
}

type ParseResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
	OriginalID     string `json:"original_id,omitempty"`
	StandardizedID string `json:"standardized_id,omitempty"`
	BirthDate      string `json:"birth_date,omitempty"`
	Gender         string `json:"gender,omitempty"`
	ProvinceCode   string `json:"province_code,omitempty"`
	ProvinceName   string `json:"province_name,omitempty"`
	Age            int    `json:"age,omitempty"`
}

type ValidateRequest struct {
	IDNumber string `json:"id_number"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid,omitempty"`
	Message string `json:"message,omitempty"`
}
