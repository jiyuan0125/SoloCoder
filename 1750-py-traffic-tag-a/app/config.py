import os
from typing import List, Optional

from app.models import (
    ColoringRule,
    GatewayConfig,
    HeaderMatch,
    IPRange,
    MirrorConfig,
    RouteRule,
    ServiceRoutes,
)


class Settings:
    def __init__(self):
        self.port: int = int(os.getenv("PORT", "8000"))
        self.default_coloring_rules: List[ColoringRule] = [
            ColoringRule(
                id="default-1",
                tag="default",
                priority=0,
            )
        ]
        self.default_service_routes: dict = {
            "sample": ServiceRoutes(
                service_name="sample",
                rules={
                    "default": RouteRule(tag="default", upstream="http://v1:8080"),
                    "gray": RouteRule(tag="gray", upstream="http://v2:8080"),
                },
            )
        }
        self.gateway_config: GatewayConfig = GatewayConfig(
            mirror=MirrorConfig(
                enabled=False,
                tags=["gray", "canary"],
                upstream=None,
            )
        )


settings = Settings()
