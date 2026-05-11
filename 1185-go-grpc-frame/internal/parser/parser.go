package parser

import (
	"errors"
)

type ParsedProtobufField struct {
	FieldNumber uint64        `json:"fieldNumber"`
	WireType    string        `json:"wireType"`
	Value       interface{}   `json:"value"`
}

type ParsedGRPCFrame struct {
	Compressed bool                 `json:"compressed"`
	Length     uint32               `json:"length"`
	Protobuf   []ParsedProtobufField `json:"protobuf,omitempty"`
	RawMessage []byte               `json:"rawMessage"`
	ParseError string               `json:"parseError,omitempty"`
}

type ParsedHTTP2Headers struct {
	Headers   map[string]string `json:"headers"`
	IsTrailer bool              `json:"isTrailer"`
	IsGRPC    bool              `json:"isGRPC"`
}

type ParsedHTTP2Frame struct {
	Type         string               `json:"type"`
	StreamID     uint32               `json:"streamId"`
	Flags        []string             `json:"flags"`
	Length       uint32               `json:"length"`
	Payload      []byte               `json:"payload"`
	HasEndStream bool                 `json:"hasEndStream"`
	HasEndHeaders bool                `json:"hasEndHeaders"`
	Headers      *ParsedHTTP2Headers  `json:"headers,omitempty"`
	GRPCFrames   []ParsedGRPCFrame    `json:"grpcFrames,omitempty"`
	ParseError   string               `json:"parseError,omitempty"`
}

type ParseResult struct {
	HTTP2Frames []ParsedHTTP2Frame `json:"http2Frames"`
	TotalStreams int               `json:"totalStreams"`
	Errors      []string           `json:"errors,omitempty"`
}

type Parser struct {
	streams map[uint32]*GRPCStreamBuffer
}

func NewParser() *Parser {
	return &Parser{
		streams: make(map[uint32]*GRPCStreamBuffer),
	}
}

func (p *Parser) getStreamBuffer(streamID uint32) *GRPCStreamBuffer {
	if streamID == 0 {
		return NewGRPCStreamBuffer()
	}
	if buf, ok := p.streams[streamID]; ok {
		return buf
	}
	buf := NewGRPCStreamBuffer()
	p.streams[streamID] = buf
	return buf
}

func (p *Parser) ParseHTTP2Frame(frame HTTP2Frame) ParsedHTTP2Frame {
	result := ParsedHTTP2Frame{
		Type:       FrameTypeName(frame.Type),
		StreamID:   frame.StreamID,
		Flags:      FrameTypeFlags(frame.Type, frame.Flags),
		Length:     frame.Length,
		Payload:    frame.Payload,
	}

	switch frame.Type {
	case FrameTypeData:
		result.HasEndStream = (frame.Flags & FlagDataEndStream) != 0

		buf := p.getStreamBuffer(frame.StreamID)
		if err := buf.Write(frame.Payload); err != nil {
			if !errors.Is(err, ErrGRPCInsufficientData) {
				result.ParseError = err.Error()
			}
		}

		grpcFrames := buf.Frames()
		result.GRPCFrames = make([]ParsedGRPCFrame, 0, len(grpcFrames))
		for _, gf := range grpcFrames {
			parsedFrame := ParsedGRPCFrame{
				Compressed: gf.Compressed,
				Length:     gf.Length,
				RawMessage: gf.Message,
			}

			if !gf.Compressed && len(gf.Message) > 0 {
				fields, err := ParseProtobuf(gf.Message)
				if err != nil {
					parsedFrame.ParseError = err.Error()
				} else {
					parsedFrame.Protobuf = make([]ParsedProtobufField, len(fields))
					for i, f := range fields {
						parsedFrame.Protobuf[i] = ParsedProtobufField{
							FieldNumber: f.FieldNumber,
							WireType:    f.WireType.String(),
							Value:       f.Value,
						}
					}
				}
			} else if gf.Compressed {
				parsedFrame.ParseError = "GZIP compression not supported for parsing"
			}

			result.GRPCFrames = append(result.GRPCFrames, parsedFrame)
		}

	case FrameTypeHeaders:
		result.HasEndStream = (frame.Flags & FlagHeadersEndStream) != 0
		result.HasEndHeaders = (frame.Flags & FlagHeadersEndHeaders) != 0

		if len(frame.Payload) > 0 {
			headers, err := DecodeHPACK(frame.Payload)
			if err != nil {
				result.ParseError = err.Error()
			} else {
				headerMap := make(map[string]string)
				for _, h := range headers {
					headerMap[h.Name] = h.Value
				}

				isTrailer := IsTrailer(headers)
				isGRPC := !isTrailer && IsGRPCRequest(headers)

				result.Headers = &ParsedHTTP2Headers{
					Headers:   headerMap,
					IsTrailer: isTrailer,
					IsGRPC:    isGRPC,
				}
			}
		}

	case FrameTypeSettings, FrameTypePing, FrameTypeGoAway, FrameTypeWindowUpdate:
	default:
	}

	return result
}

func (p *Parser) Parse(data []byte) ParseResult {
	result := ParseResult{
		Errors: make([]string, 0),
	}

	frames, err := ParseHTTP2Frames(data)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	result.HTTP2Frames = make([]ParsedHTTP2Frame, 0, len(frames))
	streamSet := make(map[uint32]struct{})

	for _, frame := range frames {
		if frame.StreamID != 0 {
			streamSet[frame.StreamID] = struct{}{}
		}
		parsed := p.ParseHTTP2Frame(frame)
		result.HTTP2Frames = append(result.HTTP2Frames, parsed)
		if parsed.ParseError != "" {
			result.Errors = append(result.Errors, parsed.ParseError)
		}
	}

	result.TotalStreams = len(streamSet)

	return result
}

func Parse(data []byte) ParseResult {
	parser := NewParser()
	return parser.Parse(data)
}

type EncodeRequest struct {
	HTTP2Frames []HTTP2Frame `json:"http2Frames"`
}

type EncodeResult struct {
	Bytes  []byte   `json:"bytes"`
	Errors []string `json:"errors,omitempty"`
}

func Encode(frames []HTTP2Frame) EncodeResult {
	result := EncodeResult{
		Errors: make([]string, 0),
	}

	encoded, err := EncodeHTTP2Frames(frames)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	result.Bytes = encoded
	return result
}

func ValidateFrame(frame HTTP2Frame) error {
	if frame.Length > HTTP2MaxFrameSize {
		return errors.New("frame payload exceeds maximum allowed size")
	}

	if frame.StreamID > HTTP2StreamReserved {
		return errors.New("invalid stream ID")
	}

	if len(frame.Payload) != int(frame.Length) {
		return errors.New("payload length mismatch")
	}

	return nil
}
