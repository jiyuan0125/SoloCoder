package mqtt

const (
	ProtocolName    = "MQTT"
	ProtocolLevel311 = 4

	MaxRemainingLength = 268435455
	MaxPacketID        = 65535
)

type PacketType byte

const (
	PacketTypeReserved    PacketType = 0
	PacketTypeCONNECT     PacketType = 1
	PacketTypeCONNACK     PacketType = 2
	PacketTypePUBLISH     PacketType = 3
	PacketTypePUBACK      PacketType = 4
	PacketTypePUBREC      PacketType = 5
	PacketTypePUBREL      PacketType = 6
	PacketTypePUBCOMP     PacketType = 7
	PacketTypeSUBSCRIBE   PacketType = 8
	PacketTypeSUBACK      PacketType = 9
	PacketTypeUNSUBSCRIBE PacketType = 10
	PacketTypeUNSUBACK    PacketType = 11
	PacketTypePINGREQ     PacketType = 12
	PacketTypePINGRESP    PacketType = 13
	PacketTypeDISCONNECT  PacketType = 14
	PacketTypeReserved2   PacketType = 15
)

func (pt PacketType) String() string {
	switch pt {
	case PacketTypeCONNECT:
		return "CONNECT"
	case PacketTypeCONNACK:
		return "CONNACK"
	case PacketTypePUBLISH:
		return "PUBLISH"
	case PacketTypePUBACK:
		return "PUBACK"
	case PacketTypePUBREC:
		return "PUBREC"
	case PacketTypePUBREL:
		return "PUBREL"
	case PacketTypePUBCOMP:
		return "PUBCOMP"
	case PacketTypeSUBSCRIBE:
		return "SUBSCRIBE"
	case PacketTypeSUBACK:
		return "SUBACK"
	case PacketTypeUNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case PacketTypeUNSUBACK:
		return "UNSUBACK"
	case PacketTypePINGREQ:
		return "PINGREQ"
	case PacketTypePINGRESP:
		return "PINGRESP"
	case PacketTypeDISCONNECT:
		return "DISCONNECT"
	default:
		return "RESERVED"
	}
}

func (pt PacketType) IsValid() bool {
	return pt >= PacketTypeCONNECT && pt <= PacketTypeDISCONNECT
}

type QoS byte

const (
	QoS0 QoS = 0
	QoS1 QoS = 1
	QoS2 QoS = 2
)

func (q QoS) IsValid() bool {
	return q == QoS0 || q == QoS1 || q == QoS2
}

func (q QoS) String() string {
	switch q {
	case QoS0:
		return "QoS 0"
	case QoS1:
		return "QoS 1"
	case QoS2:
		return "QoS 2"
	default:
		return "Invalid"
	}
}

type Packet interface {
	Type() PacketType
	Flags() byte
	Encode() ([]byte, error)
	Decode([]byte) error
}

type FixedHeader struct {
	PacketType      PacketType
	Dup             bool
	QoS             QoS
	Retain          bool
	RemainingLength uint32
}

func (fh *FixedHeader) EncodeFlags() byte {
	var flags byte
	if fh.Dup {
		flags |= 0x08
	}
	flags |= (byte(fh.QoS) << 1) & 0x06
	if fh.Retain {
		flags |= 0x01
	}
	return flags
}

func (fh *FixedHeader) DecodeFlags(flags byte) {
	fh.Dup = (flags & 0x08) != 0
	fh.QoS = QoS((flags >> 1) & 0x03)
	fh.Retain = (flags & 0x01) != 0
}

func (fh *FixedHeader) Validate() error {
	if !fh.PacketType.IsValid() {
		return ErrInvalidPacketType
	}
	if !fh.QoS.IsValid() {
		return ErrInvalidQoS
	}
	return nil
}

func EncodeString(s string) []byte {
	length := len(s)
	buf := make([]byte, 2+length)
	buf[0] = byte(length >> 8)
	buf[1] = byte(length)
	copy(buf[2:], s)
	return buf
}

func DecodeString(data []byte, offset int) (string, int, error) {
	if offset+2 > len(data) {
		return "", offset, ErrPacketTooShort
	}
	length := int(data[offset])<<8 | int(data[offset+1])
	if offset+2+length > len(data) {
		return "", offset, ErrPacketTooShort
	}
	s := string(data[offset+2 : offset+2+length])
	return s, offset + 2 + length, nil
}

func EncodeUint16(v uint16) []byte {
	return []byte{byte(v >> 8), byte(v)}
}

func DecodeUint16(data []byte, offset int) (uint16, int, error) {
	if offset+2 > len(data) {
		return 0, offset, ErrPacketTooShort
	}
	v := uint16(data[offset])<<8 | uint16(data[offset+1])
	return v, offset + 2, nil
}

func EncodeByte(v byte) []byte {
	return []byte{v}
}

func DecodeByte(data []byte, offset int) (byte, int, error) {
	if offset >= len(data) {
		return 0, offset, ErrPacketTooShort
	}
	return data[offset], offset + 1, nil
}
