package wsframe

import (
	"crypto/rand"
	"encoding/binary"
)

type FrameBuilder struct {
	frame *Frame
}

func NewFrameBuilder() *FrameBuilder {
	return &FrameBuilder{
		frame: &Frame{
			Fin: true,
		},
	}
}

func (b *FrameBuilder) SetFin(fin bool) *FrameBuilder {
	b.frame.Fin = fin
	return b
}

func (b *FrameBuilder) SetOpcode(opcode int) *FrameBuilder {
	b.frame.OpCode = opcode
	return b
}

func (b *FrameBuilder) SetMasked(masked bool) *FrameBuilder {
	b.frame.Masked = masked
	if masked {
		rand.Read(b.frame.MaskKey[:])
	}
	return b
}

func (b *FrameBuilder) SetPayload(payload []byte) *FrameBuilder {
	b.frame.Payload = make([]byte, len(payload))
	copy(b.frame.Payload, payload)
	return b
}

func (b *FrameBuilder) Build() *Frame {
	return b.frame
}

func BuildTextFrame(text string, masked bool) *Frame {
	return NewFrameBuilder().
		SetFin(true).
		SetOpcode(OpCodeText).
		SetMasked(masked).
		SetPayload([]byte(text)).
		Build()
}

func BuildBinaryFrame(data []byte, masked bool) *Frame {
	return NewFrameBuilder().
		SetFin(true).
		SetOpcode(OpCodeBinary).
		SetMasked(masked).
		SetPayload(data).
		Build()
}

func BuildPingFrame(payload []byte, masked bool) *Frame {
	return NewFrameBuilder().
		SetFin(true).
		SetOpcode(OpCodePing).
		SetMasked(masked).
		SetPayload(payload).
		Build()
}

func BuildPongFrame(payload []byte, masked bool) *Frame {
	return NewFrameBuilder().
		SetFin(true).
		SetOpcode(OpCodePong).
		SetMasked(masked).
		SetPayload(payload).
		Build()
}

func BuildCloseFrame(code int, reason string, masked bool) *Frame {
	var payload []byte
	if code > 0 {
		payload = make([]byte, 2)
		binary.BigEndian.PutUint16(payload, uint16(code))
		if reason != "" {
			payload = append(payload, []byte(reason)...)
		}
	}
	return NewFrameBuilder().
		SetFin(true).
		SetOpcode(OpCodeClose).
		SetMasked(masked).
		SetPayload(payload).
		Build()
}

func SerializeFrame(frame *Frame) []byte {
	var header []byte

	firstByte := byte(frame.OpCode)
	if frame.Fin {
		firstByte |= 0x80
	}

	secondByte := byte(0)
	if frame.Masked {
		secondByte |= 0x80
	}

	payloadLen := len(frame.Payload)
	var lengthBytes []byte

	if payloadLen <= 125 {
		secondByte |= byte(payloadLen)
	} else if payloadLen <= 65535 {
		secondByte |= 126
		lengthBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(lengthBytes, uint16(payloadLen))
	} else {
		secondByte |= 127
		lengthBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(lengthBytes, uint64(payloadLen))
	}

	header = append(header, firstByte, secondByte)
	header = append(header, lengthBytes...)

	if frame.Masked {
		header = append(header, frame.MaskKey[:]...)
	}

	result := make([]byte, len(header)+payloadLen)
	copy(result, header)

	if payloadLen > 0 {
		if frame.Masked {
			maskedPayload := make([]byte, payloadLen)
			copy(maskedPayload, frame.Payload)
			for i := range maskedPayload {
				maskedPayload[i] ^= frame.MaskKey[i%4]
			}
			copy(result[len(header):], maskedPayload)
		} else {
			copy(result[len(header):], frame.Payload)
		}
	}

	return result
}

func ParseClosePayload(payload []byte) (int, string) {
	if len(payload) < 2 {
		return CloseNoStatusReceived, ""
	}
	code := int(binary.BigEndian.Uint16(payload[:2]))
	reason := ""
	if len(payload) > 2 {
		reason = string(payload[2:])
	}
	return code, reason
}
