package snowflake

type Status struct {
	LastTimestamp    int64 `json:"last_timestamp"`
	CurrentTimestamp int64 `json:"current_timestamp"`
	MachineID        int64 `json:"machine_id"`
	LastSequence     int64 `json:"last_sequence"`
	Epoch            int64 `json:"epoch"`
	MaxSequence      int64 `json:"max_sequence"`
}

func ParseID(id int64, epoch int64) ParsedID {
	timestamp := (id >> TimestampShift) & TimestampMask
	machineID := (id >> MachineIDShift) & MachineIDMask
	sequence := id & SequenceMask
	
	return ParsedID{
		ID:             id,
		Timestamp:      timestamp,
		RealTimestamp:  epoch + timestamp,
		MachineID:      machineID,
		Sequence:       sequence,
	}
}

func ParseIDDefaultEpoch(id int64) ParsedID {
	return ParseID(id, DefaultEpoch)
}

type ParsedID struct {
	ID            int64 `json:"id"`
	Timestamp     int64 `json:"timestamp"`
	RealTimestamp int64 `json:"real_timestamp"`
	MachineID     int64 `json:"machine_id"`
	Sequence      int64 `json:"sequence"`
}
