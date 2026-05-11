package mqtt

type ConnectPacket struct {
	ProtocolName    string
	ProtocolLevel   byte
	CleanSession    bool
	WillFlag        bool
	WillQoS         QoS
	WillRetain      bool
	UsernameFlag    bool
	PasswordFlag    bool
	KeepAlive       uint16
	ClientID        string
	WillTopic       string
	WillMessage     []byte
	Username        string
	Password        []byte
}

func (p *ConnectPacket) Type() PacketType {
	return PacketTypeCONNECT
}

func (p *ConnectPacket) Flags() byte {
	return 0
}

func (p *ConnectPacket) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var payload []byte
	payload = append(payload, EncodeString(p.ClientID)...)

	if p.WillFlag {
		payload = append(payload, EncodeString(p.WillTopic)...)
		willMsgLen := len(p.WillMessage)
		payload = append(payload, byte(willMsgLen>>8), byte(willMsgLen))
		payload = append(payload, p.WillMessage...)
	}

	if p.UsernameFlag {
		payload = append(payload, EncodeString(p.Username)...)
	}
	if p.PasswordFlag {
		pwdLen := len(p.Password)
		payload = append(payload, byte(pwdLen>>8), byte(pwdLen))
		payload = append(payload, p.Password...)
	}

	var vh []byte
	vh = append(vh, EncodeString(p.ProtocolName)...)
	vh = append(vh, p.ProtocolLevel)

	var connectFlags byte
	if p.CleanSession {
		connectFlags |= 0x02
	}
	if p.WillFlag {
		connectFlags |= 0x04
	}
	connectFlags |= (byte(p.WillQoS) << 3) & 0x18
	if p.WillRetain {
		connectFlags |= 0x20
	}
	if p.PasswordFlag {
		connectFlags |= 0x40
	}
	if p.UsernameFlag {
		connectFlags |= 0x80
	}
	vh = append(vh, connectFlags)

	vh = append(vh, byte(p.KeepAlive>>8), byte(p.KeepAlive))

	remainingLength := uint32(len(vh) + len(payload))
	fh := &FixedHeader{
		PacketType:      PacketTypeCONNECT,
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

func (p *ConnectPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeCONNECT {
		return ErrInvalidPacketType
	}

	if int(fh.RemainingLength) != len(data)-offset {
		return ErrUnexpectedEndOfPacket
	}

	protocolName, offset, err := DecodeString(data, offset)
	if err != nil {
		return err
	}
	p.ProtocolName = protocolName

	protocolLevel, offset, err := DecodeByte(data, offset)
	if err != nil {
		return err
	}
	p.ProtocolLevel = protocolLevel

	connectFlags, offset, err := DecodeByte(data, offset)
	if err != nil {
		return err
	}

	p.UsernameFlag = (connectFlags & 0x80) != 0
	p.PasswordFlag = (connectFlags & 0x40) != 0
	p.WillRetain = (connectFlags & 0x20) != 0
	p.WillQoS = QoS((connectFlags >> 3) & 0x03)
	p.WillFlag = (connectFlags & 0x04) != 0
	p.CleanSession = (connectFlags & 0x02) != 0

	if p.WillFlag {
		if !p.WillQoS.IsValid() {
			return ErrInvalidQoS
		}
	} else {
		if p.WillQoS != 0 || p.WillRetain {
			return ErrInvalidFlags
		}
	}

	if p.UsernameFlag && !p.PasswordFlag {
		return ErrUsernameFlagWithoutPassword
	}

	keepAlive, offset, err := DecodeUint16(data, offset)
	if err != nil {
		return err
	}
	p.KeepAlive = keepAlive

	clientID, offset, err := DecodeString(data, offset)
	if err != nil {
		return err
	}
	p.ClientID = clientID

	if p.WillFlag {
		willTopic, newOffset, err := DecodeString(data, offset)
		if err != nil {
			return err
		}
		p.WillTopic = willTopic
		offset = newOffset

		willMsgLen, offset, err := DecodeUint16(data, offset)
		if err != nil {
			return err
		}
		if offset+int(willMsgLen) > len(data) {
			return ErrPacketTooShort
		}
		p.WillMessage = make([]byte, willMsgLen)
		copy(p.WillMessage, data[offset:offset+int(willMsgLen)])
		offset += int(willMsgLen)
	}

	if p.UsernameFlag {
		username, newOffset, err := DecodeString(data, offset)
		if err != nil {
			return err
		}
		p.Username = username
		offset = newOffset
	}

	if p.PasswordFlag {
		pwdLen, offset, err := DecodeUint16(data, offset)
		if err != nil {
			return err
		}
		if offset+int(pwdLen) > len(data) {
			return ErrPacketTooShort
		}
		p.Password = make([]byte, pwdLen)
		copy(p.Password, data[offset:offset+int(pwdLen)])
		offset += int(pwdLen)
	}

	return p.Validate()
}

func (p *ConnectPacket) Validate() error {
	if p.ProtocolName != ProtocolName {
		return ErrInvalidProtocolName
	}
	if p.ProtocolLevel != ProtocolLevel311 {
		return ErrInvalidProtocolLevel
	}
	if p.WillFlag {
		if p.WillTopic == "" {
			return ErrWillFlagSetWithoutWill
		}
		if !p.WillQoS.IsValid() {
			return ErrInvalidQoS
		}
	}
	if p.UsernameFlag && !p.PasswordFlag {
		return ErrUsernameFlagWithoutPassword
	}
	if p.ClientID == "" && !p.CleanSession {
		return ErrClientIDEmptyWithCleanSession0
	}
	return nil
}

type ConnackReturnCode byte

const (
	ConnackAccepted                     ConnackReturnCode = 0
	ConnackUnacceptableProtocolVersion  ConnackReturnCode = 1
	ConnackIdentifierRejected           ConnackReturnCode = 2
	ConnackServerUnavailable            ConnackReturnCode = 3
	ConnackBadUsernameOrPassword        ConnackReturnCode = 4
	ConnackNotAuthorized                ConnackReturnCode = 5
)

func (rc ConnackReturnCode) String() string {
	switch rc {
	case ConnackAccepted:
		return "Connection Accepted"
	case ConnackUnacceptableProtocolVersion:
		return "Unacceptable Protocol Version"
	case ConnackIdentifierRejected:
		return "Identifier Rejected"
	case ConnackServerUnavailable:
		return "Server Unavailable"
	case ConnackBadUsernameOrPassword:
		return "Bad Username or Password"
	case ConnackNotAuthorized:
		return "Not Authorized"
	default:
		return "Unknown"
	}
}

type ConnackPacket struct {
	SessionPresent bool
	ReturnCode     ConnackReturnCode
}

func (p *ConnackPacket) Type() PacketType {
	return PacketTypeCONNACK
}

func (p *ConnackPacket) Flags() byte {
	return 0
}

func (p *ConnackPacket) Encode() ([]byte, error) {
	vh := make([]byte, 2)
	if p.SessionPresent {
		vh[0] = 0x01
	}
	vh[1] = byte(p.ReturnCode)

	fh := &FixedHeader{
		PacketType:      PacketTypeCONNACK,
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

func (p *ConnackPacket) Decode(data []byte) error {
	if len(data) < 2 {
		return ErrPacketTooShort
	}

	fh, offset, err := DecodeFixedHeader(data, 0)
	if err != nil {
		return err
	}

	if fh.PacketType != PacketTypeCONNACK {
		return ErrInvalidPacketType
	}
	if fh.RemainingLength != 2 {
		return ErrUnexpectedEndOfPacket
	}

	connectAcknowledgeFlags, offset, err := DecodeByte(data, offset)
	if err != nil {
		return err
	}
	p.SessionPresent = (connectAcknowledgeFlags & 0x01) != 0

	returnCode, _, err := DecodeByte(data, offset)
	if err != nil {
		return err
	}
	p.ReturnCode = ConnackReturnCode(returnCode)

	return nil
}
