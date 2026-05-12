use std::collections::HashMap;
use std::time::Instant;

#[derive(Debug, Clone, Default)]
pub struct RouteStats {
    pub request_count: u64,
}

#[derive(Debug, Clone, Default)]
pub struct AggregateStats {
    pub request_count: u64,
    pub error_count: u64,
    pub total_duration_ms: u64,
    pub start_times: Vec<Instant>,
}

#[derive(Debug, Default)]
pub struct StatsCollector {
    route_stats: HashMap<String, RouteStats>,
    aggregate_stats: HashMap<String, AggregateStats>,
}

impl StatsCollector {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn record_route(&mut self, path: &str) {
        self.route_stats.entry(path.to_string())
            .or_insert_with(RouteStats::default)
            .request_count += 1;
    }

    pub fn record_aggregate_start(&mut self, scene: &str) {
        let stats = self.aggregate_stats.entry(scene.to_string())
            .or_insert_with(AggregateStats::default);
        stats.start_times.push(Instant::now());
    }

    pub fn record_aggregate_complete(&mut self, scene: &str) -> Option<u64> {
        let stats = self.aggregate_stats.get_mut(scene)?;
        if let Some(start) = stats.start_times.pop() {
            let duration = start.elapsed().as_millis() as u64;
            stats.request_count += 1;
            stats.total_duration_ms += duration;
            Some(duration)
        } else {
            None
        }
    }

    pub fn record_aggregate_error(&mut self, scene: &str) {
        self.aggregate_stats.entry(scene.to_string())
            .or_insert_with(AggregateStats::default)
            .error_count += 1;
    }

    pub fn get_route_stats(&self, path: &str) -> Option<&RouteStats> {
        self.route_stats.get(path)
    }

    pub fn get_aggregate_stats(&self, scene: &str) -> Option<AggregateStatsSnapshot> {
        self.aggregate_stats.get(scene).map(|s| AggregateStatsSnapshot {
            request_count: s.request_count,
            error_count: s.error_count,
            avg_duration_ms: if s.request_count > 0 {
                s.total_duration_ms / s.request_count
            } else {
                0
            },
        })
    }
}

#[derive(Debug, Clone, Copy)]
pub struct AggregateStatsSnapshot {
    pub request_count: u64,
    pub error_count: u64,
    pub avg_duration_ms: u64,
}
