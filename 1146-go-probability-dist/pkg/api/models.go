package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type NormalPDFRequest struct {
	Mu    float64 `json:"mu"`
	Sigma float64 `json:"sigma"`
	X     float64 `json:"x"`
}

type NormalCDFRequest struct {
	Mu    float64 `json:"mu"`
	Sigma float64 `json:"sigma"`
	X     float64 `json:"x"`
}

type NormalSampleRequest struct {
	Mu       float64 `json:"mu"`
	Sigma    float64 `json:"sigma"`
	NSamples int     `json:"n"`
}

type PoissonPMFRequest struct {
	Lambda float64 `json:"lambda"`
	K      int     `json:"k"`
}

type PoissonCDFRequest struct {
	Lambda float64 `json:"lambda"`
	K      int     `json:"k"`
}

type PoissonSampleRequest struct {
	Lambda   float64 `json:"lambda"`
	NSamples int     `json:"n"`
}

type UniformPDFRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
	X float64 `json:"x"`
}

type UniformCDFRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
	X float64 `json:"x"`
}

type UniformSampleRequest struct {
	A        float64 `json:"a"`
	B        float64 `json:"b"`
	NSamples int     `json:"n"`
}

type FloatResponse struct {
	Value float64 `json:"value"`
}

type FloatSampleResponse struct {
	Samples []float64 `json:"samples"`
}

type IntSampleResponse struct {
	Samples []int `json:"samples"`
}
