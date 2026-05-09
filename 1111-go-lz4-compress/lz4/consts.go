package lz4

const (
	MagicNumber         uint32 = 0x184D2204
	MagicNumberSkippableMin uint32 = 0x184D2A50
	MagicNumberSkippableMax uint32 = 0x184D2A5F

	VersionBits         = 2
	VersionCurrent      = 1

	MinMatch            = 4
	MaxOffset           = 65535
	LastLiterals        = 5
	MFLimit             = 12

	TokenLiteralMask    = 0xF0
	TokenMatchMask      = 0x0F
	TokenLiteralShift   = 4

	UncompressedFlag    = uint32(1 << 31)
	BlockSizeMask       = 0x7FFFFFFF

	MaxBlockSize        = 1 << 24
)
