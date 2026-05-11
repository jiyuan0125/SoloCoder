package mqtt

type WillMessage struct {
	Topic   string
	Payload []byte
	QoS     QoS
	Retain  bool
}

func NewWillMessage(topic string, payload []byte, qos QoS, retain bool) (*WillMessage, error) {
	if topic == "" {
		return nil, ErrInvalidTopicFilter
	}
	if !qos.IsValid() {
		return nil, ErrInvalidQoS
	}
	return &WillMessage{
		Topic:   topic,
		Payload: payload,
		QoS:     qos,
		Retain:  retain,
	}, nil
}

func (wm *WillMessage) Validate() error {
	if wm.Topic == "" {
		return ErrInvalidTopicFilter
	}
	if !wm.QoS.IsValid() {
		return ErrInvalidQoS
	}
	return nil
}

func (wm *WillMessage) ToPublish(packetID uint16) (*PublishPacket, error) {
	if err := wm.Validate(); err != nil {
		return nil, err
	}

	publish := &PublishPacket{
		Dup:     false,
		QoS:     wm.QoS,
		Retain:  wm.Retain,
		Topic:   wm.Topic,
		Payload: make([]byte, len(wm.Payload)),
	}
	copy(publish.Payload, wm.Payload)

	if wm.QoS > QoS0 {
		if packetID == 0 {
			return nil, ErrInvalidPacketID
		}
		publish.PacketID = packetID
	}

	return publish, nil
}

func BuildConnectWithWill(
	clientID string,
	cleanSession bool,
	keepAlive uint16,
	username string,
	password []byte,
	will *WillMessage,
) (*ConnectPacket, error) {
	p := &ConnectPacket{
		ProtocolName:  ProtocolName,
		ProtocolLevel: ProtocolLevel311,
		CleanSession:  cleanSession,
		KeepAlive:     keepAlive,
		ClientID:      clientID,
	}

	if username != "" {
		p.UsernameFlag = true
		p.Username = username
		if password != nil {
			p.PasswordFlag = true
			p.Password = password
		} else {
			p.PasswordFlag = false
		}
	}

	if will != nil {
		if err := will.Validate(); err != nil {
			return nil, err
		}
		p.WillFlag = true
		p.WillQoS = will.QoS
		p.WillRetain = will.Retain
		p.WillTopic = will.Topic
		p.WillMessage = will.Payload
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

func BuildWillPublish(will *WillMessage, packetID uint16) (*PublishPacket, error) {
	return will.ToPublish(packetID)
}
