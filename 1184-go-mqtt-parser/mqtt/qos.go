package mqtt

type QoSState int

const (
	QoSStateIdle QoSState = iota
	QoS1StateWaitPuback
	QoS2StateWaitPubrec
	QoS2StateWaitPubcomp
)

func (s QoSState) String() string {
	switch s {
	case QoSStateIdle:
		return "Idle"
	case QoS1StateWaitPuback:
		return "QoS 1 - Waiting for PUBACK"
	case QoS2StateWaitPubrec:
		return "QoS 2 - Waiting for PUBREC"
	case QoS2StateWaitPubcomp:
		return "QoS 2 - Waiting for PUBCOMP"
	default:
		return "Unknown"
	}
}

type QoSMessage struct {
	PacketID   uint16
	State      QoSState
	Publish    *PublishPacket
	RetryCount int
}

func NewQoS1Message(publish *PublishPacket) (*QoSMessage, error) {
	if publish.QoS != QoS1 {
		return nil, ErrInvalidQoS
	}
	if publish.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}
	return &QoSMessage{
		PacketID: publish.PacketID,
		State:    QoS1StateWaitPuback,
		Publish:  publish,
	}, nil
}

func NewQoS2Message(publish *PublishPacket) (*QoSMessage, error) {
	if publish.QoS != QoS2 {
		return nil, ErrInvalidQoS
	}
	if publish.PacketID == 0 {
		return nil, ErrInvalidPacketID
	}
	return &QoSMessage{
		PacketID: publish.PacketID,
		State:    QoS2StateWaitPubrec,
		Publish:  publish,
	}, nil
}

func (qm *QoSMessage) HandlePuback() (bool, error) {
	if qm.State != QoS1StateWaitPuback {
		return false, ErrInvalidPacketType
	}
	qm.State = QoSStateIdle
	return true, nil
}

func (qm *QoSMessage) HandlePubrec() (*PubrelPacket, error) {
	if qm.State != QoS2StateWaitPubrec {
		return nil, ErrInvalidPacketType
	}
	qm.State = QoS2StateWaitPubcomp
	return &PubrelPacket{PacketID: qm.PacketID}, nil
}

func (qm *QoSMessage) HandlePubcomp() (bool, error) {
	if qm.State != QoS2StateWaitPubcomp {
		return false, ErrInvalidPacketType
	}
	qm.State = QoSStateIdle
	return true, nil
}

func (qm *QoSMessage) CreateRetryPublish() (*PublishPacket, error) {
	if qm.Publish == nil {
		return nil, ErrInvalidPacketType
	}
	if qm.Publish.QoS == QoS0 {
		return nil, ErrInvalidQoS
	}

	retry := &PublishPacket{
		Dup:      true,
		QoS:      qm.Publish.QoS,
		Retain:   qm.Publish.Retain,
		Topic:    qm.Publish.Topic,
		PacketID: qm.Publish.PacketID,
		Payload:  make([]byte, len(qm.Publish.Payload)),
	}
	copy(retry.Payload, qm.Publish.Payload)
	qm.RetryCount++

	return retry, nil
}

func BuildPuback(packetID uint16) (*PubackPacket, error) {
	if packetID == 0 {
		return nil, ErrInvalidPacketID
	}
	return &PubackPacket{PacketID: packetID}, nil
}

func BuildPubrec(packetID uint16) (*PubrecPacket, error) {
	if packetID == 0 {
		return nil, ErrInvalidPacketID
	}
	return &PubrecPacket{PacketID: packetID}, nil
}

func BuildPubcomp(packetID uint16) (*PubcompPacket, error) {
	if packetID == 0 {
		return nil, ErrInvalidPacketID
	}
	return &PubcompPacket{PacketID: packetID}, nil
}
