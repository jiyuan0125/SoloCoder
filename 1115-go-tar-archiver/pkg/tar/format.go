package tar

const (
	BlockSize = 512
	NameSize  = 100
	PrefixSize = 155
)

type Format int

const (
	FormatPOSIX Format = iota
	FormatGNU
)

const (
	TypeFlagRegFile     = '0'
	TypeFlagRegFileAlt  = '\x00'
	TypeFlagLink        = '1'
	TypeFlagSymlink     = '2'
	TypeFlagChar        = '3'
	TypeFlagBlock       = '4'
	TypeFlagDir         = '5'
	TypeFlagFIFO        = '6'
	TypeFlagCont        = '7'
	TypeFlagXHeader     = 'x'
	TypeFlagXGlobalHead = 'g'
	TypeFlagGNULongName = 'L'
	TypeFlagGNULongLink = 'K'
	TypeFlagGNUSparse   = 'S'
)

const (
	MagicPOSIX = "ustar\x00"
	MagicGNU   = "ustar "
	VersionPOSIX = "00"
	VersionGNU   = " \x00"
)
