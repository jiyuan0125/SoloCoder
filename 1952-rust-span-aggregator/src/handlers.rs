use actix_web::{web, HttpResponse, Responder};
use std::collections::HashSet;
use crate::models::{Span, SpanWithSlow, TraceListItem, TraceQuery, TraceResponse};
use crate::storage::AppState;

pub async fn submit_span(
    data: web::Data<AppState>,
    span: web::Json<Span>,
) -> impl Responder {
    let span = span.into_inner();
    data.add_span(span);
    HttpResponse::Ok().json(serde_json::json!({"status": "accepted"}))
}

pub async fn get_trace(
    data: web::Data<AppState>,
    trace_id: web::Path<String>,
) -> impl Responder {
    match data.get_trace(&trace_id) {
        Some(trace) => {
            let spans: Vec<SpanWithSlow> = trace
                .spans
                .into_iter()
                .map(|s| {
                    let slow = s.duration > data.p99_threshold_ms;
                    SpanWithSlow { span: s, slow }
                })
                .collect();

            let total_duration = spans.iter().map(|s| s.span.duration).max().unwrap_or(0);

            let response = TraceResponse {
                trace_id: trace_id.into_inner(),
                spans,
                total_duration,
            };
            HttpResponse::Ok().json(response)
        }
        None => HttpResponse::NotFound().json(serde_json::json!({"error": "trace not found"})),
    }
}

pub async fn list_traces(
    data: web::Data<AppState>,
    query: web::Query<TraceQuery>,
) -> impl Responder {
    let q = query.into_inner();
    let traces = data.list_traces(
        q.service_name.as_deref(),
        q.min_duration,
        q.start_time,
        q.end_time,
    );

    let items: Vec<TraceListItem> = traces
        .into_iter()
        .map(|trace| {
            let services: HashSet<String> = trace
                .spans
                .iter()
                .map(|s| s.service_name.clone())
                .collect();
            
            let has_slow_spans = trace
                .spans
                .iter()
                .any(|s| s.duration > data.p99_threshold_ms);

            TraceListItem {
                trace_id: trace.trace_id,
                services: services.into_iter().collect(),
                total_duration: trace.total_duration,
                has_slow_spans,
                start_time: trace.earliest_start,
            }
        })
        .collect();

    HttpResponse::Ok().json(items)
}

pub async fn get_dependency_graph(data: web::Data<AppState>) -> impl Responder {
    let graph = data.get_dependency_graph();
    HttpResponse::Ok().json(graph)
}

pub async fn get_stats(data: web::Data<AppState>) -> impl Responder {
    let stats = data.get_stats();
    HttpResponse::Ok().json(stats)
}
