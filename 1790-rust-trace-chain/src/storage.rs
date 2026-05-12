use std::collections::{HashMap, HashSet};
use chrono::{Utc, Duration};
use crate::model::{Span, PaginatedResponse, TraceSpanDisplay};

pub struct TraceStorage {
    spans_by_trace: HashMap<String, Vec<Span>>,
    spans_by_id: HashMap<String, Span>,
    tag_index: HashMap<(String, String), HashSet<String>>,
}

impl TraceStorage {
    pub fn new() -> Self {
        Self {
            spans_by_trace: HashMap::new(),
            spans_by_id: HashMap::new(),
            tag_index: HashMap::new(),
        }
    }

    pub fn add_span(&mut self, span: Span) {
        let span_id = span.span_id.clone();
        let trace_id = span.trace_id.clone();
        
        for (key, value) in &span.tags {
            self.tag_index
                .entry((key.clone(), value.clone()))
                .or_insert_with(HashSet::new)
                .insert(span_id.clone());
        }

        self.spans_by_trace
            .entry(trace_id)
            .or_insert_with(Vec::new)
            .push(span.clone());
        
        self.spans_by_id.insert(span_id, span);
    }

    pub fn end_span(&mut self, span_id: &str) -> Option<Span> {
        let span = self.spans_by_id.get_mut(span_id)?;
        span.end_time = Some(Utc::now());
        
        let updated_span = span.clone();
        
        if let Some(spans) = self.spans_by_trace.get_mut(&updated_span.trace_id) {
            if let Some(existing) = spans.iter_mut().find(|s| s.span_id == span_id) {
                *existing = updated_span.clone();
            }
        }
        
        Some(updated_span)
    }

    pub fn get_trace(&self, trace_id: &str) -> Option<Vec<TraceSpanDisplay>> {
        let spans = self.spans_by_trace.get(trace_id)?;
        
        let mut sorted_spans: Vec<Span> = spans.clone();
        sorted_spans.sort_by(|a, b| {
            b.start_time.cmp(&a.start_time)
                .then_with(|| a.span_id.cmp(&b.span_id))
        });

        let mut span_by_id: HashMap<String, Span> = HashMap::new();
        let mut children_map: HashMap<Option<String>, Vec<String>> = HashMap::new();
        
        for span in &sorted_spans {
            span_by_id.insert(span.span_id.clone(), span.clone());
            children_map
                .entry(span.parent_span_id.clone())
                .or_insert_with(Vec::new)
                .push(span.span_id.clone());
        }

        let mut result: Vec<TraceSpanDisplay> = Vec::new();
        let mut visited: HashSet<String> = HashSet::new();
        
        for span in &sorted_spans {
            if visited.contains(&span.span_id) {
                continue;
            }
            
            self.collect_span_tree(
                &span.span_id,
                &span_by_id,
                &children_map,
                &mut visited,
                0,
                &mut result,
            );
        }
        
        Some(result)
    }

    fn collect_span_tree(
        &self,
        span_id: &str,
        span_by_id: &HashMap<String, Span>,
        children_map: &HashMap<Option<String>, Vec<String>>,
        visited: &mut HashSet<String>,
        level: usize,
        result: &mut Vec<TraceSpanDisplay>,
    ) {
        if visited.contains(span_id) {
            return;
        }
        visited.insert(span_id.to_string());
        
        if let Some(span) = span_by_id.get(span_id) {
            result.push(TraceSpanDisplay {
                span: span.clone(),
                indent_level: level,
            });
            
            if let Some(children) = children_map.get(&Some(span_id.to_string())) {
                let mut children_clone = children.clone();
                children_clone.sort_by(|a, b| {
                    let span_a = span_by_id.get(a).unwrap();
                    let span_b = span_by_id.get(b).unwrap();
                    span_b.start_time.cmp(&span_a.start_time)
                });
                
                for child_id in children_clone {
                    self.collect_span_tree(&child_id, span_by_id, children_map, visited, level + 1, result);
                }
            }
        }
    }

    pub fn search_by_tags(
        &self,
        tags: &HashMap<String, String>,
        page: usize,
        page_size: usize,
    ) -> PaginatedResponse<Span> {
        if tags.is_empty() {
            return PaginatedResponse {
                total: 0,
                page,
                page_size,
                items: Vec::new(),
            };
        }

        let mut span_ids: Option<HashSet<String>> = None;
        
        for (key, value) in tags {
            let key_span_ids = self.tag_index
                .get(&(key.clone(), value.clone()))
                .cloned()
                .unwrap_or_else(HashSet::new);
            
            span_ids = match span_ids {
                Some(existing) => Some(existing.intersection(&key_span_ids).cloned().collect()),
                None => Some(key_span_ids),
            };
        }

        let span_ids = span_ids.unwrap_or_else(HashSet::new);
        
        let mut matching_spans: Vec<Span> = span_ids
            .iter()
            .filter_map(|id| self.spans_by_id.get(id).cloned())
            .collect();
        
        matching_spans.sort_by(|a, b| {
            b.start_time.cmp(&a.start_time)
                .then_with(|| a.span_id.cmp(&b.span_id))
        });

        let total = matching_spans.len();
        let start = (page.saturating_sub(1)) * page_size;
        let items: Vec<Span> = matching_spans.into_iter().skip(start).take(page_size).collect();

        PaginatedResponse {
            total,
            page,
            page_size,
            items,
        }
    }

    pub fn cleanup_trace(&mut self, trace_id: &str) -> usize {
        if let Some(spans) = self.spans_by_trace.remove(trace_id) {
            for span in &spans {
                self.spans_by_id.remove(&span.span_id);
                
                for (key, value) in &span.tags {
                    if let Some(set) = self.tag_index.get_mut(&(key.clone(), value.clone())) {
                        set.remove(&span.span_id);
                        if set.is_empty() {
                            self.tag_index.remove(&(key.clone(), value.clone()));
                        }
                    }
                }
            }
            spans.len()
        } else {
            0
        }
    }

    pub fn cleanup_old_spans(&mut self, hours: Option<i64>) -> usize {
        let hours = hours.unwrap_or(24);
        let cutoff = Utc::now() - Duration::hours(hours);
        
        let traces_to_remove: Vec<String> = self.spans_by_trace
            .iter()
            .filter(|(_, spans)| {
                spans.iter().all(|s| s.start_time <= cutoff)
            })
            .map(|(trace_id, _)| trace_id.clone())
            .collect();

        let mut total_removed = 0;
        for trace_id in traces_to_remove {
            total_removed += self.cleanup_trace(&trace_id);
        }
        total_removed
    }
}

impl Default for TraceStorage {
    fn default() -> Self {
        Self::new()
    }
}
