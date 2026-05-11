package api

type ConvertRequest struct {
	Content   string `json:"content"`
	Delimiter string `json:"delimiter,omitempty"`
	MaxDigits int    `json:"maxDigits,omitempty"`
	Pretty    bool   `json:"pretty,omitempty"`
}

type ConvertResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

const (
	PathCSVToJSON = "/csv-to-json"
	PathJSONToCSV = "/json-to-csv"
)
