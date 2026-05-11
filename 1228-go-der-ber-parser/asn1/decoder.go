package asn1

import (
	"errors"
	"fmt"
)

type DecodeResult struct {
	TLV        *TLV
	BytesRead  int
	Indefinite bool
}

func Decode(data []byte) (*TLV, int, error) {
	result, err := DecodeOne(data)
	if err != nil {
		return nil, 0, err
	}
	return result.TLV, result.BytesRead, nil
}

func DecodeOne(data []byte) (*DecodeResult, error) {
	tlv, tagLen, err := ParseTag(data)
	if err != nil {
		return nil, err
	}

	length, lenLen, indefinite, err := ParseLength(data[tagLen:])
	if err != nil {
		return nil, err
	}

	totalLen := tagLen + lenLen

	if indefinite {
		if !tlv.Constructed && !IsEndOfContents(tlv) {
			return nil, errors.New("indefinite length only valid for constructed types")
		}

		if IsEndOfContents(tlv) {
			if length != 0 {
				return nil, errors.New("end-of-contents must have length 0")
			}
			return &DecodeResult{
				TLV:        tlv,
				BytesRead:  totalLen,
				Indefinite: false,
			}, nil
		}

		valueStart := totalLen
		remaining := data[valueStart:]
		eocIndex, err := findEndOfContents(remaining)
		if err != nil {
			return nil, err
		}

		valueBytes := remaining[:eocIndex]
		totalLen += eocIndex + 2

		if tlv.Constructed {
			children, err := decodeChildren(valueBytes)
			if err != nil {
				return nil, err
			}
			tlv.Children = children
		} else {
			tlv.Value = valueBytes
		}

		return &DecodeResult{
			TLV:        tlv,
			BytesRead:  totalLen,
			Indefinite: true,
		}, nil
	}

	if len(data) < totalLen+length {
		return nil, errors.New("insufficient data for value")
	}

	valueBytes := data[totalLen : totalLen+length]
	totalLen += length

	if tlv.Constructed {
		children, err := decodeChildren(valueBytes)
		if err != nil {
			return nil, err
		}
		tlv.Children = children
	} else {
		tlv.Value = valueBytes
	}

	return &DecodeResult{
		TLV:        tlv,
		BytesRead:  totalLen,
		Indefinite: false,
	}, nil
}

func findEndOfContents(data []byte) (int, error) {
	i := 0
	for i < len(data) {
		if i+1 < len(data) && data[i] == 0x00 && data[i+1] == 0x00 {
			return i, nil
		}

		tlv, tagLen, err := ParseTag(data[i:])
		if err != nil {
			i++
			continue
		}

		_, lenLen, indefinite, err := ParseLength(data[i+tagLen:])
		if err != nil {
			i++
			continue
		}

		if indefinite {
			if IsEndOfContents(tlv) {
				return i, nil
			}
			eocIdx, err := findEndOfContents(data[i+tagLen+lenLen:])
			if err != nil {
				return 0, err
			}
			i += tagLen + lenLen + eocIdx + 2
		} else {
			length, _, _, _ := ParseLength(data[i+tagLen:])
			i += tagLen + lenLen + length
		}
	}
	return 0, errors.New("end-of-contents marker not found")
}

func decodeChildren(data []byte) ([]*TLV, error) {
	var children []*TLV
	offset := 0
	for offset < len(data) {
		result, err := DecodeOne(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("decode child at offset %d: %w", offset, err)
		}
		children = append(children, result.TLV)
		offset += result.BytesRead
	}
	return children, nil
}

func DecodeAll(data []byte) ([]*TLV, error) {
	var results []*TLV
	offset := 0
	for offset < len(data) {
		result, err := DecodeOne(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("decode at offset %d: %w", offset, err)
		}
		results = append(results, result.TLV)
		offset += result.BytesRead
	}
	return results, nil
}
