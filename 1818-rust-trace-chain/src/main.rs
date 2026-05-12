use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::Mutex;
use warp::{Filter, Rejection, Reply};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc, Duration as ChronoDuration};

const MAX_SPANS: usize = 10000;
const EVICT_PERCENTAGE: f32 = 0.1;
const DEFAULT_PAGE_SIZE: usize = 20;
const CLEANUP_HOURS: i64 = 24;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum TagValue {
    String(String),
    Number(f64),
}

impl PartialEq for TagValue {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (TagValue::String(a), TagValue::String(b)) => a == b,
            (TagValue::Number(a), TagValue::Number(b)) => a == b,
            _ => false,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Span {
    pub span_id: String,
    pub trace_id: String,
    pub parent_span_id: Option<String>,
    pub name: String,
    pub start_time: DateTime<Utc>,
    pub end_time: Option<DateTime<Utc>>,
    pub tags: HashMap<String, TagValue>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SpanCreateRequest {
    pub trace_id: Option<String>,
    pub parent_span_id: Option<String>,
    pub name: String,
    pub tags: Option<HashMap<String, TagValue>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SpanEndRequest {
    pub end_time: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize)]
pub struct TreeSpan {
    pub span: Span,
    pub children: Vec<TreeSpan>,
}

#[derive(Debug, Clone, Serialize)]
pub struct PaginatedResponse<T> {
    pub data: Vec<T>,
    pub page: usize,
    pub page_size: usize,
    pub total: usize,
}

pub type Storage = Arc<Mutex<HashMap<String, Span>>>;

pub fn new_storage() -> Storage {
    Arc::new(Mutex::new(HashMap::new()))
}

pub async fn create_span(
    storage: Storage,
    req: SpanCreateRequest,
) -> Result<Span, Rejection> {
    let span_id = Uuid::new_v4().to_string();
    let trace_id = req.trace_id.unwrap_or_else(|| Uuid::new_v4().to_string());
    
    let span = Span {
        span_id: span_id.clone(),
        trace_id,
        parent_span_id: req.parent_span_id,
        name: req.name,
        start_time: Utc::now(),
        end_time: None,
        tags: req.tags.unwrap_or_default(),
    };
    
    let mut storage = storage.lock().await;
    let total_count = storage.len();
    
    if total_count >= MAX_SPANS {
        evict_oldest(&mut storage);
    }
    
    storage.insert(span_id, span.clone());
    
    Ok(span)
}

pub async fn end_span(
    storage: Storage,
    span_id: String,
    req: SpanEndRequest,
) -> Result<Option<Span>, Rejection> {
    let mut storage = storage.lock().await;
    
    if let Some(span) = storage.get_mut(&span_id) {
        span.end_time = Some(req.end_time.unwrap_or_else(Utc::now));
        return Ok(Some(span.clone()));
    }
    
    Ok(None)
}

pub async fn get_trace(
    storage: Storage,
    trace_id: String,
) -> Result<Option<Vec<TreeSpan>>, Rejection> {
    let storage = storage.lock().await;
    
    let mut trace_spans: Vec<Span> = storage
        .values()
        .filter(|s| s.trace_id == trace_id)
        .cloned()
        .collect();
    
    if trace_spans.is_empty() {
        return Ok(None);
    }
    
    trace_spans.sort_by(|a, b| a.start_time.cmp(&b.start_time));
    
    let tree = build_tree(trace_spans);
    
    Ok(Some(tree))
}

fn build_tree(spans: Vec<Span>) -> Vec<TreeSpan> {
    let mut span_map: HashMap<String, TreeSpan> = HashMap::new();
    let mut parent_children_map: HashMap<String, Vec<String>> = HashMap::new();
    let mut root_ids = Vec::new();
    
    for span in &spans {
        let tree_span = TreeSpan {
            span: span.clone(),
            children: Vec::new(),
        };
        span_map.insert(span.span_id.clone(), tree_span);
        
        if let Some(parent_id) = &span.parent_span_id {
            parent_children_map
                .entry(parent_id.clone())
                .or_default()
                .push(span.span_id.clone());
        } else {
            root_ids.push(span.span_id.clone());
        }
    }
    
    for parent_id in parent_children_map.keys() {
        if span_map.get_mut(parent_id).is_none() {
            if let Some(children) = parent_children_map.get(parent_id) {
                for child_id in children {
                    root_ids.push(child_id.clone());
                }
            }
        }
    }
    
    for (parent_id, child_ids) in parent_children_map {
        if span_map.contains_key(&parent_id) {
            let mut children = Vec::new();
            for child_id in child_ids {
                if let Some(child) = span_map.remove(&child_id) {
                    children.push(child);
                }
            }
            children.sort_by(|a, b| a.span.start_time.cmp(&b.span.start_time));
            if let Some(parent) = span_map.get_mut(&parent_id) {
                parent.children = children;
            }
        }
    }
    
    let mut root_spans: Vec<TreeSpan> = root_ids
        .into_iter()
        .filter_map(|id| span_map.remove(&id))
        .collect();
    
    for tree_span in span_map.into_values() {
        root_spans.push(tree_span);
    }
    
    root_spans.sort_by(|a, b| a.span.start_time.cmp(&b.span.start_time));
    sort_tree_children(&mut root_spans);
    
    root_spans
}

fn sort_tree_children(nodes: &mut Vec<TreeSpan>) {
    for node in nodes {
        node.children.sort_by(|a, b| a.span.start_time.cmp(&b.span.start_time));
        sort_tree_children(&mut node.children);
    }
}

pub fn format_tree_as_text(nodes: &[TreeSpan]) -> String {
    let mut result = String::new();
    for node in nodes {
        format_tree_node(node, 0, &mut result);
    }
    result
}

fn format_tree_node(node: &TreeSpan, depth: usize, result: &mut String) {
    let indent = "  ".repeat(depth);
    let duration = node.span.end_time
        .map(|end| {
            let dur = end - node.span.start_time;
            format!(" ({:?})", dur)
        })
        .unwrap_or_default();
    
    result.push_str(&format!(
        "{}{} [span_id={}, trace_id={}]{}\n",
        indent,
        node.span.name,
        node.span.span_id,
        node.span.trace_id,
        duration
    ));
    
    if !node.span.tags.is_empty() {
        let tags_str: Vec<String> = node.span.tags
            .iter()
            .map(|(k, v)| {
                let v_str = match v {
                    TagValue::String(s) => format!("\"{}\"", s),
                    TagValue::Number(n) => n.to_string(),
                };
                format!("{}: {}", k, v_str)
            })
            .collect();
        result.push_str(&format!("{}  tags: {{{}}}\n", indent, tags_str.join(", ")));
    }
    
    for child in &node.children {
        format_tree_node(child, depth + 1, result);
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct SearchQuery {
    pub tags: Option<HashMap<String, TagValue>>,
    pub page: Option<usize>,
    pub page_size: Option<usize>,
}

pub async fn search_by_tags(
    storage: Storage,
    query: SearchQuery,
) -> Result<PaginatedResponse<Span>, Rejection> {
    let storage = storage.lock().await;
    
    let search_tags = query.tags.unwrap_or_default();
    
    let mut matching_spans: Vec<Span> = storage
        .values()
        .filter(|span| {
            for (key, value) in &search_tags {
                match span.tags.get(key) {
                    Some(v) if v == value => continue,
                    _ => return false,
                }
            }
            true
        })
        .cloned()
        .collect();
    
    matching_spans.sort_by(|a, b| a.start_time.cmp(&b.start_time));
    
    let total = matching_spans.len();
    let page_size = query.page_size.unwrap_or(DEFAULT_PAGE_SIZE);
    let page = query.page.unwrap_or(1).max(1);
    let start = (page - 1) * page_size;
    let end = (start + page_size).min(total);
    
    let data = if start < total {
        matching_spans[start..end].to_vec()
    } else {
        Vec::new()
    };
    
    Ok(PaginatedResponse {
        data,
        page,
        page_size,
        total,
    })
}

#[derive(Debug, Clone, Deserialize)]
pub struct CleanupRequest {
    pub trace_id: Option<String>,
    pub before: Option<DateTime<Utc>>,
    pub after: Option<DateTime<Utc>>,
}

pub async fn cleanup(
    storage: Storage,
    req: CleanupRequest,
) -> Result<usize, Rejection> {
    let mut storage = storage.lock().await;
    
    let before = req.before;
    let after = req.after;
    let trace_id = req.trace_id;
    
    let keys_to_remove: Vec<String> = storage
        .keys()
        .filter(|span_id| {
            let span = storage.get(*span_id).unwrap();
            
            if let Some(tid) = &trace_id {
                if span.trace_id != *tid {
                    return false;
                }
            }
            
            if let Some(b) = before {
                if span.start_time >= b {
                    return false;
                }
            }
            
            if let Some(a) = after {
                if span.start_time <= a {
                    return false;
                }
            }
            
            true
        })
        .cloned()
        .collect();
    
    let count = keys_to_remove.len();
    for key in keys_to_remove {
        storage.remove(&key);
    }
    
    Ok(count)
}

pub async fn cleanup_old(storage: Storage) -> usize {
    let cutoff = Utc::now() - ChronoDuration::hours(CLEANUP_HOURS);
    
    let mut storage = storage.lock().await;
    
    let keys_to_remove: Vec<String> = storage
        .keys()
        .filter(|span_id| {
            let span = storage.get(*span_id).unwrap();
            span.start_time < cutoff
        })
        .cloned()
        .collect();
    
    let count = keys_to_remove.len();
    for key in keys_to_remove {
        storage.remove(&key);
    }
    
    count
}

fn evict_oldest(storage: &mut HashMap<String, Span>) {
    let evict_count = (MAX_SPANS as f32 * EVICT_PERCENTAGE) as usize;
    
    let mut sorted_spans: Vec<_> = storage
        .iter()
        .map(|(id, span)| (id.clone(), span.start_time))
        .collect();
    
    sorted_spans.sort_by(|a, b| a.1.cmp(&b.1));
    
    for (span_id, _) in sorted_spans.into_iter().take(evict_count) {
        storage.remove(&span_id);
    }
}

fn with_storage(
    storage: Storage,
) -> impl Filter<Extract = (Storage,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || storage.clone())
}

fn json_body<T: for<'de> Deserialize<'de> + Send>(
) -> impl Filter<Extract = (T,), Error = Rejection> + Clone {
    warp::body::content_length_limit(1024 * 16).and(warp::body::json())
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

async fn handle_create_span(
    storage: Storage,
    req: SpanCreateRequest,
) -> Result<impl Reply, Rejection> {
    match create_span(storage, req).await {
        Ok(span) => Ok(warp::reply::json(&span)),
        Err(e) => Ok(warp::reply::json(&ErrorResponse {
            error: format!("{:?}", e),
        })),
    }
}

async fn handle_end_span(
    span_id: String,
    storage: Storage,
    req: SpanEndRequest,
) -> Result<impl Reply, Rejection> {
    match end_span(storage, span_id, req).await {
        Ok(Some(span)) => Ok(warp::reply::json(&span)),
        Ok(None) => Err(warp::reject::not_found()),
        Err(e) => Ok(warp::reply::json(&ErrorResponse {
            error: format!("{:?}", e),
        })),
    }
}

async fn handle_get_trace(
    trace_id: String,
    storage: Storage,
) -> Result<impl Reply, Rejection> {
    match get_trace(storage, trace_id).await {
        Ok(Some(tree)) => {
            let text = format_tree_as_text(&tree);
            Ok(warp::reply::with_header(
                text,
                "Content-Type",
                "text/plain; charset=utf-8",
            ))
        }
        Ok(None) => Err(warp::reject::not_found()),
        Err(e) => Ok(warp::reply::with_header(
            format!("{:?}", e),
            "Content-Type",
            "text/plain",
        )),
    }
}

async fn handle_get_trace_json(
    trace_id: String,
    storage: Storage,
) -> Result<impl Reply, Rejection> {
    match get_trace(storage, trace_id).await {
        Ok(Some(tree)) => Ok(warp::reply::json(&tree)),
        Ok(None) => Err(warp::reject::not_found()),
        Err(e) => Ok(warp::reply::json(&ErrorResponse {
            error: format!("{:?}", e),
        })),
    }
}

async fn handle_search(
    storage: Storage,
    query: SearchQuery,
) -> Result<impl Reply, Rejection> {
    match search_by_tags(storage, query).await {
        Ok(result) => Ok(warp::reply::json(&result)),
        Err(e) => Ok(warp::reply::json(&ErrorResponse {
            error: format!("{:?}", e),
        })),
    }
}

async fn handle_cleanup(
    storage: Storage,
    req: CleanupRequest,
) -> Result<impl Reply, Rejection> {
    match cleanup(storage, req).await {
        Ok(count) => Ok(warp::reply::json(&serde_json::json!({
            "removed": count
        }))),
        Err(e) => Ok(warp::reply::json(&ErrorResponse {
            error: format!("{:?}", e),
        })),
    }
}

#[tokio::main]
async fn main() {
    let storage = new_storage();
    
    let storage_clone = storage.clone();
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(3600));
        loop {
            interval.tick().await;
            let removed = cleanup_old(storage_clone.clone()).await;
            if removed > 0 {
                eprintln!("Cleaned up {} old spans", removed);
            }
        }
    });
    
    let create_span_route = warp::path!("spans")
        .and(warp::post())
        .and(with_storage(storage.clone()))
        .and(json_body())
        .and_then(handle_create_span);
    
    let end_span_route = warp::path!("spans" / String / "end")
        .and(warp::post())
        .and(with_storage(storage.clone()))
        .and(json_body())
        .and_then(handle_end_span);
    
    let get_trace_route = warp::path!("traces" / String)
        .and(warp::get())
        .and(with_storage(storage.clone()))
        .and_then(handle_get_trace);
    
    let get_trace_json_route = warp::path!("traces" / String / "json")
        .and(warp::get())
        .and(with_storage(storage.clone()))
        .and_then(handle_get_trace_json);
    
    let search_route = warp::path!("search")
        .and(warp::post())
        .and(with_storage(storage.clone()))
        .and(json_body())
        .and_then(handle_search);
    
    let cleanup_route = warp::path!("cleanup")
        .and(warp::post())
        .and(with_storage(storage.clone()))
        .and(json_body())
        .and_then(handle_cleanup);
    
    let health_route = warp::path!("health")
        .and(warp::get())
        .map(|| warp::reply::json(&serde_json::json!({"status": "ok"})));
    
    let routes = create_span_route
        .or(end_span_route)
        .or(get_trace_route)
        .or(get_trace_json_route)
        .or(search_route)
        .or(cleanup_route)
        .or(health_route);
    
    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(8080);
    
    println!("Starting trace chain server on port {}", port);
    
    warp::serve(routes)
        .run(([0, 0, 0, 0], port))
        .await;
}
