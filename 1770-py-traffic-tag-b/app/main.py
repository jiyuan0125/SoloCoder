from fastapi import FastAPI
from app.config import settings
from app.middleware import TrafficRoutingMiddleware
from app.routers import api, admin
from app.rules_engine import rule_engine
from app.models import GrayRule, GrayStrategy


def create_app() -> FastAPI:
    app = FastAPI(
        title="Gray Release API",
        description="灰度发布系统 - 支持流量分流、动态规则配置、平滑过渡",
        version="1.0.0"
    )

    app.add_middleware(TrafficRoutingMiddleware)

    app.include_router(api.router)
    app.include_router(admin.router)

    _init_default_rules()

    @app.on_event("startup")
    async def startup_event():
        print(f"Gray Release API starting on port {settings.port}")
        print(f"Gray release enabled: {settings.gray_enable}")

    @app.on_event("shutdown")
    async def shutdown_event():
        print("Gray Release API shutting down...")

    return app


def _init_default_rules():
    default_weight_rule = GrayRule(
        rule_id="default_weight",
        strategy=GrayStrategy.WEIGHT,
        enabled=True,
        version="2.0",
        priority=10,
        weight=10.0
    )
    rule_engine.add_or_update_rule(default_weight_rule)
    
    default_gray_user_rule = GrayRule(
        rule_id="gray_users",
        strategy=GrayStrategy.USER_ID,
        enabled=True,
        version="2.0",
        priority=20,
        user_ids=["user_001", "user_002", "tester_01", "tester_02"]
    )
    rule_engine.add_or_update_rule(default_gray_user_rule)

    default_header_rule = GrayRule(
        rule_id="beta_testers",
        strategy=GrayStrategy.HEADER,
        enabled=True,
        version="2.0",
        priority=15,
        header_name="X-Beta-Tester",
        header_values=["true", "yes", "1"]
    )
    rule_engine.add_or_update_rule(default_header_rule)


app = create_app()
