package frameparser

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

type FrameHeader struct {
	Length   uint32
	Type     uint8
	Flags    uint8
	StreamID uint32
}

type Frame struct {
	Header  FrameHeader
	Payload []byte
}

type ParserConfig struct {
	MaxFrameSize    uint32
	HeaderTableSize uint32
}

type Parser struct {
	config    ParserConfig
	hpack     *HPACKDecoder
}

type ParsedFrameInternal struct {
	Offset   int
	Type     string
	TypeCode uint8
	Flags    []string
	FlagsCode uint8
	StreamID uint32
	Length   uint32
	Payload  interface{}
	RawHex   string
}

func NewParser() *Parser {
	return &Parser{
		config: ParserConfig{
			MaxFrameSize:    DefaultMaxFrameSize,
			HeaderTableSize: DefaultHeaderTableSize,
		},
		hpack: NewHPACKDecoder(),
	}
}

func (p *Parser) SetMaxFrameSize(size uint32) {
	if size >= 16384 && size <= 16777215 {
		p.config.MaxFrameSize = size
	}
}

func (p *Parser) SetHeaderTableSize(size uint32) {
	p.config.HeaderTableSize = size
}

func (p *Parser) Parse(data []byte) ([]ParsedFrameInternal, error) {
	var frames []ParsedFrameInternal
	offset := 0

	for offset < len(data) {
		if len(data)-offset < 9 {
			return frames, fmt.Errorf("incomplete frame header at offset %d", offset)
		}

		header, err := parseFrameHeader(data[offset:])
		if err != nil {
			return frames, err
		}

		if header.Length > p.config.MaxFrameSize {
			return frames, fmt.Errorf("frame size error at offset %d: length %d exceeds max %d", offset, header.Length, p.config.MaxFrameSize)
		}

		frameEnd := offset + 9 + int(header.Length)
		if frameEnd > len(data) {
			return frames, fmt.Errorf("incomplete frame at offset %d: need %d bytes, have %d", offset, frameEnd, len(data))
		}

		payload := data[offset+9 : frameEnd]
		rawFrame := data[offset:frameEnd]

		parsed, err := p.parsePayload(header, payload)
		if err != nil {
			return frames, fmt.Errorf("parse payload error at offset %d: %w", offset, err)
		}

		flags := parseFlags(header.Type, header.Flags)

		frames = append(frames, ParsedFrameInternal{
			Offset:   offset,
			Type:     FrameTypeNames[header.Type],
			TypeCode: header.Type,
			Flags:    flags,
			FlagsCode: header.Flags,
			StreamID: header.StreamID,
			Length:   header.Length,
			Payload:  parsed,
			RawHex:   hex.EncodeToString(rawFrame),
		})

		offset = frameEnd
	}

	return frames, nil
}

func parseFrameHeader(data []byte) (FrameHeader, error) {
	length := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	typ := data[3]
	flags := data[4]
	streamID := binary.BigEndian.Uint32(data[5:9]) & 0x7FFFFFFF

	return FrameHeader{
		Length:   length,
		Type:     typ,
		Flags:    flags,
		StreamID: streamID,
	}, nil
}

func parseFlags(frameType uint8, flags uint8) []string {
	var result []string

	switch frameType {
	case FrameTypeData, FrameTypeHeaders:
		if flags&FlagEndStream != 0 {
			result = append(result, "END_STREAM")
		}
		if flags&FlagPadded != 0 {
			result = append(result, "PADDED")
		}
		if frameType == FrameTypeHeaders && flags&FlagEndHeaders != 0 {
			result = append(result, "END_HEADERS")
		}
		if frameType == FrameTypeHeaders && flags&FlagPriority != 0 {
			result = append(result, "PRIORITY")
		}
	case FrameTypeSettings, FrameTypePing:
		if flags&FlagAck != 0 {
			result = append(result, "ACK")
		}
	case FrameTypePushPromise, FrameTypeContinuation:
		if flags&FlagEndHeaders != 0 {
			result = append(result, "END_HEADERS")
		}
		if flags&FlagPadded != 0 {
			result = append(result, "PADDED")
		}
	}

	if len(result) == 0 {
		result = append(result, "NONE")
	}

	return result
}

func (p *Parser) parsePayload(header FrameHeader, payload []byte) (interface{}, error) {
	switch header.Type {
	case FrameTypeData:
		return p.parseDataFrame(header, payload)
	case FrameTypeHeaders:
		return p.parseHeadersFrame(header, payload)
	case FrameTypePriority:
		return parsePriorityFrame(payload)
	case FrameTypeRSTStream:
		return parseRSTStreamFrame(payload)
	case FrameTypeSettings:
		return parseSettingsFrame(payload)
	case FrameTypePushPromise:
		return p.parsePushPromiseFrame(header, payload)
	case FrameTypePing:
		return parsePingFrame(payload)
	case FrameTypeGoAway:
		return parseGoAwayFrame(payload)
	case FrameTypeWindowUpdate:
		return parseWindowUpdateFrame(payload)
	case FrameTypeContinuation:
		return p.parseContinuationFrame(payload)
	default:
		return map[string]interface{}{
			"raw_hex": hex.EncodeToString(payload),
		}, nil
	}
}

func parseHexDump(data []byte) string {
	return hex.EncodeToString(data)
}

func (p *Parser) parseDataFrame(header FrameHeader, payload []byte) (interface{}, error) {
	result := make(map[string]interface{})
	offset := 0

	if header.Flags&FlagPadded != 0 {
		if len(payload) < 1 {
			return nil, fmt.Errorf("data frame: missing pad length")
		}
		padLength := int(payload[0])
		result["pad_length"] = padLength
		offset = 1

		if len(payload) < offset+padLength {
			return nil, fmt.Errorf("data frame: padding exceeds payload")
		}

		dataLength := len(payload) - offset - padLength
		if dataLength < 0 {
			return nil, fmt.Errorf("data frame: invalid padding")
		}
		result["data_hex"] = hex.EncodeToString(payload[offset : offset+dataLength])
		result["data_text"] = sanitizeString(payload[offset : offset+dataLength])
	} else {
		result["data_hex"] = hex.EncodeToString(payload)
		result["data_text"] = sanitizeString(payload)
	}

	return result, nil
}

func sanitizeString(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		if b >= 32 && b <= 126 {
			sb.WriteByte(b)
		} else {
			sb.WriteByte('.')
		}
	}
	return sb.String()
}

func (p *Parser) parseHeadersFrame(header FrameHeader, payload []byte) (interface{}, error) {
	result := make(map[string]interface{})
	offset := 0

	if header.Flags&FlagPadded != 0 {
		if len(payload) < 1 {
			return nil, fmt.Errorf("headers frame: missing pad length")
		}
		padLength := int(payload[0])
		result["pad_length"] = padLength
		offset = 1
	}

	if header.Flags&FlagPriority != 0 {
		if len(payload) < offset+5 {
			return nil, fmt.Errorf("headers frame: missing priority info")
		}
		exclusive := payload[offset]&0x80 != 0
		streamDep := binary.BigEndian.Uint32(payload[offset:offset+4]) & 0x7FFFFFFF
		weight := int(payload[offset+4]) + 1
		result["exclusive"] = exclusive
		result["stream_dependency"] = streamDep
		result["weight"] = weight
		offset += 5
	}

	var blockData []byte
	if header.Flags&FlagPadded != 0 {
		padLength := int(result["pad_length"].(uint8) + 1)
		if padLength > 0 {
			if len(payload) < offset+padLength {
				return nil, fmt.Errorf("headers frame: invalid padding")
			}
			blockData = payload[offset : len(payload)-padLength]
		} else {
			blockData = payload[offset:]
		}
	} else {
		blockData = payload[offset:]
	}

	headers, err := p.hpack.Decode(blockData)
	if err != nil {
		result["hpack_error"] = err.Error()
		result["header_block_fragment_hex"] = hex.EncodeToString(blockData)
	} else {
		result["headers"] = headers
	}

	return result, nil
}

func parsePriorityFrame(payload []byte) (interface{}, error) {
	if len(payload) != 5 {
		return nil, fmt.Errorf("priority frame: invalid length")
	}

	exclusive := payload[0]&0x80 != 0
	streamDep := binary.BigEndian.Uint32(payload[0:4]) & 0x7FFFFFFF
	weight := int(payload[4]) + 1

	return map[string]interface{}{
		"exclusive":        exclusive,
		"stream_dependency": streamDep,
		"weight":          weight,
	}, nil
}

func parseRSTStreamFrame(payload []byte) (interface{}, error) {
	if len(payload) != 4 {
		return nil, fmt.Errorf("rst_stream frame: invalid length")
	}

	errorCode := binary.BigEndian.Uint32(payload)
	errorName := RSTStreamErrorCodes[errorCode]
	if errorName == "" {
		errorName = fmt.Sprintf("UNKNOWN(0x%x)", errorCode)
	}

	return map[string]interface{}{
		"error_code": errorCode,
		"error_name": errorName,
	}, nil
}

func parseSettingsFrame(payload []byte) (interface{}, error) {
	if len(payload)%6 != 0 {
		return nil, fmt.Errorf("settings frame: invalid length")
	}

	settings := []map[string]interface{}{}

	for i := 0; i < len(payload); i += 6 {
		id := binary.BigEndian.Uint16(payload[i : i+2])
		value := binary.BigEndian.Uint32(payload[i+2 : i+6])

		setting := map[string]interface{}{
			"id":    id,
			"value": value,
		}

		switch id {
		case SettingsHeaderTableSize:
			setting["name"] = "HEADER_TABLE_SIZE"
		case SettingsEnablePush:
			setting["name"] = "ENABLE_PUSH"
		case SettingsMaxConcurrentStreams:
			setting["name"] = "MAX_CONCURRENT_STREAMS"
		case SettingsInitialWindowSize:
			setting["name"] = "INITIAL_WINDOW_SIZE"
		case SettingsMaxFrameSize:
			setting["name"] = "MAX_FRAME_SIZE"
		case SettingsMaxHeaderListSize:
			setting["name"] = "MAX_HEADER_LIST_SIZE"
		default:
			setting["name"] = fmt.Sprintf("UNKNOWN(0x%x)", id)
		}

		settings = append(settings, setting)
	}

	return map[string]interface{}{
		"settings": settings,
	}, nil
}

func (p *Parser) parsePushPromiseFrame(header FrameHeader, payload []byte) (interface{}, error) {
	result := make(map[string]interface{})
	offset := 0

	if header.Flags&FlagPadded != 0 {
		if len(payload) < 1 {
			return nil, fmt.Errorf("push_promise frame: missing pad length")
		}
		padLength := int(payload[0])
		result["pad_length"] = padLength
		offset = 1
	}

	if len(payload) < offset+4 {
		return nil, fmt.Errorf("push_promise frame: missing promised stream id")
	}

	promisedID := binary.BigEndian.Uint32(payload[offset:offset+4]) & 0x7FFFFFFF
	result["promised_stream_id"] = promisedID
	offset += 4

	var blockData []byte
	if header.Flags&FlagPadded != 0 {
		padLength := int(result["pad_length"].(uint8))
		blockData = payload[offset : len(payload)-padLength]
	} else {
		blockData = payload[offset:]
	}

	headers, err := p.hpack.Decode(blockData)
	if err != nil {
		result["hpack_error"] = err.Error()
		result["header_block_fragment_hex"] = hex.EncodeToString(blockData)
	} else {
		result["headers"] = headers
	}

	return result, nil
}

func parsePingFrame(payload []byte) (interface{}, error) {
	if len(payload) != 8 {
		return nil, fmt.Errorf("ping frame: invalid length")
	}

	return map[string]interface{}{
		"opaque_data_hex": hex.EncodeToString(payload),
	}, nil
}

func parseGoAwayFrame(payload []byte) (interface{}, error) {
	if len(payload) < 8 {
		return nil, fmt.Errorf("goaway frame: invalid length")
	}

	lastStreamID := binary.BigEndian.Uint32(payload[0:4]) & 0x7FFFFFFF
	errorCode := binary.BigEndian.Uint32(payload[4:8])
	errorName := RSTStreamErrorCodes[errorCode]
	if errorName == "" {
		errorName = fmt.Sprintf("UNKNOWN(0x%x)", errorCode)
	}

	result := map[string]interface{}{
		"last_stream_id": lastStreamID,
		"error_code":    errorCode,
		"error_name":    errorName,
	}

	if len(payload) > 8 {
		result["additional_debug_data_hex"] = hex.EncodeToString(payload[8:])
		result["additional_debug_data_text"] = sanitizeString(payload[8:])
	}

	return result, nil
}

func parseWindowUpdateFrame(payload []byte) (interface{}, error) {
	if len(payload) != 4 {
		return nil, fmt.Errorf("window_update frame: invalid length")
	}

	increment := binary.BigEndian.Uint32(payload) & 0x7FFFFFFF

	if increment == 0 {
		return nil, fmt.Errorf("window_update frame: window size increment cannot be 0")
	}

	return map[string]interface{}{
		"window_size_increment": increment,
	}, nil
}

func (p *Parser) parseContinuationFrame(payload []byte) (interface{}, error) {
	headers, err := p.hpack.Decode(payload)
	result := make(map[string]interface{})

	if err != nil {
		result["hpack_error"] = err.Error()
		result["header_block_fragment_hex"] = hex.EncodeToString(payload)
	} else {
		result["headers"] = headers
	}

	return result, nil
}
