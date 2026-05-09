package rs

type Shard struct {
	Index int
	Data  []byte
}

func Encode(data []byte, N, M int) ([]Shard, int, error) {
	if N < 1 || M < 1 || N+M > 255 {
		return nil, 0, ErrInvalidParams
	}

	originalLen := len(data)
	shardSize := (originalLen + N - 1) / N
	paddedLen := shardSize * N
	padded := make([]byte, paddedLen)
	copy(padded, data)

	shards := make([][]byte, N+M)
	for i := 0; i < N; i++ {
		shards[i] = make([]byte, shardSize)
		start := i * shardSize
		copy(shards[i], padded[start:start+shardSize])
	}

	for i := N; i < N+M; i++ {
		shards[i] = make([]byte, shardSize)
	}

	generator := Vandermonde(N+M, N)

	for bytePos := 0; bytePos < shardSize; bytePos++ {
		dataVec := make([]byte, N)
		for i := 0; i < N; i++ {
			dataVec[i] = shards[i][bytePos]
		}

		for row := 0; row < N+M; row++ {
			sum := byte(0)
			for col := 0; col < N; col++ {
				sum = Add(sum, Mul(generator[row][col], dataVec[col]))
			}
			shards[row][bytePos] = sum
		}
	}

	result := make([]Shard, N+M)
	for i := 0; i < N+M; i++ {
		result[i] = Shard{Index: i, Data: shards[i]}
	}

	return result, originalLen, nil
}

func Verify(shards []Shard, N, M int) error {
	if N < 1 || M < 1 || N+M > 255 {
		return ErrInvalidParams
	}

	if len(shards) < N {
		return ErrInsufficientShards
	}

	if len(shards) == 0 {
		return ErrAllShardsLost
	}

	recoveryMatrix := NewMatrix(N, N)
	for i := 0; i < N; i++ {
		idx := shards[i].Index
		if idx < 0 || idx >= N+M {
			return ErrInvalidShardIndex
		}
		for j := 0; j < N; j++ {
			recoveryMatrix[i][j] = Pow(byte(idx), j)
		}
	}

	_, err := recoveryMatrix.Invert()
	return err
}

func Decode(shards []Shard, N, M int, originalLen int) ([]byte, error) {
	if N < 1 || M < 1 || N+M > 255 {
		return nil, ErrInvalidParams
	}

	if len(shards) < N {
		return nil, ErrInsufficientShards
	}

	if len(shards) == 0 {
		return nil, ErrAllShardsLost
	}

	shardSize := len(shards[0].Data)
	for _, s := range shards {
		if len(s.Data) != shardSize {
			return nil, ErrDecodingFailed
		}
	}

	recoveryMatrix := NewMatrix(N, N)
	for i := 0; i < N; i++ {
		idx := shards[i].Index
		if idx < 0 || idx >= N+M {
			return nil, ErrInvalidShardIndex
		}
		for j := 0; j < N; j++ {
			recoveryMatrix[i][j] = Pow(byte(idx), j)
		}
	}

	invMatrix, err := recoveryMatrix.Invert()
	if err != nil {
		return nil, ErrVerificationFailed
	}

	recoveredShards := make([][]byte, N)
	for i := 0; i < N; i++ {
		recoveredShards[i] = make([]byte, shardSize)
	}

	for bytePos := 0; bytePos < shardSize; bytePos++ {
		shardVec := make([]byte, N)
		for i := 0; i < N; i++ {
			shardVec[i] = shards[i].Data[bytePos]
		}

		for i := 0; i < N; i++ {
			sum := byte(0)
			for j := 0; j < N; j++ {
				sum = Add(sum, Mul(invMatrix[i][j], shardVec[j]))
			}
			recoveredShards[i][bytePos] = sum
		}
	}

	result := Join(recoveredShards)
	if originalLen > 0 && originalLen < len(result) {
		result = result[:originalLen]
	}

	return result, nil
}

func Join(shards [][]byte) []byte {
	var result []byte
	for _, s := range shards {
		result = append(result, s...)
	}
	return result
}
