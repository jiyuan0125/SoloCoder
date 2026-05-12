import os
import time
from enum import Enum
from typing import Any, Dict, List, Optional
from fastapi import FastAPI, HTTPException, Query, Request
from fastapi.responses import PlainTextResponse
from pydantic import BaseModel, Field, field_validator

from storage import storage, MetricStorage


class MetricType(str, Enum):
    COUNTER = "counter"
    GAUGE = "gauge"
    HISTOGRAM = "histogram"


class GaugeOperation(str, Enum):
    SET = "set"
    INCREMENT = "inc"


class MetricDataPoint(BaseModel):
    name: str = Field(..., description="指标名称")
    metric_type: MetricType = Field(..., description="指标类型: counter, gauge, histogram")
    labels: Dict[str, str] = Field(default_factory=dict, description="标签键值对")
    value: float = Field(..., description="指标值")
    timestamp: Optional[float] = Field(None, description="时间戳(秒), 为空则使用服务端时间")
    operation: Optional[GaugeOperation] = Field(None, description="Gauge 操作类型: set 或 inc")
    buckets: Optional[List[float]] = Field(None, description="Histogram 自定义桶边界")

    @field_validator('labels')
    @classmethod
    def validate_labels(cls, v: Dict[str, str]) -> Dict[str, str]:
        error = MetricStorage.validate_labels(v)
        if error:
            raise ValueError(error)
        return v


class MetricsBatchRequest(BaseModel):
    metrics: List[MetricDataPoint] = Field(..., max_length=100, description="批量上报的指标数据点, 最多100个")


app = FastAPI(title="Metrics Collector Service", version="1.0.0")


@app.get("/", response_model=Dict[str, str])
async def root():
    return {"status": "ok", "service": "metrics-collector"}


@app.post("/metrics/ingest", response_model=Dict[str, Any])
async def ingest_metrics(request: MetricsBatchRequest):
    if len(request.metrics) > 100:
        raise HTTPException(status_code=400, detail="Maximum 100 metrics per batch allowed")
    
    processed = 0
    errors: List[Dict[str, Any]] = []
    
    for i, point in enumerate(request.metrics):
        try:
            labels_error = MetricStorage.validate_labels(point.labels)
            if labels_error:
                raise ValueError(labels_error)
            
            ts = point.timestamp if point.timestamp is not None else time.time()
            
            if point.metric_type == MetricType.COUNTER:
                if point.value < 0:
                    raise ValueError("Counter value must be non-negative")
                await storage.update_counter(point.name, point.labels, point.value, ts)
            
            elif point.metric_type == MetricType.GAUGE:
                op = point.operation if point.operation is not None else GaugeOperation.SET
                if op == GaugeOperation.SET:
                    await storage.update_gauge(point.name, point.labels, point.value, ts)
                else:
                    await storage.update_gauge_inc(point.name, point.labels, point.value, ts)
            
            elif point.metric_type == MetricType.HISTOGRAM:
                buckets = tuple(point.buckets) if point.buckets else None
                await storage.update_histogram(point.name, point.labels, point.value, buckets, ts)
            
            processed += 1
        
        except Exception as e:
            errors.append({
                "index": i,
                "name": point.name,
                "error": str(e)
            })
    
    if errors and processed == 0:
        raise HTTPException(
            status_code=400,
            detail={"message": "All metrics failed to process", "errors": errors}
        )
    
    return {
        "status": "success",
        "processed": processed,
        "total": len(request.metrics),
        "errors": errors
    }


def _format_prometheus_labels(labels: Dict[str, str]) -> str:
    if not labels:
        return ""
    label_parts = []
    for k, v in sorted(labels.items()):
        escaped_value = v.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n')
        label_parts.append(f'{k}="{escaped_value}"')
    return "{" + ",".join(label_parts) + "}"


def _format_bucket_label(le: float) -> str:
    if le == float('inf'):
        return 'le="+Inf"'
    return f'le="{le}"'


@app.get("/metrics", response_class=PlainTextResponse)
async def prometheus_metrics():
    series_list = await storage.get_all_series()
    
    output_lines: List[str] = []
    metric_names_seen: Dict[str, str] = {}
    
    for series in series_list:
        if series.name not in metric_names_seen:
            metric_names_seen[series.name] = series.metric_type
            output_lines.append(f'# HELP {series.name} {series.metric_type} metric')
            output_lines.append(f'# TYPE {series.name} {series.metric_type}')
        
        label_str = _format_prometheus_labels(series.labels)
        
        if series.metric_type == 'counter':
            output_lines.append(f'{series.name}{label_str} {series.value}')
        
        elif series.metric_type == 'gauge':
            output_lines.append(f'{series.name}{label_str} {series.value}')
        
        elif series.metric_type == 'histogram':
            bucket_labels_base = series.labels.copy()
            for bucket_le, bucket_count in sorted(series.buckets.items()):
                bucket_labels = bucket_labels_base.copy()
                bucket_labels['le'] = '+Inf' if bucket_le == float('inf') else str(bucket_le)
                bucket_label_str = _format_prometheus_labels(bucket_labels)
                output_lines.append(f'{series.name}_bucket{bucket_label_str} {bucket_count}')
            
            sum_label_str = _format_prometheus_labels(series.labels)
            output_lines.append(f'{series.name}_sum{sum_label_str} {series.sum}')
            output_lines.append(f'{series.name}_count{sum_label_str} {series.count}')
    
    output_lines.append('')
    return PlainTextResponse(content='\n'.join(output_lines), media_type="text/plain")


@app.get("/api/metrics/{name}", response_model=Dict[str, Any])
async def query_metric(
    name: str,
    request: Request
):
    label_filters: Dict[str, str] = {}
    
    for key, value in request.query_params.items():
        if key not in ['name']:
            label_filters[key] = value
    
    result = await storage.query(name, label_filters)
    
    if not result['series']:
        raise HTTPException(
            status_code=404,
            detail=f"Metric '{name}' not found with the specified labels"
        )
    
    return result


@app.get("/api/metrics", response_model=Dict[str, Any])
async def list_metrics():
    series_list = await storage.get_all_series()
    
    metrics_by_name: Dict[str, Dict[str, Any]] = {}
    
    for series in series_list:
        if series.name not in metrics_by_name:
            metrics_by_name[series.name] = {
                "name": series.name,
                "type": series.metric_type,
                "label_keys": set(),
                "series_count": 0
            }
        
        metrics_by_name[series.name]["label_keys"].update(series.labels.keys())
        metrics_by_name[series.name]["series_count"] += 1
    
    for name, info in metrics_by_name.items():
        info["label_keys"] = sorted(info["label_keys"])
    
    return {
        "metrics": list(metrics_by_name.values()),
        "total_metrics": len(metrics_by_name)
    }


@app.get("/health", response_model=Dict[str, str])
async def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=False
    )
