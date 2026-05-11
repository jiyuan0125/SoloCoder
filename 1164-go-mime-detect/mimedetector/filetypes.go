package mimedetector

var SupportedFileTypes = []FileType{
	{
		MIMEType:     "image/jpeg",
		Extensions:   []string{".jpg", ".jpeg"},
		MagicNumbers: [][]byte{{0xFF, 0xD8}},
	},
	{
		MIMEType:     "image/png",
		Extensions:   []string{".png"},
		MagicNumbers: [][]byte{{0x89, 0x50, 0x4E, 0x47}},
	},
	{
		MIMEType:     "image/gif",
		Extensions:   []string{".gif"},
		MagicNumbers: [][]byte{{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, {0x47, 0x49, 0x46, 0x38, 0x39, 0x61}},
	},
	{
		MIMEType:     "image/webp",
		Extensions:   []string{".webp"},
		MagicNumbers: [][]byte{{0x52, 0x49, 0x46, 0x46}},
	},
	{
		MIMEType:     "application/pdf",
		Extensions:   []string{".pdf"},
		MagicNumbers: [][]byte{{0x25, 0x50, 0x44, 0x46}},
	},
	{
		MIMEType:       "application/zip",
		Extensions:     []string{".zip"},
		MagicNumbers:   [][]byte{{0x50, 0x4B, 0x03, 0x04}, {0x50, 0x4B, 0x05, 0x06}, {0x50, 0x4B, 0x07, 0x08}},
		ContainerCheck: true,
	},
	{
		MIMEType:     "application/vnd.rar",
		Extensions:   []string{".rar"},
		MagicNumbers: [][]byte{{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}, {0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x01, 0x00}},
	},
	{
		MIMEType:     "application/gzip",
		Extensions:   []string{".gz"},
		MagicNumbers: [][]byte{{0x1F, 0x8B}},
	},
	{
		MIMEType:     "application/x-bzip2",
		Extensions:   []string{".bz2"},
		MagicNumbers: [][]byte{{0x42, 0x5A, 0x68}},
	},
	{
		MIMEType:     "application/x-7z-compressed",
		Extensions:   []string{".7z"},
		MagicNumbers: [][]byte{{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}},
	},
	{
		MIMEType:     "audio/mpeg",
		Extensions:   []string{".mp3"},
		MagicNumbers: [][]byte{{0xFF, 0xFB}, {0xFF, 0xF3}, {0xFF, 0xF2}, {0x49, 0x44, 0x33}},
	},
	{
		MIMEType:     "video/mp4",
		Extensions:   []string{".mp4"},
		MagicNumbers: [][]byte{},
	},
	{
		MIMEType:     "video/x-msvideo",
		Extensions:   []string{".avi"},
		MagicNumbers: [][]byte{{0x52, 0x49, 0x46, 0x46}},
	},
	{
		MIMEType:     "application/x-dosexec",
		Extensions:   []string{".exe", ".dll"},
		MagicNumbers: [][]byte{{0x4D, 0x5A}},
	},
	{
		MIMEType:     "application/x-executable",
		Extensions:   []string{},
		MagicNumbers: [][]byte{{0x7F, 0x45, 0x4C, 0x46}},
	},
}
