package shared

type HeaderInfo struct {
	MagicString        string `json:"magic_string"`
	PageSize           uint16 `json:"page_size"`
	FileFormatWrite    uint8  `json:"file_format_write"`
	FileFormatRead     uint8  `json:"file_format_read"`
	ReservedSpace      uint8  `json:"reserved_space"`
	MaxEmbeddedPayload uint8 `json:"max_embedded_payload"`
	MinEmbeddedPayload uint8  `json:"min_embedded_payload"`
	LeafPayloadFraction uint8 `json:"leaf_payload_fraction"`
	FileChangeCounter  uint32 `json:"file_change_counter"`
	PageCount          uint32 `json:"page_count"`
	FirstFreelistPage  uint32 `json:"first_freelist_page"`
	FreelistPageCount  uint32 `json:"freelist_page_count"`
	SchemaCookie       uint32 `json:"schema_cookie"`
	SchemaFormat       uint32 `json:"schema_format"`
	DefaultPageCache   uint32 `json:"default_page_cache"`
	AutoVacuumTop      uint32 `json:"auto_vacuum_top"`
	IncrementalVacuum  uint32 `json:"incremental_vacuum"`
	TextEncoding       uint32 `json:"text_encoding"`
	UserVersion        uint32 `json:"user_version"`
	ApplicationID      uint32 `json:"application_id"`
	VersionValidFor    uint32 `json:"version_valid_for"`
	SQLiteVersion      uint32 `json:"sqlite_version"`
}

type TableSchema struct {
	Name       string `json:"name"`
	CreateSQL  string `json:"create_sql"`
	RootPage    uint32 `json:"root_page"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TablesResponse struct {
	Tables []TableSchema `json:"tables"`
}

type HeaderResponse struct {
	Header *HeaderInfo `json:"header"`
}

type QueryRequest struct {
	Table  string `json:"table"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type QueryResponse struct {
	Table   string                   `json:"table"`
	Rows    []map[string]interface{} `json:"rows"`
	Count   int                      `json:"count"`
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
