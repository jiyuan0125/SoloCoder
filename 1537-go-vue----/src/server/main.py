from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

from core import settings, init_db
from .routers import industries, companies, emission_sources, emission_data, quota_transactions, reports, analytics

def create_app() -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        init_db()
        yield
    
    app = FastAPI(
        title=settings.APP_NAME,
        version=settings.VERSION,
        lifespan=lifespan
    )
    
    app.include_router(industries.router, prefix="/api/v1/industries", tags=["行业管理"])
    app.include_router(companies.router, prefix="/api/v1/companies", tags=["企业管理"])
    app.include_router(emission_sources.router, prefix="/api/v1/emission-sources", tags=["排放源管理"])
    app.include_router(emission_data.router, prefix="/api/v1/emission-data", tags=["排放数据管理"])
    app.include_router(quota_transactions.router, prefix="/api/v1/quota-transactions", tags=["配额交易管理"])
    app.include_router(reports.router, prefix="/api/v1/reports", tags=["碳排放报告管理"])
    app.include_router(analytics.router, prefix="/api/v1/analytics", tags=["数据分析"])
    
    @app.get("/health")
    async def health_check():
        return {"status": "healthy", "version": settings.VERSION}
    
    @app.exception_handler(HTTPException)
    async def http_exception_handler(request, exc):
        return JSONResponse(
            status_code=exc.status_code,
            content={"detail": exc.detail}
        )
    
    return app

app = create_app()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "server.main:app",
        host=settings.SERVER_HOST,
        port=settings.SERVER_PORT,
        reload=False
    )
