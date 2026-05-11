package idgen

import (
	"time"

	"idgen/api"
)

const (
	TimestampBits uint8 = 41
	NodeIDBits    uint8 = 10
	SequenceBits  uint8 = 12

	MaxSequence  int64 = (1 << SequenceBits) - 1
	MaxNodeIDVal int64 = (1 << NodeIDBits) - 1
)

var Epoch = time.Date(api.EpochYear, time.Month(api.EpochMonth), api.EpochDay, 0, 0, 0, 0, time.UTC).UnixMilli()
