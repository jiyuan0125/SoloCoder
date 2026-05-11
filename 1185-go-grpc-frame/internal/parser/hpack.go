package parser

import (
	"errors"
	"strings"
)

type HPACKHeaderField struct {
	Name  string
	Value string
}

type HPACKStaticTableEntry struct {
	Name  string
	Value string
}

var hpackStaticTable = []HPACKStaticTableEntry{
	{Name: ":authority", Value: ""},
	{Name: ":method", Value: "GET"},
	{Name: ":method", Value: "POST"},
	{Name: ":path", Value: "/"},
	{Name: ":path", Value: "/index.html"},
	{Name: ":scheme", Value: "http"},
	{Name: ":scheme", Value: "https"},
	{Name: ":status", Value: "200"},
	{Name: ":status", Value: "204"},
	{Name: ":status", Value: "206"},
	{Name: ":status", Value: "304"},
	{Name: ":status", Value: "400"},
	{Name: ":status", Value: "404"},
	{Name: ":status", Value: "500"},
	{Name: "accept-charset", Value: ""},
	{Name: "accept-encoding", Value: "gzip, deflate"},
	{Name: "accept-language", Value: ""},
	{Name: "accept-ranges", Value: ""},
	{Name: "accept", Value: ""},
	{Name: "access-control-allow-origin", Value: ""},
	{Name: "age", Value: ""},
	{Name: "allow", Value: ""},
	{Name: "authorization", Value: ""},
	{Name: "cache-control", Value: ""},
	{Name: "content-disposition", Value: ""},
	{Name: "content-encoding", Value: ""},
	{Name: "content-language", Value: ""},
	{Name: "content-length", Value: ""},
	{Name: "content-location", Value: ""},
	{Name: "content-range", Value: ""},
	{Name: "content-type", Value: ""},
	{Name: "cookie", Value: ""},
	{Name: "date", Value: ""},
	{Name: "etag", Value: ""},
	{Name: "expect", Value: ""},
	{Name: "expires", Value: ""},
	{Name: "from", Value: ""},
	{Name: "host", Value: ""},
	{Name: "if-match", Value: ""},
	{Name: "if-modified-since", Value: ""},
	{Name: "if-none-match", Value: ""},
	{Name: "if-range", Value: ""},
	{Name: "if-unmodified-since", Value: ""},
	{Name: "last-modified", Value: ""},
	{Name: "link", Value: ""},
	{Name: "location", Value: ""},
	{Name: "max-forwards", Value: ""},
	{Name: "proxy-authenticate", Value: ""},
	{Name: "proxy-authorization", Value: ""},
	{Name: "range", Value: ""},
	{Name: "referer", Value: ""},
	{Name: "refresh", Value: ""},
	{Name: "retry-after", Value: ""},
	{Name: "server", Value: ""},
	{Name: "set-cookie", Value: ""},
	{Name: "strict-transport-security", Value: ""},
	{Name: "transfer-encoding", Value: ""},
	{Name: "user-agent", Value: ""},
	{Name: "vary", Value: ""},
	{Name: "via", Value: ""},
	{Name: "www-authenticate", Value: ""},
}

type HPACKDecoder struct {
	dynamicTable []HPACKHeaderField
	maxSize      int
}

func NewHPACKDecoder() *HPACKDecoder {
	return &HPACKDecoder{
		dynamicTable: make([]HPACKHeaderField, 0),
		maxSize:      4096,
	}
}

func (h *HPACKDecoder) Lookup(index int) (*HPACKHeaderField, bool) {
	if index == 0 {
		return nil, false
	}

	if index <= len(hpackStaticTable) {
		return &HPACKHeaderField{
			Name:  hpackStaticTable[index-1].Name,
			Value: hpackStaticTable[index-1].Value,
		}, true
	}

	dynamicIndex := index - len(hpackStaticTable) - 1
	if dynamicIndex >= 0 && dynamicIndex < len(h.dynamicTable) {
		return &h.dynamicTable[dynamicIndex], true
	}

	return nil, false
}

func decodeInteger(data []byte, prefixBits uint) (uint64, int, error) {
	if len(data) == 0 {
		return 0, 0, errors.New("empty data")
	}

	prefixMask := uint8((1 << prefixBits) - 1)
	result := uint64(data[0] & prefixMask)
	offset := 1

	if result < uint64(prefixMask) {
		return result, offset, nil
	}

	var shift uint
	for {
		if offset >= len(data) {
			return 0, offset, errors.New("incomplete integer encoding")
		}

		b := data[offset]
		offset++
		result += uint64(b&0x7F) << shift

		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}

	return result, offset, nil
}

func (h *HPACKDecoder) decodeStringLiteral(data []byte) (string, int, error) {
	if len(data) == 0 {
		return "", 0, errors.New("empty data")
	}

	isHuffman := (data[0] & 0x80) != 0
	length, n, err := decodeInteger(data, 7)
	if err != nil {
		return "", n, err
	}

	if isHuffman {
		return "[Huffman-encoded]", n + int(length), nil
	}

	if len(data) < n+int(length) {
		return "", n, errors.New("insufficient data for string literal")
	}

	return string(data[n : n+int(length)]), n + int(length), nil
}

func (h *HPACKDecoder) Decode(data []byte) ([]HPACKHeaderField, int, error) {
	var headers []HPACKHeaderField
	offset := 0

	for offset < len(data) {
		if offset >= len(data) {
			break
		}

		b := data[offset]

		switch {
		case (b & 0x80) != 0:
			index, n, err := decodeInteger(data[offset:], 7)
			if err != nil {
				return headers, offset, err
			}
			offset += n

			entry, ok := h.Lookup(int(index))
			if !ok {
				return headers, offset, errors.New("invalid index")
			}
			headers = append(headers, *entry)

		case (b & 0x40) != 0:
			index, n, err := decodeInteger(data[offset:], 6)
			offset += n
			if err != nil {
				return headers, offset, err
			}

			var name string
			if index > 0 {
				entry, ok := h.Lookup(int(index))
				if !ok {
					return headers, offset, errors.New("invalid index")
				}
				name = entry.Name
			} else {
				decodedName, n2, err := h.decodeStringLiteral(data[offset:])
				if err != nil {
					return headers, offset, err
				}
				name = decodedName
				offset += n2
			}

			value, n2, err := h.decodeStringLiteral(data[offset:])
			if err != nil {
				return headers, offset, err
			}
			offset += n2

			headers = append(headers, HPACKHeaderField{Name: name, Value: value})

		case (b & 0x20) != 0:
			_, n, err := decodeInteger(data[offset:], 5)
			if err != nil {
				return headers, offset, err
			}
			offset += n

		default:
			index, n, err := decodeInteger(data[offset:], 4)
			offset += n
			if err != nil {
				return headers, offset, err
			}

			var name string
			if index > 0 {
				entry, ok := h.Lookup(int(index))
				if !ok {
					return headers, offset, errors.New("invalid index")
				}
				name = entry.Name
			} else {
				decodedName, n2, err := h.decodeStringLiteral(data[offset:])
				if err != nil {
					return headers, offset, err
				}
				name = decodedName
				offset += n2
			}

			value, n2, err := h.decodeStringLiteral(data[offset:])
			if err != nil {
				return headers, offset, err
			}
			offset += n2

			headers = append(headers, HPACKHeaderField{Name: name, Value: value})
		}
	}

	return headers, offset, nil
}

func DecodeHPACK(data []byte) ([]HPACKHeaderField, error) {
	decoder := NewHPACKDecoder()
	headers, _, err := decoder.Decode(data)
	return headers, err
}

func IsTrailer(headers []HPACKHeaderField) bool {
	if len(headers) == 0 {
		return false
	}

	for _, h := range headers {
		if strings.HasPrefix(h.Name, ":") {
			return false
		}
	}
	return true
}

func GetHeader(headers []HPACKHeaderField, name string) (string, bool) {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value, true
		}
	}
	return "", false
}

func IsGRPCRequest(headers []HPACKHeaderField) bool {
	contentType, ok := GetHeader(headers, "content-type")
	if !ok {
		return false
	}
	return strings.HasPrefix(contentType, "application/grpc")
}

func GetGRPCStatus(headers []HPACKHeaderField) (string, bool) {
	return GetHeader(headers, "grpc-status")
}
