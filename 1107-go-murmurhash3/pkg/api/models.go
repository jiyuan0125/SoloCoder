package api

type HashRequest struct {
	Data string `json:"data"`
	Seed uint32 `json:"seed,omitempty"`
	Type string `json:"type,omitempty"`
}

type Hash32Response struct {
	Hash uint32 `json:"hash"`
	Hex  string `json:"hex"`
	Seed uint32 `json:"seed"`
}

type Hash128Response struct {
	High   uint64 `json:"high"`
	Low    uint64 `json:"low"`
	Hex    string `json:"hex"`
	Seed   uint32 `json:"seed"`
}

type DistributionRequest struct {
	Data        []string `json:"data"`
	Seed        uint32   `json:"seed,omitempty"`
	HashType    string   `json:"hash_type,omitempty"`
	BucketCount int      `json:"bucket_count,omitempty"`
}

type DistributionReport struct {
	Min          uint64  `json:"min"`
	Max          uint64  `json:"max"`
	Mean         float64 `json:"mean"`
	StdDev       float64 `json:"std_dev"`
	Variance     float64 `json:"variance"`
	ChiSquare    float64 `json:"chi_square"`
	UniformScore float64 `json:"uniform_score"`
	BucketCount  int     `json:"bucket_count"`
	SampleCount  int     `json:"sample_count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
