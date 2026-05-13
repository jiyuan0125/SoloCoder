package pool

import (
	"context"
	"time"

	"shardpool/internal/consistent"
)

type PoolStats struct {
	TotalShards        int
	TotalActiveConns   int
	TotalIdleConns     int
	TotalRequests      int64
	AvgAcquireLatency  time.Duration
	ShardDistribution  map[int]float64
}

type ShardPool struct {
	cfg          ShardConfig
	shards       map[int]*Shard
	consistent   *consistent.Consistent
}

func NewShardPool(cfg ShardConfig) *ShardPool {
	shards := make(map[int]*Shard)
	consistent := consistent.New(cfg.VirtualNodeCount, nil)

	for i := 0; i < cfg.ShardCount; i++ {
		shards[i] = NewShard(i, cfg)
		consistent.AddNode(i)
	}

	return &ShardPool{
		cfg:        cfg,
		shards:     shards,
		consistent: consistent,
	}
}

func (p *ShardPool) GetShard(key string) *Shard {
	shardID := p.consistent.Get(key)
	if shardID < 0 {
		return nil
	}
	return p.shards[shardID]
}

func (p *ShardPool) Acquire(ctx context.Context, key string) (*Connection, int, error) {
	shard := p.GetShard(key)
	if shard == nil {
		return nil, -1, context.DeadlineExceeded
	}

	conn, err := shard.Acquire(ctx)
	if err != nil {
		return nil, shard.ID(), err
	}

	return conn, shard.ID(), nil
}

func (p *ShardPool) Release(key string, connID string) {
	shard := p.GetShard(key)
	if shard == nil {
		return
	}
	shard.Release(connID)
}

func (p *ShardPool) AllShardsStats() []ShardStats {
	stats := make([]ShardStats, 0, len(p.shards))
	for i := 0; i < len(p.shards); i++ {
		stats = append(stats, p.shards[i].Stats())
	}
	return stats
}

func (p *ShardPool) Stats() PoolStats {
	shardStats := p.AllShardsStats()

	stats := PoolStats{
		TotalShards:       len(p.shards),
		ShardDistribution: make(map[int]float64),
	}

	var totalLatency time.Duration
	var totalRequests int64

	for _, ss := range shardStats {
		stats.TotalActiveConns += ss.ActiveConns
		stats.TotalIdleConns += ss.IdleConns
		stats.TotalRequests += ss.TotalRequests
		totalRequests += ss.TotalRequests
		totalLatency += time.Duration(ss.TotalRequests) * ss.AcquireLatency
	}

	if totalRequests > 0 {
		stats.AvgAcquireLatency = totalLatency / time.Duration(totalRequests)
	}

	for _, ss := range shardStats {
		if totalRequests > 0 {
			stats.ShardDistribution[ss.ShardID] = float64(ss.TotalRequests) / float64(totalRequests)
		} else {
			stats.ShardDistribution[ss.ShardID] = 0
		}
	}

	return stats
}

func (p *ShardPool) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			for _, shard := range p.shards {
				shard.Cleanup()
			}
		}
	}()
}
