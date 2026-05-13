use dashmap::DashMap;
use parking_lot::RwLock;
use std::collections::{HashMap, HashSet, VecDeque};
use chrono::{DateTime, Utc};
use crate::models::{Span, Trace, DependencyEdge, Stats};
use crate::models::DependencyGraph;

const MAX_TRACES: usize = 10000;

pub struct AppState {
    pub traces: DashMap<String, Trace>,
    pub trace_order: RwLock<VecDeque<String>>,
    pub p99_threshold_ms: u64,
}

impl AppState {
    pub fn new(p99_threshold_ms: u64) -> Self {
        Self {
            traces: DashMap::new(),
            trace_order: RwLock::new(VecDeque::new()),
            p99_threshold_ms,
        }
    }

    pub fn add_span(&self, span: Span) {
        let trace_id = span.trace_id.clone();
        let span_duration = span.duration;
        let span_start = span.start_time;

        let should_create = !self.traces.contains_key(&trace_id);

        if should_create {
            let mut order = self.trace_order.write();
            order.push_back(trace_id.clone());
            if order.len() > MAX_TRACES {
                if let Some(removed) = order.pop_front() {
                    self.traces.remove(&removed);
                }
            }
        }

        self.traces
            .entry(trace_id.clone())
            .and_modify(|trace| {
                trace.spans.push(span.clone());
                if span_start < trace.earliest_start {
                    trace.earliest_start = span_start;
                }
                trace.total_duration = trace.total_duration.max(span_duration);
            })
            .or_insert_with(|| Trace {
                trace_id: trace_id.clone(),
                spans: vec![span],
                earliest_start: span_start,
                total_duration: span_duration,
            });
    }

    pub fn get_trace(&self, trace_id: &str) -> Option<Trace> {
        self.traces.get(trace_id).map(|t| t.clone())
    }

    pub fn list_traces(
        &self,
        service_name: Option<&str>,
        min_duration: Option<u64>,
        start_time: Option<DateTime<Utc>>,
        end_time: Option<DateTime<Utc>>,
    ) -> Vec<Trace> {
        self.traces
            .iter()
            .filter(|entry| {
                let trace = entry.value();
                
                if let Some(svc) = service_name {
                    if !trace.spans.iter().any(|s| s.service_name == svc) {
                        return false;
                    }
                }

                if let Some(min_dur) = min_duration {
                    if trace.total_duration < min_dur {
                        return false;
                    }
                }

                if let Some(start) = start_time {
                    if trace.earliest_start < start {
                        return false;
                    }
                }

                if let Some(end) = end_time {
                    if trace.earliest_start > end {
                        return false;
                    }
                }

                true
            })
            .map(|entry| entry.value().clone())
            .collect()
    }

    pub fn get_dependency_graph(&self) -> DependencyGraph {
        let mut edge_data: HashMap<(String, String), (u64, u64)> = HashMap::new();
        let mut nodes: HashMap<String, ()> = HashMap::new();

        for trace_entry in self.traces.iter() {
            let trace = trace_entry.value();
            let span_map: HashMap<&str, &Span> = trace
                .spans
                .iter()
                .map(|s| (s.span_id.as_str(), s))
                .collect();

            for span in &trace.spans {
                nodes.insert(span.service_name.clone(), ());

                if let Some(parent_id) = &span.parent_span_id {
                    if let Some(parent_span) = span_map.get(parent_id.as_str()) {
                        let from = parent_span.service_name.clone();
                        let to = span.service_name.clone();
                        let key = (from, to);
                        let entry = edge_data.entry(key).or_insert((0, 0));
                        entry.0 += 1;
                        entry.1 += span.duration;
                    }
                }
            }
        }

        let edges: Vec<DependencyEdge> = edge_data
            .into_iter()
            .map(|((from, to), (count, total_latency))| DependencyEdge {
                from,
                to,
                call_count: count,
                avg_latency_ms: if count > 0 {
                    total_latency as f64 / count as f64
                } else {
                    0.0
                },
            })
            .collect();

        let nodes: Vec<String> = nodes.into_keys().collect();
        let cycles = self.detect_cycles(&nodes, &edges);

        DependencyGraph {
            nodes,
            edges,
            cycles,
        }
    }

    fn detect_cycles(&self, nodes: &[String], edges: &[DependencyEdge]) -> Vec<Vec<String>> {
        let mut adjacency: HashMap<String, Vec<String>> = HashMap::new();
        for edge in edges {
            adjacency
                .entry(edge.from.clone())
                .or_default()
                .push(edge.to.clone());
        }

        let mut cycles: Vec<Vec<String>> = Vec::new();
        let mut visited: HashSet<String> = HashSet::new();
        let mut rec_stack: HashSet<String> = HashSet::new();
        let mut path: Vec<String> = Vec::new();

        for node in nodes {
            Self::dfs_cycle(
                node.clone(),
                &adjacency,
                &mut visited,
                &mut rec_stack,
                &mut path,
                &mut cycles,
            );
        }

        cycles
    }

    fn dfs_cycle(
        node: String,
        adjacency: &HashMap<String, Vec<String>>,
        visited: &mut HashSet<String>,
        rec_stack: &mut HashSet<String>,
        path: &mut Vec<String>,
        cycles: &mut Vec<Vec<String>>,
    ) -> bool {
        if rec_stack.contains(&node) {
            if let Some(idx) = path.iter().position(|p| p == &node) {
                let cycle: Vec<String> = path[idx..].to_vec();
                if cycle.len() >= 2 {
                    cycles.push(cycle);
                }
                return true;
            }
            return true;
        }

        if visited.contains(&node) {
            return false;
        }

        visited.insert(node.clone());
        rec_stack.insert(node.clone());
        path.push(node.clone());

        if let Some(neighbors) = adjacency.get(&node) {
            for neighbor in neighbors {
                if Self::dfs_cycle(
                    neighbor.clone(),
                    adjacency,
                    visited,
                    rec_stack,
                    path,
                    cycles,
                ) {
                    path.pop();
                    rec_stack.remove(&node);
                    return true;
                }
            }
        }

        path.pop();
        rec_stack.remove(&node);
        false
    }

    pub fn get_stats(&self) -> Stats {
        let mut all_spans: Vec<u64> = Vec::new();
        for entry in self.traces.iter() {
            for span in &entry.value().spans {
                all_spans.push(span.duration);
            }
        }

        let total_spans = all_spans.len();
        let total_duration: u64 = all_spans.iter().sum();
        let avg_latency = if total_spans > 0 {
            total_duration as f64 / total_spans as f64
        } else {
            0.0
        };

        let mut sorted = all_spans.clone();
        sorted.sort_unstable();

        let p50 = Self::percentile(&sorted, 0.50);
        let p95 = Self::percentile(&sorted, 0.95);
        let p99 = Self::percentile(&sorted, 0.99);

        let slow_count = all_spans
            .iter()
            .filter(|&&d| d > self.p99_threshold_ms)
            .count();
        let slow_ratio = if total_spans > 0 {
            slow_count as f64 / total_spans as f64
        } else {
            0.0
        };

        Stats {
            total_traces: self.traces.len() as u64,
            avg_latency_ms: avg_latency,
            p50_ms: p50,
            p95_ms: p95,
            p99_ms: p99,
            slow_span_ratio: slow_ratio,
        }
    }

    fn percentile(sorted: &[u64], p: f64) -> f64 {
        if sorted.is_empty() {
            return 0.0;
        }
        let idx = (sorted.len() as f64) * p;
        let idx = idx.floor() as usize;
        let idx = idx.min(sorted.len() - 1);
        sorted[idx] as f64
    }
}
