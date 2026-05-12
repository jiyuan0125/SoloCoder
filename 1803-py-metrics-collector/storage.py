import asyncio
import time
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple
from collections import defaultdict


@dataclass
class TimeSeriesPoint:
    timestamp: float
    value: float


@dataclass
class BucketInfo:
    upper_bound: float
    count: float


@dataclass
class MetricSeries:
    name: str
    labels: Dict[str, str]
    metric_type: str
    value: float = 0.0
    buckets: Dict[float, float] = field(default_factory=dict)
    sum: float = 0.0
    count: float = 0.0
    history: List[TimeSeriesPoint] = field(default_factory=list)
    lock: asyncio.Lock = field(default_factory=asyncio.Lock)


class MetricStorage:
    MAX_LABELS = 10
    MAX_LABEL_VALUE_LEN = 64
    MAX_HISTORY_SIZE = 10000
    
    DEFAULT_HISTOGRAM_BUCKETS = (
        0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0,
        2.5, 5.0, 7.5, 10.0, float('inf')
    )
    
    def __init__(self):
        self._metrics: Dict[str, MetricSeries] = {}
        self._lock = asyncio.Lock()
    
    @staticmethod
    def _make_key(name: str, labels: Dict[str, str]) -> str:
        sorted_labels = sorted(labels.items())
        label_str = ','.join(f'{k}={v}' for k, v in sorted_labels)
        return f'{name}|{label_str}'
    
    @staticmethod
    def validate_labels(labels: Dict[str, str]) -> Optional[str]:
        if len(labels) > MetricStorage.MAX_LABELS:
            return f'Too many labels, maximum {MetricStorage.MAX_LABELS} allowed'
        for key, value in labels.items():
            if len(value) > MetricStorage.MAX_LABEL_VALUE_LEN:
                return f'Label value for {key} exceeds maximum length of {MetricStorage.MAX_LABEL_VALUE_LEN} characters'
        return None
    
    async def update_counter(self, name: str, labels: Dict[str, str], value: float, timestamp: Optional[float] = None):
        key = self._make_key(name, labels)
        ts = timestamp if timestamp is not None else time.time()
        
        async with self._lock:
            if key not in self._metrics:
                self._metrics[key] = MetricSeries(
                    name=name,
                    labels=labels,
                    metric_type='counter'
                )
            series = self._metrics[key]
        
        async with series.lock:
            series.value += value
            series.history.append(TimeSeriesPoint(timestamp=ts, value=series.value))
            if len(series.history) > self.MAX_HISTORY_SIZE:
                series.history = series.history[-self.MAX_HISTORY_SIZE:]
    
    async def update_gauge(self, name: str, labels: Dict[str, str], value: float, timestamp: Optional[float] = None):
        key = self._make_key(name, labels)
        ts = timestamp if timestamp is not None else time.time()
        
        async with self._lock:
            if key not in self._metrics:
                self._metrics[key] = MetricSeries(
                    name=name,
                    labels=labels,
                    metric_type='gauge'
                )
            series = self._metrics[key]
        
        async with series.lock:
            series.value = value
            series.history.append(TimeSeriesPoint(timestamp=ts, value=series.value))
            if len(series.history) > self.MAX_HISTORY_SIZE:
                series.history = series.history[-self.MAX_HISTORY_SIZE:]
    
    async def update_gauge_inc(self, name: str, labels: Dict[str, str], value: float, timestamp: Optional[float] = None):
        key = self._make_key(name, labels)
        ts = timestamp if timestamp is not None else time.time()
        
        async with self._lock:
            if key not in self._metrics:
                self._metrics[key] = MetricSeries(
                    name=name,
                    labels=labels,
                    metric_type='gauge'
                )
            series = self._metrics[key]
        
        async with series.lock:
            series.value += value
            series.history.append(TimeSeriesPoint(timestamp=ts, value=series.value))
            if len(series.history) > self.MAX_HISTORY_SIZE:
                series.history = series.history[-self.MAX_HISTORY_SIZE:]
    
    async def update_histogram(self, name: str, labels: Dict[str, str], value: float, 
                                buckets: Optional[Tuple[float, ...]] = None, 
                                timestamp: Optional[float] = None):
        key = self._make_key(name, labels)
        ts = timestamp if timestamp is not None else time.time()
        bucket_values = buckets if buckets is not None else self.DEFAULT_HISTOGRAM_BUCKETS
        
        async with self._lock:
            if key not in self._metrics:
                self._metrics[key] = MetricSeries(
                    name=name,
                    labels=labels,
                    metric_type='histogram',
                    buckets={b: 0.0 for b in bucket_values}
                )
            series = self._metrics[key]
        
        async with series.lock:
            series.sum += value
            series.count += 1.0
            
            for bucket in series.buckets.keys():
                if value <= bucket:
                    series.buckets[bucket] += 1.0
            
            series.history.append(TimeSeriesPoint(timestamp=ts, value=value))
            if len(series.history) > self.MAX_HISTORY_SIZE:
                series.history = series.history[-self.MAX_HISTORY_SIZE:]
    
    async def get_series(self, name: str, labels: Optional[Dict[str, str]] = None,
                          filters: Optional[Dict[str, str]] = None) -> List[MetricSeries]:
        results = []
        filter_dict = labels if labels is not None else (filters if filters is not None else {})
        
        async with self._lock:
            for series in self._metrics.values():
                if series.name != name:
                    continue
                
                match = True
                for k, v in filter_dict.items():
                    if series.labels.get(k) != v:
                        match = False
                        break
                
                if match:
                    results.append(series)
        
        return results
    
    async def get_all_series(self) -> List[MetricSeries]:
        async with self._lock:
            return list(self._metrics.values())
    
    async def query(self, name: str, label_filters: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
        series_list = await self.get_series(name, filters=label_filters)
        
        result = {
            'name': name,
            'metric_type': None,
            'series': []
        }
        
        for series in series_list:
            series_data = {
                'labels': series.labels,
                'value': series.value,
                'history': [{'timestamp': p.timestamp, 'value': p.value} for p in series.history]
            }
            
            if series.metric_type == 'histogram':
                series_data['buckets'] = series.buckets
                series_data['sum'] = series.sum
                series_data['count'] = series.count
                
                if series.count > 0 and len(series.history) > 0:
                    values = sorted([p.value for p in series.history])
                    series_data['p50'] = self._percentile(values, 0.50)
                    series_data['p90'] = self._percentile(values, 0.90)
                    series_data['p99'] = self._percentile(values, 0.99)
            
            result['series'].append(series_data)
            result['metric_type'] = series.metric_type
        
        return result
    
    @staticmethod
    def _percentile(sorted_values: List[float], p: float) -> float:
        if not sorted_values:
            return 0.0
        k = (len(sorted_values) - 1) * p
        f = int(k)
        c = f + 1
        if f >= len(sorted_values) - 1:
            return sorted_values[-1]
        if f == c:
            return sorted_values[f]
        return sorted_values[f] + (sorted_values[c] - sorted_values[f]) * (k - f)


storage = MetricStorage()
