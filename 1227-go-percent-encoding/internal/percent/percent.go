package percent

import (
	"bytes"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeURL  Mode = "url"
	ModeURI  Mode = "uri"
	ModeForm Mode = "form"
)

func IsValidMode(m Mode) bool {
	switch m {
	case ModeURL, ModeURI, ModeForm:
		return true
	}
	return false
}

type Component string

const (
	ComponentPath     Component = "path"
	ComponentQuery    Component = "query"
	ComponentFragment Component = "fragment"
	ComponentAll      Component = "all"
)

type EncodeOptions struct {
	Component Component
}

func Encode(s string, mode Mode, opts ...EncodeOptions) string {
	var opt EncodeOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Component == "" {
		opt.Component = ComponentAll
	}

	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEncode(c, mode, opt.Component) {
			buf.WriteByte('%')
			hex := strconv.FormatUint(uint64(c), 16)
			if len(hex) == 1 {
				buf.WriteByte('0')
			}
			buf.WriteString(strings.ToUpper(hex))
		} else if mode == ModeForm && c == ' ' {
			buf.WriteByte('+')
		} else {
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

func Decode(s string, mode Mode) (string, error) {
	if mode == ModeForm {
		s = strings.ReplaceAll(s, "+", " ")
	}
	return decodePercent(s)
}

func shouldEncode(c byte, mode Mode, comp Component) bool {
	if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
		return false
	}

	if c == '-' || c == '.' || c == '_' {
		return false
	}

	if c == '~' {
		if mode == ModeURI {
			return true
		}
		return false
	}

	if mode == ModeForm && c == ' ' {
		return false
	}

	if c == '?' || c == '#' {
		if comp == ComponentQuery || comp == ComponentFragment {
			return false
		}
		return true
	}

	if c == '+' && mode == ModeForm {
		return true
	}

	if c == ' ' {
		return true
	}

	return true
}

func decodePercent(s string) (string, error) {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' {
			if i+2 >= len(s) {
				return "", nil
			}
			hex := s[i+1 : i+3]
			val, err := strconv.ParseUint(hex, 16, 8)
			if err != nil {
				return "", err
			}
			buf.WriteByte(byte(val))
			i += 2
		} else {
			buf.WriteByte(c)
		}
	}
	return buf.String(), nil
}
