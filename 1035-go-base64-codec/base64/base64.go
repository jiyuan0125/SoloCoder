package base64

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const (
	StandardPadding    byte = '='
	StandardAlphabet        = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	URLSafeAlphabet         = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	DefaultMIMELineWidth    = 76
)

type Mode string

const (
	ModeStandard Mode = "standard"
	ModeURLSafe  Mode = "urlsafe"
	ModeMIME     Mode = "mime"
)

type EncodeOptions struct {
	Mode              Mode
	Padding           bool
	MIMELineWidth     int
}

func NewEncodeOptions(mode Mode) *EncodeOptions {
	opts := &EncodeOptions{
		Mode:          mode,
		Padding:       true,
		MIMELineWidth: DefaultMIMELineWidth,
	}
	
	if mode == ModeURLSafe {
		opts.Padding = false
	}
	
	return opts
}

func (opts *EncodeOptions) getAlphabet() string {
	if opts.Mode == ModeURLSafe {
		return URLSafeAlphabet
	}
	return StandardAlphabet
}

type DecodeOptions struct {
	Mode Mode
}

func NewDecodeOptions(mode Mode) *DecodeOptions {
	return &DecodeOptions{Mode: mode}
}

func Encode(data []byte, opts *EncodeOptions) (string, error) {
	if opts == nil {
		opts = NewEncodeOptions(ModeStandard)
	}
	
	alphabet := opts.getAlphabet()
	var result []byte
	n := len(data)
	
	for i := 0; i < n; i += 3 {
		var (
			b0, b1, b2 byte
			chars      int
		)
		
		if i < n {
			b0 = data[i]
			chars++
		}
		if i+1 < n {
			b1 = data[i+1]
			chars++
		}
		if i+2 < n {
			b2 = data[i+2]
			chars++
		}
		
		c0 := b0 >> 2
		result = append(result, alphabet[c0])
		
		c1 := ((b0 & 0x03) << 4) | (b1 >> 4)
		result = append(result, alphabet[c1])
		
		if chars == 1 {
			if opts.Padding {
				result = append(result, StandardPadding, StandardPadding)
			}
			continue
		}
		
		c2 := ((b1 & 0x0F) << 2) | (b2 >> 6)
		result = append(result, alphabet[c2])
		
		if chars == 2 {
			if opts.Padding {
				result = append(result, StandardPadding)
			}
			continue
		}
		
		c3 := b2 & 0x3F
		result = append(result, alphabet[c3])
	}
	
	if opts.Mode == ModeMIME {
		return insertCRLF(result, opts.MIMELineWidth), nil
	}
	
	return string(result), nil
}

func insertCRLF(data []byte, lineWidth int) string {
	if lineWidth <= 0 {
		lineWidth = DefaultMIMELineWidth
	}
	
	var result bytes.Buffer
	for i := 0; i < len(data); i += lineWidth {
		end := i + lineWidth
		if end > len(data) {
			end = len(data)
		}
		result.Write(data[i:end])
		if end < len(data) {
			result.WriteString("\r\n")
		}
	}
	
	return result.String()
}

func removeWhitespace(data string) (string, []int) {
	var builder strings.Builder
	var positions []int
	
	for i, ch := range data {
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			continue
		}
		builder.WriteRune(ch)
		positions = append(positions, i)
	}
	
	return builder.String(), positions
}

func Decode(encoded string, opts *DecodeOptions) ([]byte, error) {
	if opts == nil {
		opts = NewDecodeOptions(ModeStandard)
	}
	
	cleaned, originalPositions := removeWhitespace(encoded)
	n := len(cleaned)
	
	if n == 0 {
		return []byte{}, nil
	}
	
	alphabet := StandardAlphabet
	if opts.Mode == ModeURLSafe {
		alphabet = URLSafeAlphabet
	}
	
	decoderMap := make([]int, 256)
	for i := range decoderMap {
		decoderMap[i] = -1
	}
	for i, ch := range alphabet {
		decoderMap[ch] = i
	}
	
	var result []byte
	var i int
	
	for i < n {
		var quartet [4]int
		count := 0
		paddingCount := 0
		
		for j := 0; j < 4 && i < n; j++ {
			ch := cleaned[i]
			
			if ch == StandardPadding {
				quartet[j] = -1
				paddingCount++
			} else {
				value := decoderMap[ch]
				if value == -1 {
					originalIndex := 0
					if i < len(originalPositions) {
						originalIndex = originalPositions[i]
					}
					return nil, fmt.Errorf("illegal base64 data at input byte %d: char '%c'", originalIndex, ch)
				}
				quartet[j] = value
			}
			count++
			i++
		}
		
		if count < 2 {
			return nil, errors.New("invalid base64: too few characters")
		}
		
		if i == n {
			switch count {
			case 2:
				paddingCount = 2
			case 3:
				paddingCount = 1
			}
		}
		
		b0 := quartet[0]
		b1 := quartet[1]
		b2 := quartet[2]
		b3 := quartet[3]
		
		if b1 == -1 {
			return nil, errors.New("invalid base64: padding at position 2")
		}
		
		byte1 := (b0 << 2) | (b1 >> 4)
		result = append(result, byte(byte1))
		
		if paddingCount <= 1 {
			byte2 := ((b1 & 0x0F) << 4) | (b2 >> 2)
			result = append(result, byte(byte2))
		}
		
		if paddingCount == 0 {
			byte3 := ((b2 & 0x03) << 6) | b3
			result = append(result, byte(byte3))
		}
	}
	
	return result, nil
}

func EncodeStandard(data []byte) (string, error) {
	return Encode(data, NewEncodeOptions(ModeStandard))
}

func EncodeURLSafe(data []byte, padding bool) (string, error) {
	opts := NewEncodeOptions(ModeURLSafe)
	opts.Padding = padding
	return Encode(data, opts)
}

func EncodeMIME(data []byte, lineWidth int) (string, error) {
	opts := NewEncodeOptions(ModeMIME)
	if lineWidth > 0 {
		opts.MIMELineWidth = lineWidth
	}
	return Encode(data, opts)
}

func DecodeStandard(encoded string) ([]byte, error) {
	return Decode(encoded, NewDecodeOptions(ModeStandard))
}

func DecodeURLSafe(encoded string) ([]byte, error) {
	return Decode(encoded, NewDecodeOptions(ModeURLSafe))
}

func DecodeMIME(encoded string) ([]byte, error) {
	return Decode(encoded, NewDecodeOptions(ModeMIME))
}
