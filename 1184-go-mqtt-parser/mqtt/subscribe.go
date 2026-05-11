package mqtt

type TopicSubscription struct {
	TopicFilter string
	RequestedQoS QoS
}

type SubscribePacket struct {
	PacketID uint16
	Subscriptions []TopicSubscription
}

func (p *SubscribePacket) Type() PacketType {
	return PacketTypeSUBSCRIBE
}

func (p *SubscribePacket) Flags() byte {
	return 0x02
}

func (p *SubscribePacket) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var payload []byte
	for _, sub := range p.Subscriptions {
		payload = append(payload, EncodeString(sub.TopicFilter)...)
		payload = append(payload, byte(sub.RequestedQoS))
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	remainingLength := uint32(len(vh) + len(payload))

	fh := &FixedHeader{
		PacketType:      PacketTypeSUBSCRIBE,
		RemainingLength: remainingLength,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}
	fhBytes[len(fhBytes)-2] |= 0x02

	result := make([]byte, 0, len(fhBytes)+len(vh)+len(payload))
	result = append(result, fhBytes...)
	result = append(result, vh...)
	result = append(result, payload...)

	return result, nil
}

func (p *SubscribePacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeSUBSCRIBE {
		return ErrInvalidPacketType
	}
	if int(fh.RemainingLength) != len(data)-offset {
		return ErrUnexpectedEndOfPacket
	}

	packetID, offset, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	p.Subscriptions = nil
	currentOffset := offset
	for currentOffset < len(data) {
		topicFilter, newOffset, err := DecodeString(data, currentOffset)
		if err != nil {
			return err
		}
		currentOffset = newOffset

		qosByte, newOffset, err := DecodeByte(data, currentOffset)
		if err != nil {
			return err
		}
		currentOffset = newOffset

		qos := QoS(qosByte)
		if !qos.IsValid() {
			return ErrInvalidQoS
		}

		p.Subscriptions = append(p.Subscriptions, TopicSubscription{
			TopicFilter:  topicFilter,
			RequestedQoS: qos,
		})
	}

	return p.Validate()
}

func (p *SubscribePacket) Validate() error {
	if p.PacketID == 0 {
		return ErrInvalidPacketID
	}
	if len(p.Subscriptions) == 0 {
		return ErrInvalidTopicFilter
	}
	for _, sub := range p.Subscriptions {
		if sub.TopicFilter == "" {
			return ErrInvalidTopicFilter
		}
		if !sub.RequestedQoS.IsValid() {
			return ErrInvalidQoS
		}
	}
	return nil
}

type SubackReturnCode byte

const (
	SubackSuccessQoS0 SubackReturnCode = 0
	SubackSuccessQoS1 SubackReturnCode = 1
	SubackSuccessQoS2 SubackReturnCode = 2
	SubackFailure     SubackReturnCode = 0x80
)

func (rc SubackReturnCode) String() string {
	switch rc {
	case SubackSuccessQoS0:
		return "Success - QoS 0"
	case SubackSuccessQoS1:
		return "Success - QoS 1"
	case SubackSuccessQoS2:
		return "Success - QoS 2"
	case SubackFailure:
		return "Failure"
	default:
		return "Unknown"
	}
}

type SubackPacket struct {
	PacketID      uint16
	ReturnCodes   []SubackReturnCode
}

func (p *SubackPacket) Type() PacketType {
	return PacketTypeSUBACK
}

func (p *SubackPacket) Flags() byte {
	return 0
}

func (p *SubackPacket) Encode() ([]byte, error) {
	if p.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}
	if len(p.ReturnCodes) == 0 {
		return nil, ErrInvalidTopicFilter
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	payload := make([]byte, len(p.ReturnCodes))
	for i, rc := range p.ReturnCodes {
		payload[i] = byte(rc)
	}

	remainingLength := uint32(len(vh) + len(payload))

	fh := &FixedHeader{
		PacketType:      PacketTypeSUBACK,
		RemainingLength: remainingLength,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, len(fhBytes)+len(vh)+len(payload))
	result = append(result, fhBytes...)
	result = append(result, vh...)
	result = append(result, payload...)

	return result, nil
}

func (p *SubackPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeSUBACK {
		return ErrInvalidPacketType
	}
	if int(fh.RemainingLength) != len(data)-offset {
		return ErrUnexpectedEndOfPacket
	}

	packetID, offset, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	p.ReturnCodes = nil
	for offset < len(data) {
		rc, newOffset, err := DecodeByte(data, offset)
		if err != nil {
			return err
		}
		p.ReturnCodes = append(p.ReturnCodes, SubackReturnCode(rc))
		offset = newOffset
	}

	if len(p.ReturnCodes) == 0 {
		return ErrInvalidTopicFilter
	}

	return nil
}

type UnsubscribePacket struct {
	PacketID     uint16
	TopicFilters []string
}

func (p *UnsubscribePacket) Type() PacketType {
	return PacketTypeUNSUBSCRIBE
}

func (p *UnsubscribePacket) Flags() byte {
	return 0x02
}

func (p *UnsubscribePacket) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var payload []byte
	for _, topicFilter := range p.TopicFilters {
		payload = append(payload, EncodeString(topicFilter)...)
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	remainingLength := uint32(len(vh) + len(payload))

	fh := &FixedHeader{
		PacketType:      PacketTypeUNSUBSCRIBE,
		RemainingLength: remainingLength,
	}
	fhBytes, err := EncodeFixedHeader(fh)
	if err != nil {
		return nil, err
	}
	fhBytes[len(fhBytes)-2] |= 0x02

	result := make([]byte, 0, len(fhBytes)+len(vh)+len(payload))
	result = append(result, fhBytes...)
	result = append(result, vh...)
	result = append(result, payload...)

	return result, nil
}

func (p *UnsubscribePacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeUNSUBSCRIBE {
		return ErrInvalidPacketType
	}
	if int(fh.RemainingLength) != len(data)-offset {
		return ErrUnexpectedEndOfPacket
	}

	packetID, offset, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	if packetID == 0 {
		return ErrInvalidPacketID
	}
	p.PacketID = packetID

	p.TopicFilters = nil
	for offset < len(data) {
		topicFilter, newOffset, err := DecodeString(data, offset)
		if err != nil {
			return err
		}
		p.TopicFilters = append(p.TopicFilters, topicFilter)
		offset = newOffset
	}

	return p.Validate()
}

func (p *UnsubscribePacket) Validate() error {
	if p.PacketID == 0 {
		return ErrInvalidPacketID
	}
	if len(p.TopicFilters) == 0 {
		return ErrInvalidTopicFilter
	}
	for _, tf := range p.TopicFilters {
		if tf == "" {
			return ErrInvalidTopicFilter
		}
	}
	return nil
}

type UnsubackPacket struct {
	PacketID uint16
}

func (p *UnsubackPacket) Type() PacketType {
	return PacketTypeUNSUBACK
}

func (p *UnsubackPacket) Flags() byte {
	return 0
}

func (p *UnsubackPacket) Encode() ([]byte, error) {
	if p.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}

	vh := []byte{byte(p.PacketID >> 8), byte(p.PacketID)}

	fh := &FixedHeader{
		PacketType:      PacketTypeUNSUBACK,
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

func (p *UnsubackPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeUNSUBACK {
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
