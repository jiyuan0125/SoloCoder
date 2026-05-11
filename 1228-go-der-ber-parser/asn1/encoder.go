package asn1

import (
	"errors"
)

type EncodingMode int

const (
	ModeDER EncodingMode = iota
	ModeBER
)

type EncodeOptions struct {
	Mode         EncodingMode
	UseIndefinite bool
}

func DefaultEncodeOptions(mode EncodingMode) *EncodeOptions {
	return &EncodeOptions{
		Mode:         mode,
		UseIndefinite: mode == ModeBER,
	}
}

func Encode(tlv *TLV, opts *EncodeOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultEncodeOptions(ModeDER)
	}

	if opts.Mode == ModeDER && opts.UseIndefinite {
		return nil, errors.New("DER does not support indefinite length")
	}

	var result []byte

	tagBytes := tlv.TagBytes()
	result = append(result, tagBytes...)

	var valueBytes []byte
	if tlv.Constructed {
		for _, child := range tlv.Children {
			childBytes, err := Encode(child, opts)
			if err != nil {
				return nil, err
			}
			valueBytes = append(valueBytes, childBytes...)
		}
	} else {
		valueBytes = tlv.Value
	}

	definite := true
	if opts.Mode == ModeBER && opts.UseIndefinite && tlv.Constructed {
		definite = false
	}

	lengthBytes := EncodeLength(len(valueBytes), definite)
	result = append(result, lengthBytes...)
	result = append(result, valueBytes...)

	if !definite {
		result = append(result, 0x00, 0x00)
	}

	return result, nil
}
