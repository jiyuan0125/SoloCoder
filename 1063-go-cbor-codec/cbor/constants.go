package cbor

const (
	MajorTypeUnsignedInt  byte = 0
	MajorTypeNegativeInt  byte = 1
	MajorTypeByteString   byte = 2
	MajorTypeTextString   byte = 3
	MajorTypeArray        byte = 4
	MajorTypeMap          byte = 5
	MajorTypeTag          byte = 6
	MajorTypeSimpleFloat  byte = 7
)

const (
	AdditionalInfo8Bit    byte = 24
	AdditionalInfo16Bit   byte = 25
	AdditionalInfo32Bit   byte = 26
	AdditionalInfo64Bit   byte = 27
	AdditionalInfoIndef   byte = 31
)

const (
	SimpleFalse       byte = 20
	SimpleTrue        byte = 21
	SimpleNull        byte = 22
	SimpleUndefined   byte = 23
	SimpleHalfFloat   byte = 25
	SimpleSingleFloat byte = 26
	SimpleDoubleFloat byte = 27
	BreakCode         byte = 31
)

type MajorType byte

type SimpleValue byte
