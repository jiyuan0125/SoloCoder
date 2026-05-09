package rs

import "errors"

var (
	ErrInvalidParams     = errors.New("invalid parameters: N >= 1, M >= 1, N+M <= 255")
	ErrInsufficientShards = errors.New("insufficient shards: need at least N shards")
	ErrAllShardsLost     = errors.New("all shards are lost")
	ErrNonSquareMatrix   = errors.New("matrix is not square")
	ErrSingularMatrix    = errors.New("matrix is singular and cannot be inverted")
	ErrMatrixMismatch    = errors.New("matrix dimensions mismatch")
	ErrInvalidShardIndex = errors.New("invalid shard index")
	ErrDecodingFailed    = errors.New("decoding failed")
	ErrVerificationFailed = errors.New("verification failed: cannot reconstruct from given shards")
)
