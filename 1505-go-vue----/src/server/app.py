from fastapi import FastAPI

from server.routers import batches, exports, inspections, materials, recipes, todos

app = FastAPI(
    title="食品加工厂生产管理系统",
    description="食品加工厂生产管理后端服务",
    version="0.1.0",
)

app.include_router(recipes.router)
app.include_router(batches.router)
app.include_router(inspections.router)
app.include_router(todos.router)
app.include_router(materials.router)
app.include_router(exports.router)


@app.get("/")
def root():
    return {"message": "食品加工厂生产管理系统 API", "version": "0.1.0"}


@app.get("/health")
def health():
    return {"status": "ok"}
