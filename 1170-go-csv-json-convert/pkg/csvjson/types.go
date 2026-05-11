package csvjson

type Delimiter rune

const (
	DelimiterComma   Delimiter = ','
	DelimiterTab     Delimiter = '\t'
	DefaultMaxDigits int       = 10
)

type ConvertOptions struct {
	Delimiter  Delimiter
	MaxDigits  int
	PrettyJSON bool
}

func DefaultOptions() *ConvertOptions {
	return &ConvertOptions{
		Delimiter:  DelimiterComma,
		MaxDigits:  DefaultMaxDigits,
		PrettyJSON: false,
	}
}
