package mqtt

type PublishPacket struct {
	Dup       bool
	QoS       QoS
	Retain    bool
	Topic     string
	PacketID  uint16
	Payload   []byte
}

func (p *PublishPacket) Type() PacketType {
	return PacketTypePUBLISH
}

func (p *PublishPacket) Flags() byte {
	var flags byte
	if p.Dup {
		flags |= 0x08
	}
	flags |= (byte(p.QoS) << 1) & 0x06
	if p.Retain {
		flags |= 0x01
	}
	return flags
}

func (p *PublishPacket) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var vh []byte
	vh = append(vh, EncodeString(p.Topic)...)

	if p.QoS > QoS0 {
		vh = append(vh, byte(p.PacketID>>8), byte(p.PacketID))
	}

	remainingLength := uint32(len(vh) + len(p.Payload))

	fh := &FixedHeader{
		PacketType:      PacketTypePUBLISH,
		Dup:             p.Dup,
		QoS:             p.QoS,
		Retain:          p.Retain,
		RemainingLength: remainingLength,
	}

	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, len(fhBytes)+len(vh)+len(p.Payload))
	result = append(result, fhBytes...)
	result = append(result, vh...)
	result = append(result, p.Payload...)

	return result, nil
}

func (p *PublishPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePUBLISH {
		return ErrInvalidPacketType
	}

	if int(fh.RemainingLength) != len(data)-offset {
		return ErrUnexpectedEndOfPacket
	}

	p.Dup = fh.Dup
	p.QoS = fh.QoS
	p.Retain = fh.Retain

	if !p.QoS.IsValid() {
		return ErrInvalidQoS
	}

	if p.Dup && p.QoS == QoS0 {
		return ErrInvalidFlags
	}

	topic, offset, err := DecodeString(data, offset)
	if err != nil {
		return err
	}
	p.Topic = topic

	if p.QoS > QoS0 {
		packetID, newOffset, err := DecodeUint16(data, offset)
		if err != nil {
			return err
		}
		if packetID == 0 {
			return ErrInvalidPacketID
		}
		p.PacketID = packetID
		offset = newOffset
	}

	p.Payload = make([]byte, len(data)-offset)
	copy(p.Payload, data[offset:])

	return nil
}

func (p *PublishPacket) Validate() error {
	if !p.QoS.IsValid() {
		return ErrInvalidQoS
	}
	if p.Dup && p.QoS == QoS0 {
		return ErrInvalidFlags
	}
	if p.QoS > QoS0 && p.PacketID == 0 {
		return ErrInvalidPacketID
	}
	return nil
}

type PubackPacket struct {
	PacketID uint16
}

func (p *PubackPacket) Type() PacketType {
	return PacketTypePUBACK
}

func (p *PubackPacket) Flags() byte {
	return 0
}

func (p *PubackPacket) Encode() ([]byte, error) {
	if p.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	fh := &FixedHeader{
		PacketType:      PacketTypePUBACK,
		RemainingLength: 2,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, len(fhBytes)+2)
	result = append(result, fhBytes...)
	result = append(result, vh...)

	return result, nil
}

func (p *PubackPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePUBACK {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 2 {
		return ErrUnexpectedEndOfPacket
	}

	packetID, _, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	return nil
}

type PubrecPacket struct {
	PacketID uint16
}

func (p *PubrecPacket) Type() PacketType {
	return PacketTypePUBREC
}

func (p *PubrecPacket) Flags() byte {
	return 0
}

func (p *PubrecPacket) Encode() ([]byte, error) {
	if p.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	fh := &FixedHeader{
		PacketType:      PacketTypePUBREC,
		RemainingLength: 2,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, len(fhBytes)+2)
	result = append(result, fhBytes...)
	result = append(result, vh...)

	return result, nil
}

func (p *PubrecPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePUBREC {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 2 {
		return ErrUnexpectedEndOfPacket
	}

	packetID, _, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	return nil
}

type PubrelPacket struct {
	PacketID uint16
}

func (p *PubrelPacket) Type() PacketType {
	return PacketTypePUBREL
}

func (p *PubrelPacket) Flags() byte {
	return 0x02
}

func (p *PubrelPacket) Encode() ([]byte, error) {
	if p.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	fh := &FixedHeader{
		PacketType:      PacketTypePUBREL,
		RemainingLength: 2,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}
	fhBytes[len(fhBytes)-2] |= 0x02

	result := make([]byte, 0, len(fhBytes)+2)
	result = append(result, fhBytes...)
	result = append(result, vh...)

	return result, nil
}

func (p *PubrelPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePUBREL {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 2 {
		return ErrUnexpectedEndOfPacket
	}

	packetID, _, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	return nil
}

type PubcompPacket struct {
	PacketID uint16
}

func (p *PubcompPacket) Type() PacketType {
	return PacketTypePUBCOMP
}

func (p *PubcompPacket) Flags() byte {
	return 0
}

func (p *PubcompPacket) Encode() ([]byte, error) {
	if p.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	fh := &FixedHeader{
		PacketType:      PacketTypePUBCOMP,
		RemainingLength: 2,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, len(fhBytes)+2)
	result = append(result, fhBytes...)
	result = append(result, vh...)

	return result, nil
}

func (p *PubcompPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePUBCOMP {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 2 {
		return ErrUnexpectedEndOfPacket
	}

	packetID, _, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	return nil
}
