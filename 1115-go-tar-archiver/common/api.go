package common

const (
	FormatPOSIX = "posix"
	FormatGNU   = "gnu"
)

type ArchiveRequest struct {
	Paths  []string `json:"paths"`
	Format string   `json:"format"`
}

type ExtractRequest struct {
	Format string `json:"format"`
}
