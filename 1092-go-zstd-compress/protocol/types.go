package protocol

type CompressionLevel int16

const (
	LevelFastest CompressionLevel = 1
	LevelDefault CompressionLevel = 3
	LevelBest    CompressionLevel = 19
)

var (
	MagicNumber = [8]byte{0x5A, 0x53, 0x54, 0x44, 0x43, 0x4F, 0x4D, 0x50}
)

type FileHeader struct {
	MagicNumber    [8]byte
	OriginalSize   uint32
	CompressedSize uint32
	Level          CompressionLevel
}

type CompressRequest struct {
	FileName    string           `json:"file_name"`
	Dictionary  string           `json:"dictionary,omitempty"`
	Level       CompressionLevel `json:"level"`
}

type CompressResponse struct {
	Success        bool   `json:"success"`
	FileName       string `json:"file_name"`
	OriginalSize   int64  `json:"original_size"`
	CompressedSize int64  `json:"compressed_size"`
	Error          string `json:"error,omitempty"`
}

type DecompressRequest struct {
	FileName string `json:"file_name"`
}

type DecompressResponse struct {
	Success        bool   `json:"success"`
	FileName       string `json:"file_name"`
	OriginalSize   int64  `json:"original_size"`
	CompressedSize int64  `json:"compressed_size"`
	Error          string `json:"error,omitempty"`
}

type TrainDictionaryRequest struct {
	DictionaryName string   `json:"dictionary_name"`
	FilePaths      []string `json:"file_paths"`
}

type TrainDictionaryResponse struct {
	Success      bool   `json:"success"`
	DictName     string `json:"dict_name,omitempty"`
	DictSize     int    `json:"dict_size,omitempty"`
	SampleCount  int    `json:"sample_count,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type DictionaryInfo struct {
	Name        string `json:"name"`
	SampleCount int    `json:"sample_count"`
	DictSize    int    `json:"dict_size"`
	CreatedAt   string `json:"created_at"`
}

type ListDictionaryResponse struct {
	Success       bool             `json:"success"`
	Dictionaries  []DictionaryInfo `json:"dictionaries"`
	ErrorMessage  string           `json:"error_message,omitempty"`
}

type BatchCompressResult struct {
	FileName       string `json:"file_name"`
	OriginalSize   int64  `json:"original_size"`
	CompressedSize int64  `json:"compressed_size"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

type BatchCompressResponse struct {
	Success bool                  `json:"success"`
	Results []BatchCompressResult `json:"results"`
	Error   string                `json:"error,omitempty"`
}
