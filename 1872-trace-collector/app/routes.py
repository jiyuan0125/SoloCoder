from typing import List, Dict, Any, Optional, Union
from datetime import datetime, timedelta
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import and_

from .database import get_db
from .models import SpanModel, RedactionRuleModel
from .schemas import (
    Span, SpanCreate, TraceTree, SpanTreeNode, 
    SlowSpan, RedactionRule, RedactionType
)
from .redaction import get_redaction_rules, redact_tags
from .config import settings

router = APIRouter()


def model_to_schema(span_model: SpanModel) -> Span:
    return Span(
        trace_id=span_model.trace_id,
        span_id=span_model.span_id,
        operation=span_model.operation,
        start_time=span_model.start_time,
        end_time=span_model.end_time,
        duration_ms=span_model.duration_ms,
        tags=span_model.tags,
        parent_span_id=span_model.parent_span_id,
    )


@router.post("/spans", status_code=201)
def create_spans(
    spans: Union[SpanCreate, List[SpanCreate]],
    db: Session = Depends(get_db)
):
    if isinstance(spans, SpanCreate):
        spans = [spans]
    
    rules = get_redaction_rules(db)
    now = datetime.utcnow()
    
    for span in spans:
        duration_ms = (span.end_time - span.start_time).total_seconds() * 1000
        redacted_tags = redact_tags(span.tags, rules)
        
        span_model = SpanModel(
            trace_id=span.trace_id,
            span_id=span.span_id,
            operation=span.operation,
            start_time=span.start_time,
            end_time=span.end_time,
            duration_ms=duration_ms,
            tags=redacted_tags,
            parent_span_id=span.parent_span_id,
            created_at=now,
        )
        db.merge(span_model)
    
    db.commit()
    return {"status": "ok", "count": len(spans)}


@router.get("/traces/{trace_id}", response_model=TraceTree)
def get_trace(trace_id: str, db: Session = Depends(get_db)):
    span_models = (
        db.query(SpanModel)
        .filter(SpanModel.trace_id == trace_id)
        .all()
    )
    
    if not span_models:
        raise HTTPException(status_code=404, detail=f"Trace {trace_id} not found")
    
    spans = {s.span_id: model_to_schema(s) for s in span_models}
    span_ids = set(spans.keys())
    
    children_map: Dict[str, List[str]] = {}
    
    for span_model in span_models:
        if span_model.parent_span_id:
            if span_model.parent_span_id not in children_map:
                children_map[span_model.parent_span_id] = []
            children_map[span_model.parent_span_id].append(span_model.span_id)
    
    root_spans: List[SpanTreeNode] = []
    orphan_spans: List[Span] = []
    
    for span_id, span in spans.items():
        if not span.parent_span_id:
            node = SpanTreeNode(span=span, children=[], is_orphan=False)
            root_spans.append(node)
        elif span.parent_span_id not in span_ids:
            orphan_spans.append(span)
        else:
            pass
    
    def build_node(span_id: str) -> SpanTreeNode:
        span = spans[span_id]
        children = []
        if span_id in children_map:
            for child_id in children_map[span_id]:
                if child_id in spans:
                    children.append(build_node(child_id))
        
        return SpanTreeNode(span=span, children=children, is_orphan=False)
    
    for root in root_spans:
        if root.span.span_id in children_map:
            children_list = []
            for child_id in children_map[root.span.span_id]:
                if child_id in spans:
                    children_list.append(build_node(child_id))
            root.children = children_list
    
    root_spans.sort(key=lambda x: x.span.start_time)
    
    return TraceTree(
        trace_id=trace_id,
        root_spans=root_spans,
        orphan_spans=orphan_spans,
    )


@router.get("/spans/slow", response_model=List[SlowSpan])
def get_slow_spans(
    service: str = Query(..., description="Service name (from tags.service)"),
    threshold: float = Query(..., gt=0, description="Duration threshold in milliseconds"),
    start: datetime = Query(..., description="Start time"),
    end: datetime = Query(..., description="End time"),
    db: Session = Depends(get_db)
):
    span_models = (
        db.query(SpanModel)
        .filter(
            and_(
                SpanModel.tags.op("->>")("service") == service,
                SpanModel.duration_ms >= threshold,
                SpanModel.start_time >= start,
                SpanModel.start_time <= end,
            )
        )
        .order_by(SpanModel.duration_ms.desc())
        .all()
    )
    
    result = []
    for span in span_models:
        result.append(SlowSpan(
            trace_id=span.trace_id,
            span_id=span.span_id,
            operation=span.operation,
            service=span.tags.get("service") if span.tags else None,
            start_time=span.start_time,
            end_time=span.end_time,
            duration_ms=span.duration_ms,
        ))
    
    return result


@router.post("/redaction/rules", status_code=201)
def create_redaction_rule(
    rule: RedactionRule,
    db: Session = Depends(get_db)
):
    if rule.redaction_type == RedactionType.replace and rule.replacement is None:
        raise HTTPException(
            status_code=400, 
            detail="replacement is required for 'replace' type"
        )
    
    existing = (
        db.query(RedactionRuleModel)
        .filter(RedactionRuleModel.tag_key == rule.tag_key)
        .first()
    )
    
    if existing:
        existing.redaction_type = rule.redaction_type.value
        existing.replacement = rule.replacement
    else:
        rule_model = RedactionRuleModel(
            tag_key=rule.tag_key,
            redaction_type=rule.redaction_type.value,
            replacement=rule.replacement,
        )
        db.add(rule_model)
    
    db.commit()
    return {"status": "ok", "rule": rule}


@router.get("/redaction/rules", response_model=List[RedactionRule])
def list_redaction_rules(db: Session = Depends(get_db)):
    rules = db.query(RedactionRuleModel).all()
    return [
        RedactionRule(
            tag_key=r.tag_key,
            redaction_type=RedactionType(r.redaction_type),
            replacement=r.replacement,
        )
        for r in rules
    ]


@router.get("/health")
def health_check():
    return {"status": "ok"}
