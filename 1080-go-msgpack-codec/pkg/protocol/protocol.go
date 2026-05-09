package protocol

const (
	EndpointDecode = "/decode"
	EndpointEncode = "/encode"
)

type ErrorResponse struct {
	Error string `json:"error"`
}
