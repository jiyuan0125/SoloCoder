package protocol

type EncodeRequest struct {
	Input string `json:"input"`
}

type EncodeResponse struct {
	Success       bool   `json:"success"`
	Pattern       string `json:"pattern,omitempty"`
	WidthSequence []int  `json:"width_sequence,omitempty"`
	TotalModules  int    `json:"total_modules,omitempty"`
	Input         string `json:"input,omitempty"`
	Error         string `json:"error,omitempty"`
}

func NewSuccessResponse(pattern string, widthSequence []int, totalModules int, input string) *EncodeResponse {
	return &EncodeResponse{
		Success:       true,
		Pattern:       pattern,
		WidthSequence: widthSequence,
		TotalModules:  totalModules,
		Input:         input,
	}
}

func NewErrorResponse(errorMsg string) *EncodeResponse {
	return &EncodeResponse{
		Success: false,
		Error:   errorMsg,
	}
}
