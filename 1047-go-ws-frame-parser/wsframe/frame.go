package wsframe

const (
	OpCodeContinuation = 0x0
	OpCodeText         = 0x1
	OpCodeBinary       = 0x2
	OpCodeClose        = 0x8
	OpCodePing         = 0x9
	OpCodePong         = 0xA
)

const (
	CloseNormalClosure       = 1000
	CloseGoingAway           = 1001
	CloseProtocolError       = 1002
	CloseUnsupportedData     = 1003
	CloseNoStatusReceived    = 1005
	CloseAbnormalClosure     = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation     = 1008
	CloseMessageTooBig       = 1009
	CloseMandatoryExtension  = 1010
	CloseInternalServerError = 1011
	CloseServiceRestart      = 1012
	CloseTryAgainLater       = 1013
)

const (
	DefaultMaxMessageSize = 1 << 20
	MaxControlFramePayload = 125
	MinHeaderSize = 2
	MaxHeaderSize = 14
	MaskKeySize   = 4
)

type Frame struct {
	Fin     bool
	Rsv1    bool
	Rsv2    bool
	Rsv3    bool
	OpCode  int
	Masked  bool
	MaskKey [4]byte
	Payload []byte
}

func (f *Frame) IsControl() bool {
	return f.OpCode >= OpCodeClose
}

func (f *Frame) IsData() bool {
	return f.OpCode == OpCodeText || f.OpCode == OpCodeBinary || f.OpCode == OpCodeContinuation
}
