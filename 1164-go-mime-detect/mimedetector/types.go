package mimedetector

type DetectionResult struct {
	ExtensionMIME string
	MagicMIME     string
	FinalMIME     string
	Warning       string
	Confidence    string
}

type FileType struct {
	MIMEType       string
	Extensions     []string
	MagicNumbers   [][]byte
	MagicOffset    int
	ContainerCheck bool
}
