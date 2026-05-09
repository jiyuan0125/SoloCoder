package common

type SortRequest struct {
	Strings          []string `json:"strings"`
	CaseSensitive    bool     `json:"case_sensitive"`
	Descending       bool     `json:"descending"`
	KeepLeadingZeros bool     `json:"keep_leading_zeros"`
}

type SortResponse struct {
	Sorted []string `json:"sorted"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
