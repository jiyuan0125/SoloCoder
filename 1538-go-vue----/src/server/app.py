from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.core.database import init_db


def create_app() -> FastAPI:
    init_db()

    app = FastAPI(
        title="固废管理系统",
        description="固体废物全流程管理系统 API",
        version="1.0.0"
    )

    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    from src.server.routers import (
        units, qualifications, ledgers, waybills, monthly_ledgers, alerts, exports
    )

    app.include_router(units.router, prefix="/api/units", tags=["单位管理"])
    app.include_router(qualifications.router, prefix="/api/qualifications", tags=["资质管理"])
    app.include_router(ledgers.router, prefix="/api/ledgers", tags=["固废台账"])
    app.include_router(waybills.router, prefix="/api/waybills", tags=["电子联单"])
    app.include_router(monthly_ledgers.router, prefix="/api/monthly-ledgers", tags=["月度台账"])
    app.include_router(alerts.router, prefix="/api/alerts", tags=["预警管理"])
    app.include_router(exports.router, prefix="/api/exports", tags=["数据导出"])

    return app


app = create_app()
