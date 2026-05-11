package mqtt

import "errors"

var (
	ErrInvalidRemainingLength = errors.New("invalid remaining length")
	ErrInvalidPacketType      = errors.New("invalid packet type")
	ErrInvalidFlags           = errors.New("invalid flags")
	ErrInvalidQoS             = errors.New("invalid QoS")
	ErrInvalidPacketID        = errors.New("invalid packet ID")
	ErrInvalidProtocolName    = errors.New("invalid protocol name")
	ErrInvalidProtocolLevel   = errors.New("invalid protocol level")
	ErrClientIDEmptyWithCleanSession0 = errors.New("client id cannot be empty when clean session is 0")
	ErrWillFlagSetWithoutWill = errors.New("will flag set but will topic or message missing")
	ErrUsernameFlagWithoutPassword = errors.New("username flag set but password flag not set")
	ErrInvalidTopicFilter     = errors.New("invalid topic filter")
	ErrPacketTooShort         = errors.New("packet too short")
	ErrNotEnoughBytes         = errors.New("not enough bytes")
	ErrUnexpectedEndOfPacket  = errors.New("unexpected end of packet")
)
