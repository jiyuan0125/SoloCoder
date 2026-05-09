package api

type EncodeRequest struct {
	InputJSON string `json:"input_json"`
}

type EncodeResponse struct {
	Success      bool   `json:"success"`
	OutputCBOR   string `json:"output_cbor"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type DecodeRequest struct {
	InputCBOR string `json:"input_cbor"`
}

type DecodeResponse struct {
	Success      bool   `json:"success"`
	OutputJSON   string `json:"output_json"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type ConvertJSONToCBORRequest struct {
	InputJSON string `json:"input_json"`
}

type ConvertJSONToCBORResponse struct {
	Success      bool   `json:"success"`
	OutputCBOR   string `json:"output_cbor"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type ConvertCBORToJSONRequest struct {
	InputCBOR string `json:"input_cbor"`
}

type ConvertCBORToJSONResponse struct {
	Success      bool   `json:"success"`
	OutputJSON   string `json:"output_json"`
	ErrorMessage string `json:"error_message,omitempty"`
}
