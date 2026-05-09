package rs

import "github.com/solocoder/reed-solomon/internal/rs"

type Shard = rs.Shard

var (
	ErrInvalidParams     = rs.ErrInvalidParams
	ErrInsufficientShards = rs.ErrInsufficientShards
	ErrAllShardsLost     = rs.ErrAllShardsLost
	ErrInvalidShardIndex = rs.ErrInvalidShardIndex
	ErrDecodingFailed    = rs.ErrDecodingFailed
	ErrVerificationFailed = rs.ErrVerificationFailed
)

func Encode(data []byte, N, M int) ([]Shard, int, error) {
	return rs.Encode(data, N, M)
}

func Decode(shards []Shard, N, M int, originalLen int) ([]byte, error) {
	return rs.Decode(shards, N, M, originalLen)
}

func Verify(shards []Shard, N, M int) error {
	return rs.Verify(shards, N, M)
}

func Join(shards [][]byte) []byte {
	return rs.Join(shards)
}
