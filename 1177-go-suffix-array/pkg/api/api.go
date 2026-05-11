package api

type BuildRequest struct {
	Text string `json:"text"`
}

type BuildResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Text    string `json:"text"`
	SA      []int  `json:"sa"`
	Rank    []int  `json:"rank"`
	LCP     []int  `json:"lcp"`
}

type LRSRequest struct {
	Text string `json:"text"`
}

type LRSResponse struct {
	Success    bool                `json:"success"`
	Error      string              `json:"error,omitempty"`
	Text       string              `json:"text"`
	MaxLength  int                 `json:"max_length"`
	Substrings []RepeatedSubstring `json:"substrings"`
}

type RepeatedSubstring struct {
	Substring string `json:"substring"`
	Positions []int  `json:"positions"`
}

type CountRequest struct {
	Text    string `json:"text"`
	Pattern string `json:"pattern"`
}

type CountResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Text      string `json:"text"`
	Pattern   string `json:"pattern"`
	Count     int    `json:"count"`
	Positions []int  `json:"positions"`
}

type DemoResponse struct {
	Success            bool                `json:"success"`
	Error              string              `json:"error,omitempty"`
	Text               string              `json:"text"`
	SA                 []int               `json:"sa"`
	LCP                []int               `json:"lcp"`
	LongestRepeated    []RepeatedSubstring `json:"longest_repeated"`
	ExampleQueries     []QueryResult       `json:"example_queries"`
}

type QueryResult struct {
	Pattern   string `json:"pattern"`
	Count     int    `json:"count"`
	Positions []int  `json:"positions"`
}
