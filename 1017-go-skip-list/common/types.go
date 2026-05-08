package common

type PutRequest struct {
	Value string `json:"value"`
}

type GetResponse struct {
	Found bool   `json:"found"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type RangeResponse struct {
	Data []*KVPair `json:"data"`
}

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RankResponse struct {
	Found bool `json:"found"`
	Rank  int  `json:"rank,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}
