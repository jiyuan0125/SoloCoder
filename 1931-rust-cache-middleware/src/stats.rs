use serde::Serialize;

#[derive(Clone, Serialize)]
pub struct Stats {
    pub l1_hits: u64,
    pub l1_misses: u64,
    pub l2_hits: u64,
    pub l2_misses: u64,
    pub l1_miss_l2_hit: u64,
}

impl Stats {
    pub fn new() -> Self {
        Self {
            l1_hits: 0,
            l1_misses: 0,
            l2_hits: 0,
            l2_misses: 0,
            l1_miss_l2_hit: 0,
        }
    }
}

#[derive(Serialize)]
pub struct StatsResponse {
    pub l1: LevelStats,
    pub l2: LevelStats,
    pub l1_miss_l2_hit: u64,
}

#[derive(Serialize)]
pub struct LevelStats {
    pub hits: u64,
    pub misses: u64,
    pub size: usize,
    pub hit_rate: Option<f64>,
}

impl StatsResponse {
    pub fn new(stats: Stats, l1_size: usize, l2_size: usize) -> Self {
        let l1_total = stats.l1_hits + stats.l1_misses;
        let l2_total = stats.l2_hits + stats.l2_misses;

        Self {
            l1: LevelStats {
                hits: stats.l1_hits,
                misses: stats.l1_misses,
                size: l1_size,
                hit_rate: if l1_total > 0 {
                    Some(stats.l1_hits as f64 / l1_total as f64)
                } else {
                    None
                },
            },
            l2: LevelStats {
                hits: stats.l2_hits,
                misses: stats.l2_misses,
                size: l2_size,
                hit_rate: if l2_total > 0 {
                    Some(stats.l2_hits as f64 / l2_total as f64)
                } else {
                    None
                },
            },
            l1_miss_l2_hit: stats.l1_miss_l2_hit,
        }
    }
}
