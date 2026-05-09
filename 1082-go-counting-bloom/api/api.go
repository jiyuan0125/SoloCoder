package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type AddRequest struct {
	Item string `json:"item"`
}

type AddResponse struct {
	Success bool `json:"success"`
}

type DeleteRequest struct {
	Item string `json:"item"`
}

type DeleteResponse struct {
	Success bool `json:"success"`
}

type QueryRequest struct {
	Item string `json:"item"`
}

type QueryResponse struct {
	MayContain bool `json:"may_contain"`
}

type InfoResponse struct {
	NumHashes       uint64 `json:"num_hashes"`
	CounterBits     uint8  `json:"counter_bits"`
	Capacity        uint64 `json:"capacity"`
	Inserted       uint64 `json:"inserted"`
	NonZeroCounters uint64 `json:"non_zero_counters"`
	OverflowCount  uint64 `json:"overflow_count"`
}
