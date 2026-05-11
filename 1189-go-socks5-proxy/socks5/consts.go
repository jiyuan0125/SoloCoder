package socks5

const (
	Version = 0x05

	MethodNoAuth       = 0x00
	MethodUserPassAuth = 0x02
	MethodNoAcceptable = 0xFF

	AuthUserPassVersion = 0x01
	AuthSuccess = 0x00
	AuthFailure = 0x01

	CmdConnect   = 0x01
	CmdUDPAssociate = 0x03

	AddrTypeIPv4   = 0x01
	AddrTypeDomain = 0x03
	AddrTypeIPv6   = 0x04

	ReplySuccess              = 0x00
	ReplyFailure             = 0x01
	ReplyConnNotAllowed        = 0x02
	ReplyNetworkUnreachable  = 0x03
	ReplyHostUnreachable     = 0x04
	ReplyConnRefused          = 0x05
	ReplyTTLExpired           = 0x06
	ReplyCmdNotSupported    = 0x07
	ReplyAddrTypeNotSupported = 0x08

	RSV = 0x00
)
