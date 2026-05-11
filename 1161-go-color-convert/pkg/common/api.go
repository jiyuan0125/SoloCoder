package common

type ConvertRequest struct {
	From   string   `json:"from"`
	Values []string `json:"values"`
	To     string   `json:"to"`
}

type ConvertResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

type RGBResult struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

type HSLResult struct {
	H float64 `json:"h"`
	S float64 `json:"s"`
	L float64 `json:"l"`
}

type HSVResult struct {
	H float64 `json:"h"`
	S float64 `json:"s"`
	V float64 `json:"v"`
}

type CMYKResult struct {
	C float64 `json:"c"`
	M float64 `json:"m"`
	Y float64 `json:"y"`
	K float64 `json:"k"`
}

type HexResult struct {
	Hex string `json:"hex"`
}

type BrightnessRequest struct {
	Hex string `json:"hex"`
}

type BrightnessResponse struct {
	Success   bool    `json:"success"`
	Error     string  `json:"error,omitempty"`
	Luminance float64 `json:"luminance,omitempty"`
}
