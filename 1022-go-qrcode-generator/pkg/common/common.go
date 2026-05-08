package common

type ErrorLevel string

const (
	ErrorLevelL ErrorLevel = "L"
	ErrorLevelM ErrorLevel = "M"
	ErrorLevelQ ErrorLevel = "Q"
	ErrorLevelH ErrorLevel = "H"
)

const (
	DefaultErrorLevel = ErrorLevelM
	DefaultSize       = 300
	MaxLogoRatio      = 0.2
)

type GenerateRequest struct {
	Content    string     `json:"content" form:"content"`
	ErrorLevel ErrorLevel `json:"error_level" form:"error_level"`
	Size       int        `json:"size" form:"size"`
	Logo       []byte     `json:"-" form:"-"`
}

type GenerateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Image   []byte `json:"-"`
}
