package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type ParseResponse struct {
	Success   bool                   `json:"success"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	HasThumbnail bool                 `json:"has_thumbnail"`
}

type QueryRequest struct {
	Field string `json:"field"`
}

type QueryResponse struct {
	Success bool        `json:"success"`
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Found   bool        `json:"found"`
}

const (
	FieldDateTime     = "DateTime"
	FieldExposureTime = "ExposureTime"
	FieldFNumber      = "FNumber"
	FieldISO          = "ISO"
	FieldGPSLatitude  = "GPSLatitude"
	FieldGPSLongitude = "GPSLongitude"
	FieldGPSAltitude  = "GPSAltitude"
)
