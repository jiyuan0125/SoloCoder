package api

type CompressRequest struct {
	Data     []byte `json:"-"`
	Text     string `json:"text,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

type CompressResponse struct {
	Compressed []byte `json:"-"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Size       int    `json:"size"`
}

type DecompressRequest struct {
	Data     []byte `json:"-"`
	FilePath string `json:"file_path,omitempty"`
}

type DecompressResponse struct {
	Decompressed []byte `json:"-"`
	Text         string `json:"text,omitempty"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	Size         int    `json:"size"`
}
