package core

type Format int

const (
	FormatUnknown Format = iota
	FormatPOSIX
	FormatGNU
)

const (
	blockSize    = 512
	nameLen      = 100
	prefixLen    = 155
	modeLen      = 8
	uidLen       = 8
	gidLen       = 8
	sizeLen      = 12
	mtimeLen     = 12
	chksumLen    = 8
	linknameLen  = 100
	magicLen     = 6
	versionLen   = 2
	unameLen     = 32
	gnameLen     = 32
	devmajorLen  = 8
	devminorLen  = 8
)

const (
	magicPOSIX = "ustar\x00"
	versionPOSIX = "00"
	magicGNU   = "ustar "
	versionGNU = " \x00"
)

const (
	TypeReg      = '0'
	TypeLink     = '1'
	TypeSymlink  = '2'
	TypeChar     = '3'
	TypeBlock    = '4'
	TypeDir      = '5'
	TypeFifo     = '6'
	TypeCont     = '7'
	TypeXHeader  = 'x'
	TypeGNULongName = 'L'
	TypeGNUSparse = 'S'
)

func parseMagic(buf []byte) Format {
	if len(buf) < magicLen+versionLen {
		return FormatUnknown
	}
	magic := string(buf[257 : 257+magicLen])
	version := string(buf[263 : 263+versionLen])
	if magic == magicPOSIX && version == versionPOSIX {
		return FormatPOSIX
	}
	if magic == magicGNU && version == versionGNU {
		return FormatGNU
	}
	return FormatUnknown
}
