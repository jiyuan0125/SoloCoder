from typing import Dict, List, Optional
from pydantic import BaseModel, Field


class HeaderMatch(BaseModel):
    name: str
    value: str
    exact: bool = True


class CookieMatch(BaseModel):
    name: str
    value: str
    exact: bool = True


class IPRange(BaseModel):
    cidr: str


class ColoringRule(BaseModel):
    id: str
    tag: str
    priority: int = 0
    headers: Optional[List[HeaderMatch]] = None
    cookies: Optional[List[CookieMatch]] = None
    ip_ranges: Optional[List[IPRange]] = None


class RouteRule(BaseModel):
    tag: str
    upstream: str
    weight: int = 100


class ServiceRoutes(BaseModel):
    service_name: str
    rules: Dict[str, RouteRule]


class Stats(BaseModel):
    total_requests: int = 0
    error_requests: int = 0
    total_latency_ms: float = 0.0

    @property
    def error_rate(self) -> float:
        if self.total_requests == 0:
            return 0.0
        return self.error_requests / self.total_requests

    @property
    def avg_latency_ms(self) -> float:
        if self.total_requests == 0:
            return 0.0
        return self.total_latency_ms / self.total_requests


class TagStats(BaseModel):
    tag: str
    stats: Stats


class ServiceStats(BaseModel):
    service_name: str
    tags: List[TagStats]


class MirrorConfig(BaseModel):
    enabled: bool = False
    tags: List[str] = Field(default_factory=lambda: ["gray", "canary"])
    upstream: Optional[str] = None


class GatewayConfig(BaseModel):
    mirror: MirrorConfig = Field(default_factory=MirrorConfig)
