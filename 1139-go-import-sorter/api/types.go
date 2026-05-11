package api

type FormatRequest struct {
	Source string `json:"source"`
	Config Config `json:"config"`
}

type FormatResponse struct {
	Source string `json:"source"`
	Stats  Stats  `json:"stats"`
}

type Config struct {
	LineWidth        int  `json:"line_width"`
	RemoveEmptyLines bool `json:"remove_empty_lines"`
	ModuleName       string `json:"module_name"`
}

type Stats struct {
	LinesModified int `json:"lines_modified"`
	ImportsMoved  int `json:"imports_moved"`
	LinesSplit    int `json:"lines_split"`
}

type BatchFormatRequest struct {
	Files  []FileEntry `json:"files"`
	Config Config      `json:"config"`
}

type BatchFormatResponse struct {
	Files []FileResult `json:"files"`
}

type FileEntry struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type FileResult struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Stats  Stats  `json:"stats"`
}

type ConfigRequest struct {
	Config Config `json:"config"`
}

type ConfigResponse struct {
	Success bool `json:"success"`
}

func DefaultConfig() Config {
	return Config{
		LineWidth:        120,
		RemoveEmptyLines: true,
		ModuleName:       "",
	}
}
