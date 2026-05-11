package api

type File struct {
	Name    string
	Content string
}

type TargetPlatform struct {
	GOOS   string
	GOARCH string
	Tags   []string
}

type AnalyzeRequest struct {
	Files  []File
	Target TargetPlatform
}

type FileAnalysis struct {
	FileName string
	Included   bool
	Reason     string
}

type GenerateDirective struct {
	FileName string
	Command  string
	Position int
}

type AnalyzeResponse struct {
	IncludedFiles []FileAnalysis
	ExcludedFiles []FileAnalysis
	Warnings []string
	GenerateDirectives []GenerateDirective
}
