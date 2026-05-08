package common

type RecordType string

const (
	TypeA     RecordType = "A"
	TypeAAAA  RecordType = "AAAA"
	TypeCNAME RecordType = "CNAME"
	TypeMX    RecordType = "MX"
	TypeTXT   RecordType = "TXT"
)

type DNSRecord struct {
	Type  RecordType `json:"type"`
	Value string     `json:"value"`
	TTL   int        `json:"ttl"`
}

type QueryRequest struct {
	Domain string       `json:"domain"`
	Types  []RecordType `json:"types"`
}

type QueryResult struct {
	Type       RecordType `json:"type"`
	Records    []DNSRecord `json:"records"`
	HitCache   bool       `json:"hit_cache"`
	RemainingTTL int      `json:"remaining_ttl"`
	IsNXDOMAIN bool       `json:"is_nxdomain"`
}

type QueryResponse struct {
	Domain  string        `json:"domain"`
	Results []QueryResult `json:"results"`
}

type RefreshRequest struct {
	Domain string       `json:"domain"`
	Types  []RecordType `json:"types,omitempty"`
}

type RefreshResponse struct {
	Domain  string        `json:"domain"`
	Results []QueryResult `json:"results"`
}

type StatsResponse struct {
	TotalEntries    int     `json:"total_entries"`
	TotalQueries    int64   `json:"total_queries"`
	CacheHits       int64   `json:"cache_hits"`
	CacheMisses     int64   `json:"cache_misses"`
	HitRate         float64 `json:"hit_rate"`
}

type PreloadRequest struct {
	Domains []string `json:"domains"`
}

type PreloadResponse struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

type CacheEntryResponse struct {
	Domain       string       `json:"domain"`
	Type         RecordType   `json:"type"`
	Records      []DNSRecord  `json:"records"`
	ExpiresAt    string       `json:"expires_at"`
	RemainingTTL int          `json:"remaining_ttl"`
	IsNXDOMAIN   bool         `json:"is_nxdomain"`
}

type AllEntriesResponse struct {
	Entries []CacheEntryResponse `json:"entries"`
	Total   int                  `json:"total"`
}
