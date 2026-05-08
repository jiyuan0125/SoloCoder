package wsframe

import (
	"encoding/binary"
	"unicode/utf8"
)

func ParseFrame(data []byte) (*Frame, int, error) {
	if len(data) < MinHeaderSize {
		return nil, 0, ErrInsufficientData
	}

	frame := &Frame{}

	firstByte := data[0]
	frame.Fin = (firstByte & 0x80) != 0
	frame.Rsv1 = (firstByte & 0x40) != 0
	frame.Rsv2 = (firstByte & 0x20) != 0
	frame.Rsv3 = (firstByte & 0x10) != 0
	frame.OpCode = int(firstByte & 0x0F)

	if frame.Rsv1 || frame.Rsv2 || frame.Rsv3 {
		return nil, 0, ErrReservedBitSet
	}

	if !isValidOpcode(frame.OpCode) {
		return nil, 0, ErrInvalidOpcode
	}

	secondByte := data[1]
	frame.Masked = (secondByte & 0x80) != 0
	length7 := int(secondByte & 0x7F)

	if frame.IsControl() && !frame.Fin {
		return nil, 0, ErrControlFrameFragmented
	}

	offset := 2
	var payloadLength int64

	switch length7 {
	case 126:
		if len(data) < offset+2 {
			return nil, 0, ErrInsufficientData
		}
		payloadLength = int64(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
	case 127:
		if len(data) < offset+8 {
			return nil, 0, ErrInsufficientData
		}
		payloadLength = int64(binary.BigEndian.Uint64(data[offset : offset+8]))
		offset += 8
	default:
		payloadLength = int64(length7)
	}

	if frame.IsControl() && payloadLength > MaxControlFramePayload {
		return nil, 0, ErrControlFrameTooLarge
	}

	if frame.Masked {
		if len(data) < offset+MaskKeySize {
			return nil, 0, ErrInsufficientData
		}
		copy(frame.MaskKey[:], data[offset:offset+MaskKeySize])
		offset += MaskKeySize
	}

	totalLength := offset + int(payloadLength)
	if len(data) < totalLength {
		return nil, 0, ErrInsufficientData
	}

	frame.Payload = make([]byte, payloadLength)
	if payloadLength > 0 {
		copy(frame.Payload, data[offset:totalLength])
	}

	if frame.Masked {
		unmaskData(frame.Payload, frame.MaskKey)
	}

	if frame.OpCode == OpCodeText {
		if !utf8.Valid(frame.Payload) {
			return nil, 0, ErrInvalidUTF8
		}
	}

	return frame, totalLength, nil
}

func isValidOpcode(opcode int) bool {
	switch opcode {
	case OpCodeContinuation, OpCodeText, OpCodeBinary, OpCodeClose, OpCodePing, OpCodePong:
		return true
	default:
		return false
	}
}

func unmaskData(data []byte, key [4]byte) {
	for i := range data {
		data[i] ^= key[i%4]
	}
}
