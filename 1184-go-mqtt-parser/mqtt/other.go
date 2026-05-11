package mqtt

type PingreqPacket struct{}

func (p *PingreqPacket) Type() PacketType {
	return PacketTypePINGREQ
}

func (p *PingreqPacket) Flags() byte {
	return 0
}

func (p *PingreqPacket) Encode() ([]byte, error) {
	fh := &FixedHeader{
		PacketType:      PacketTypePINGREQ,
		RemainingLength: 0,
	}
	return EncodeFixedHeader(fh)
}

func (p *PingreqPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, _, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePINGREQ {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 0 {
		return ErrUnexpectedEndOfPacket
	}

	return nil
}

type PingrespPacket struct{}

func (p *PingrespPacket) Type() PacketType {
	return PacketTypePINGRESP
}

func (p *PingrespPacket) Flags() byte {
	return 0
}

func (p *PingrespPacket) Encode() ([]byte, error) {
	fh := &FixedHeader{
		PacketType:      PacketTypePINGRESP,
		RemainingLength: 0,
	}
	return EncodeFixedHeader(fh)
}

func (p *PingrespPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, _, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypePINGRESP {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 0 {
		return ErrUnexpectedEndOfPacket
	}

	return nil
}

type DisconnectPacket struct{}

func (p *DisconnectPacket) Type() PacketType {
	return PacketTypeDISCONNECT
}

func (p *DisconnectPacket) Flags() byte {
	return 0
}

func (p *DisconnectPacket) Encode() ([]byte, error) {
	fh := &FixedHeader{
		PacketType:      PacketTypeDISCONNECT,
		RemainingLength: 0,
	}
	return EncodeFixedHeader(fh)
}

func (p *DisconnectPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, _, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeDISCONNECT {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 0 {
		return ErrUnexpectedEndOfPacket
	}

	return nil
}
