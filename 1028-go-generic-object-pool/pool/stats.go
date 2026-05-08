package pool

type Stats struct {
	Free      int
	Created   int64
	Discarded int64
	Hits      int64
	Misses    int64
	HitRate   float64
}
