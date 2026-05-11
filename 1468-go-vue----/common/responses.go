package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(message string) Response {
	return Response{
		Success: false,
		Message: message,
	}
}

type ProjectListData struct {
	Projects []Project `json:"projects"`
}

type ProjectDetailData struct {
	Project Project `json:"project"`
}

type BidDetailData struct {
	Bid Bid `json:"bid"`
}

type OpeningResultData struct {
	Result OpeningResult `json:"result"`
}

type IDData struct {
	ID string `json:"id"`
}
