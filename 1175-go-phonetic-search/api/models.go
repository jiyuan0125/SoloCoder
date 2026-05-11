package api

type EncodeRequest struct {
	Name string `json:"name"`
}

type EncodeResponse struct {
	Name      string `json:"name"`
	Soundex   string `json:"soundex"`
	Metaphone string `json:"metaphone"`
	Message   string `json:"message,omitempty"`
}

type SearchRequest struct {
	Name string `json:"name"`
}

type SearchResponse struct {
	QueryName string   `json:"query_name"`
	Soundex   string   `json:"soundex"`
	Metaphone string   `json:"metaphone"`
	Matches   []string `json:"matches"`
	Message   string   `json:"message,omitempty"`
}

type AddRequest struct {
	Name string `json:"name"`
}

type AddResponse struct {
	Name    string `json:"name"`
	Added   bool   `json:"added"`
	Message string `json:"message,omitempty"`
}

type RemoveRequest struct {
	Name string `json:"name"`
}

type RemoveResponse struct {
	Name    string `json:"name"`
	Removed bool   `json:"removed"`
	Message string `json:"message,omitempty"`
}

type ImportRequest struct {
	Names []string `json:"names"`
}

type ImportResponse struct {
	Total    int      `json:"total"`
	Added    int      `json:"added"`
	Existing int      `json:"existing"`
	Invalid  []string `json:"invalid,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type StatsResponse struct {
	TotalNames        int         `json:"total_names"`
	UniqueSoundex     int         `json:"unique_soundex"`
	UniqueMetaphone   int         `json:"unique_metaphone"`
	TopSoundexCodes   []CodeStats `json:"top_soundex_codes"`
	TopMetaphoneCodes []CodeStats `json:"top_metaphone_codes"`
}

type CodeStats struct {
	Code  string   `json:"code"`
	Count int      `json:"count"`
	Names []string `json:"names"`
}
