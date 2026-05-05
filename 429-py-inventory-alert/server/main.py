from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from server.routers import alerts, inventory, products, replenishment, reports

app = FastAPI(
    title="Inventory Alert System",
    description="库存预警系统 - 避免断货和积压的智能库存监控",
    version="0.1.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(products.router)
app.include_router(inventory.router)
app.include_router(alerts.router)
app.include_router(replenishment.router)
app.include_router(reports.router)


@app.get("/")
async def root() -> dict[str, str]:
    return {"message": "Inventory Alert System API", "version": "0.1.0"}


@app.get("/health")
async def health_check() -> dict[str, str]:
    return {"status": "healthy"}
