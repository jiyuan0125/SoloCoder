package common

type ReadSheetRequest struct {
	SheetName     string `json:"sheet_name"`
	StartRow      int    `json:"start_row"`
	EndRow        int    `json:"end_row"`
	SkipEmptyRows bool   `json:"skip_empty_rows"`
}

type ReadSheetResponse struct {
	Success    bool       `json:"success"`
	Data       [][]string `json:"data,omitempty"`
	SheetNames []string   `json:"sheet_names,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type ListSheetsResponse struct {
	Success    bool     `json:"success"`
	SheetNames []string `json:"sheet_names,omitempty"`
	Error      string   `json:"error,omitempty"`
}
