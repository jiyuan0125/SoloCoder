package api

type EncodeRequest struct {
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Pixels [][]byte `json:"pixels"`
}

type EncodeResponse struct {
	Data string `json:"data"`
}

type DecodeRequest struct {
	Data  string `json:"data"`
	Width int    `json:"width"`
}

type DecodeResponse struct {
	Pixels [][]byte `json:"pixels"`
}
