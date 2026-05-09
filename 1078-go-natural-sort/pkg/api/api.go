package api

type SortRequest struct {
	Strings            []string `json:"strings"`
	Ascending          bool     `json:"ascending"`
	CaseSensitive      bool     `json:"case_sensitive"`
	IgnoreLeadingZeros bool     `json:"ignore_leading_zeros"`
}

type SortResponse struct {
	Strings []string `json:"strings"`
	Error   string   `json:"error,omitempty"`
}

func DefaultSortRequest() SortRequest {
	return SortRequest{
		Strings:            []string{},
		Ascending:          true,
		CaseSensitive:      false,
		IgnoreLeadingZeros: true,
	}
}
