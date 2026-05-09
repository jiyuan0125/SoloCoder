package http2frame

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	FrameTypeData         = 0x00
	FrameTypeHeaders      = 0x01
	FrameTypePriority     = 0x02
	FrameTypeRSTStream    = 0x03
	FrameTypeSettings     = 0x04
	FrameTypePushPromise  = 0x05
	FrameTypePing         = 0x06
	FrameTypeGoAway       = 0x07
	FrameTypeWindowUpdate = 0x08
	FrameTypeContinuation = 0x09

	FlagEndStream  = 0x01
	FlagEndHeaders = 0x04
	FlagPadded     = 0x08
	FlagPriority   = 0x20
	FlagAck        = 0x01

	DefaultMaxFrameSize = 16384
	MaxMaxFrameSize     = 16777215
)

var FrameTypeNames = map[uint8]string{
	FrameTypeData:         "DATA",
	FrameTypeHeaders:      "HEADERS",
	FrameTypePriority:     "PRIORITY",
	FrameTypeRSTStream:    "RST_STREAM",
	FrameTypeSettings:     "SETTINGS",
	FrameTypePushPromise:  "PUSH_PROMISE",
	FrameTypePing:         "PING",
	FrameTypeGoAway:       "GOAWAY",
	FrameTypeWindowUpdate: "WINDOW_UPDATE",
	FrameTypeContinuation: "CONTINUATION",
}

var ErrorCodeNames = map[uint32]string{
	0x0: "NO_ERROR",
	0x1: "PROTOCOL_ERROR",
	0x2: "INTERNAL_ERROR",
	0x3: "FLOW_CONTROL_ERROR",
	0x4: "SETTINGS_TIMEOUT",
	0x5: "STREAM_CLOSED",
	0x6: "FRAME_SIZE_ERROR",
	0x7: "REFUSED_STREAM",
	0x8: "CANCEL",
	0x9: "COMPRESSION_ERROR",
	0xa: "CONNECT_ERROR",
	0xb: "ENHANCE_YOUR_CALM",
	0xc: "INADEQUATE_SECURITY",
	0xd: "HTTP_1_1_REQUIRED",
}

var SettingIDNames = map[uint16]string{
	0x1: "SETTINGS_HEADER_TABLE_SIZE",
	0x2: "SETTINGS_ENABLE_PUSH",
	0x3: "SETTINGS_MAX_CONCURRENT_STREAMS",
	0x4: "SETTINGS_INITIAL_WINDOW_SIZE",
	0x5: "SETTINGS_MAX_FRAME_SIZE",
	0x6: "SETTINGS_MAX_HEADER_LIST_SIZE",
}

type Frame struct {
	Length   uint32
	Type     uint8
	Flags    uint8
	StreamID uint32
	Payload  Payload
}

type Payload interface {
	isPayload()
	String() string
}

type DataPayload struct {
	PaddingLength uint8
	Data          []byte
	Padding       []byte
}

func (d *DataPayload) isPayload() {}

func (d *DataPayload) String() string {
	return fmt.Sprintf("Data{len=%d, padding=%d}", len(d.Data), d.PaddingLength)
}

type HeadersPayload struct {
	PaddingLength    uint8
	StreamDependency uint32
	Weight           uint8
	Headers          []HeaderField
	HeaderBlock      []byte
}

func (h *HeadersPayload) isPayload() {}

func (h *HeadersPayload) String() string {
	return fmt.Sprintf("Headers{headers=%d, streamDep=%d, weight=%d}", len(h.Headers), h.StreamDependency, h.Weight)
}

type RSTStreamPayload struct {
	ErrorCode uint32
}

func (r *RSTStreamPayload) isPayload() {}

func (r *RSTStreamPayload) String() string {
	name, ok := ErrorCodeNames[r.ErrorCode]
	if !ok {
		name = "UNKNOWN"
	}
	return fmt.Sprintf("RST_STREAM{errorCode=%s(0x%x)}", name, r.ErrorCode)
}

type SettingsPayload struct {
	Settings []Setting
}

func (s *SettingsPayload) isPayload() {}

func (s *SettingsPayload) String() string {
	return fmt.Sprintf("SETTINGS{count=%d}", len(s.Settings))
}

type Setting struct {
	ID    uint16
	Value uint32
}

func (s *Setting) String() string {
	name, ok := SettingIDNames[s.ID]
	if !ok {
		name = "UNKNOWN"
	}
	return fmt.Sprintf("%s=0x%x(%d)", name, s.Value, s.Value)
}

type WindowUpdatePayload struct {
	WindowSizeIncrement uint32
}

func (w *WindowUpdatePayload) isPayload() {}

func (w *WindowUpdatePayload) String() string {
	return fmt.Sprintf("WINDOW_UPDATE{increment=%d}", w.WindowSizeIncrement)
}

type PriorityPayload struct {
	StreamDependency uint32
	Weight           uint8
}

func (p *PriorityPayload) isPayload() {}

func (p *PriorityPayload) String() string {
	return fmt.Sprintf("PRIORITY{streamDep=%d, weight=%d}", p.StreamDependency, p.Weight)
}

type PingPayload struct {
	OpaqueData [8]byte
}

func (p *PingPayload) isPayload() {}

func (p *PingPayload) String() string {
	return fmt.Sprintf("PING{opaque=%x}", p.OpaqueData)
}

type GoAwayPayload struct {
	LastStreamID uint32
	ErrorCode    uint32
	DebugData    []byte
}

func (g *GoAwayPayload) isPayload() {}

func (g *GoAwayPayload) String() string {
	name, ok := ErrorCodeNames[g.ErrorCode]
	if !ok {
		name = "UNKNOWN"
	}
	return fmt.Sprintf("GOAWAY{lastStream=%d, errorCode=%s(0x%x), debugLen=%d}", g.LastStreamID, name, g.ErrorCode, len(g.DebugData))
}

type PushPromisePayload struct {
	PaddingLength    uint8
	PromisedStreamID uint32
	HeaderBlock      []byte
	Headers          []HeaderField
}

func (p *PushPromisePayload) isPayload() {}

func (p *PushPromisePayload) String() string {
	return fmt.Sprintf("PUSH_PROMISE{promisedStream=%d, headers=%d}", p.PromisedStreamID, len(p.Headers))
}

type ContinuationPayload struct {
	HeaderBlock []byte
	Headers     []HeaderField
}

func (c *ContinuationPayload) isPayload() {}

func (c *ContinuationPayload) String() string {
	return fmt.Sprintf("CONTINUATION{headers=%d}", len(c.Headers))
}

type UnknownPayload struct {
	Data []byte
}

func (u *UnknownPayload) isPayload() {}

func (u *UnknownPayload) String() string {
	return fmt.Sprintf("UNKNOWN{len=%d}", len(u.Data))
}

type HeaderField struct {
	Name  string
	Value string
}

func (h *HeaderField) String() string {
	return fmt.Sprintf("%s: %s", h.Name, h.Value)
}

func (f *Frame) String() string {
	typeName, ok := FrameTypeNames[f.Type]
	if !ok {
		typeName = fmt.Sprintf("UNKNOWN(0x%x)", f.Type)
	}
	return fmt.Sprintf("Frame{type=%s, flags=0x%x, stream=%d, length=%d, payload=%s}",
		typeName, f.Flags, f.StreamID, f.Length, f.Payload.String())
}

var ErrInvalidFrameSize = errors.New("frame size exceeds maximum")
var ErrInvalidWindowIncrement = errors.New("window size increment cannot be zero")

func ReadFrameHeader(r io.Reader, maxFrameSize uint32) (uint32, uint8, uint8, uint32, error) {
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, 0, 0, 0, err
	}

	length := uint32(header[0])<<16 | uint32(header[1])<<8 | uint32(header[2])
	typ := header[3]
	flags := header[4]
	streamID := binary.BigEndian.Uint32(header[5:9]) & 0x7fffffff

	if length > maxFrameSize {
		return 0, 0, 0, 0, ErrInvalidFrameSize
	}

	return length, typ, flags, streamID, nil
}

func ReadFramePayload(r io.Reader, length uint32) ([]byte, error) {
	payload := make([]byte, length)
	if length == 0 {
		return payload, nil
	}
	_, err := io.ReadFull(r, payload)
	return payload, err
}
