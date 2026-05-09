package zstdlib

import "zstd-tool/protocol"

const (
	HeaderSize        = 18
	DefaultBufferSize = 64 * 1024
)

var (
	LevelToInt = map[protocol.CompressionLevel]int{
		protocol.LevelFastest: 1,
		protocol.LevelDefault: 3,
		protocol.LevelBest:    19,
	}

	IntToLevel = map[int]protocol.CompressionLevel{
		1:  protocol.LevelFastest,
		3:  protocol.LevelDefault,
		19: protocol.LevelBest,
	}
)

func ValidateLevel(level protocol.CompressionLevel) bool {
	_, ok := LevelToInt[level]
	return ok
}
