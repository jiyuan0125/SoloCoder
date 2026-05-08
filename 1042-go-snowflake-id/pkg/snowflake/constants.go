package snowflake

const (
	TimestampBits  = 41
	MachineIDBits  = 10
	SequenceBits   = 12

	TimestampMask  = (1 << TimestampBits) - 1
	MachineIDMask  = (1 << MachineIDBits) - 1
	SequenceMask   = (1 << SequenceBits) - 1

	TimestampShift = MachineIDBits + SequenceBits
	MachineIDShift = SequenceBits

	MaxMachineID   = MachineIDMask
	MaxSequence    = SequenceMask
	DefaultEpoch   = 1704067200000
	DefaultMaxWait  = 500
)
