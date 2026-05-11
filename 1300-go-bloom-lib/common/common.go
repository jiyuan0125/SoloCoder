package common

type CreateFilterRequest struct {
	Name              string  `json:"name"`
	ExpectedCapacity   uint64  `json:"expected_capacity"`
	FalsePositiveRate float64 `json:"false_positive_rate"`
}

type CreateFilterResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message,omitempty"`
	Name          string  `json:"name,omitempty"`
	BitArraySize  uint64  `json:"bit_array_size"`
	HashCount      uint64  `json:"hash_count"`
	ExpectedCapacity uint64 `json:"expected_capacity"`
}

type AddElementRequest struct {
	Name    string `json:"name"`
	Element string `json:"element"`
}

type AddElementResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ContainsElementRequest struct {
	Name    string `json:"name"`
	Element string `json:"element"`
}

type ContainsElementResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Contains bool   `json:"contains"`
}

type GetFilterInfoRequest struct {
	Name string `json:"name"`
}

type GetFilterInfoResponse struct {
	Success               bool    `json:"success"`
	Message               string  `json:"message,omitempty"`
	Name                  string  `json:"name,omitempty"`
	ElementsAdded         uint64  `json:"elements_added"`
	FillRate             float64 `json:"fill_rate"`
	CurrentFalsePositive float64 `json:"current_false_positive"`
	RemainingCapacity    uint64  `json:"remaining_capacity"`
	ExpectedCapacity      uint64  `json:"expected_capacity"`
	TargetFalsePositive  float64 `json:"target_false_positive"`
	BitArraySize        uint64  `json:"bit_array_size"`
	HashCount           uint64  `json:"hash_count"`
}

type ClearFilterRequest struct {
	Name string `json:"name"`
}

type ClearFilterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ListFiltersResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message,omitempty"`
	Filters  []string `json:"filters,omitempty"`
}

type ExportFilterRequest struct {
	Name   string `json:"name"`
	Format string `json:"format,omitempty"`
}

type ExportFilterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Data    string `json:"data,omitempty"`
}

type ImportFilterRequest struct {
	Name   string `json:"name"`
	Format string `json:"format,omitempty"`
	Data   string `json:"data"`
}

type ImportFilterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type DeleteFilterRequest struct {
	Name string `json:"name"`
}

type DeleteFilterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
