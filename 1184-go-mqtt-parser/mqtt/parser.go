package mqtt

func Parse(data []byte) (Packet, error) {
	if len(data) < 2 {
		return nil, ErrPacketTooShort
	}

	fh, _, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return nil, err
	}

	switch fh.PacketType {
	case PacketTypeCONNECT:
		p := &ConnectPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypeCONNACK:
		p := &ConnackPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePUBLISH:
		p := &PublishPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePUBACK:
		p := &PubackPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePUBREC:
		p := &PubrecPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePUBREL:
		p := &PubrelPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePUBCOMP:
		p := &PubcompPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypeSUBSCRIBE:
		p := &SubscribePacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypeSUBACK:
		p := &SubackPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypeUNSUBSCRIBE:
		p := &UnsubscribePacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypeUNSUBACK:
		p := &UnsubackPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePINGREQ:
		p := &PingreqPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypePINGRESP:
		p := &PingrespPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	case PacketTypeDISCONNECT:
		p := &DisconnectPacket{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, ErrInvalidPacketType
	}
}

func GetFixedHeader(data []byte) (*FixedHeader, error) {
	if len(data) < 2 {
		return nil, ErrPacketTooShort
	}
	fh, _, err := DecodeFixedHeader(data, 0)
	return fh, err
}
