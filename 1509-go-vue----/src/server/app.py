from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import PlainTextResponse
from typing import List, Optional

from core import (
    InMemoryStorage,
    SupplyChainService,
    InspectionService,
    RecallService,
    ExportService,
    SupplyChainRecord,
    InspectionRecord,
    Recall,
    TodoItem
)


storage = InMemoryStorage()
supply_chain_service = SupplyChainService(storage)
inspection_service = InspectionService(storage)
recall_service = RecallService(storage)
export_service = ExportService(storage)


def create_app() -> FastAPI:
    app = FastAPI(
        title="食品安全溯源管理系统",
        description="食品供应链溯源、质量检测和召回管理后端服务",
        version="1.0.0"
    )

    @app.get("/", tags=["系统"])
    async def root():
        return {"name": "食品安全溯源管理系统", "version": "1.0.0", "status": "running"}

    @app.post("/api/supply-chain", response_model=SupplyChainRecord, tags=["供应链"])
    async def add_supply_chain_record(record: SupplyChainRecord):
        try:
            return supply_chain_service.add_record(record)
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/api/supply-chain/{batch_number}", response_model=List[SupplyChainRecord], tags=["供应链"])
    async def get_supply_chain_timeline(batch_number: str):
        return supply_chain_service.get_timeline(batch_number)

    @app.get("/api/supply-chain", response_model=List[SupplyChainRecord], tags=["供应链"])
    async def list_supply_chain():
        return supply_chain_service.get_all()

    @app.post("/api/inspections", response_model=InspectionRecord, tags=["检测管理"])
    async def add_inspection_record(record: InspectionRecord):
        try:
            return inspection_service.add_record(record)
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/api/inspections/{batch_number}", response_model=List[InspectionRecord], tags=["检测管理"])
    async def get_inspections_by_batch(batch_number: str):
        return inspection_service.get_by_batch(batch_number)

    @app.get("/api/inspections", response_model=List[InspectionRecord], tags=["检测管理"])
    async def list_inspections():
        return inspection_service.get_all()

    @app.post("/api/recalls", response_model=Recall, tags=["召回管理"])
    async def create_recall(batch_number: str, reason: str):
        try:
            return recall_service.create_recall(batch_number, reason)
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/api/recalls/{recall_id}", response_model=Recall, tags=["召回管理"])
    async def get_recall(recall_id: str):
        recall = recall_service.get_recall(recall_id)
        if recall is None:
            raise HTTPException(status_code=404, detail="召回记录不存在")
        return recall

    @app.get("/api/recalls", response_model=List[Recall], tags=["召回管理"])
    async def list_recalls(batch_number: Optional[str] = Query(None)):
        if batch_number:
            return recall_service.get_by_batch(batch_number)
        return recall_service.get_all()

    @app.post("/api/recalls/{recall_id}/complete", response_model=Recall, tags=["召回管理"])
    async def complete_recall(recall_id: str, completed_by: str):
        try:
            recall = recall_service.complete_recall(recall_id, completed_by)
            if recall is None:
                raise HTTPException(status_code=404, detail="召回记录不存在")
            return recall
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.post("/api/recalls/{recall_id}/cancel", response_model=Recall, tags=["召回管理"])
    async def cancel_recall(recall_id: str, cancelled_by: str):
        recall = recall_service.cancel_recall(recall_id, cancelled_by)
        if recall is None:
            raise HTTPException(status_code=404, detail="召回记录不存在")
        return recall

    @app.get("/api/recalls/{recall_id}/todos", response_model=List[TodoItem], tags=["召回管理"])
    async def get_recall_todos(recall_id: str):
        return recall_service.get_todos(recall_id)

    @app.get("/api/todos", response_model=List[TodoItem], tags=["召回管理"])
    async def list_todos():
        return recall_service.get_all_todos()

    @app.post("/api/todos/{todo_id}/confirm", response_model=TodoItem, tags=["召回管理"])
    async def confirm_todo(todo_id: str, confirmed_by: str, remarks: Optional[str] = None):
        todo = recall_service.confirm_todo(todo_id, confirmed_by, remarks)
        if todo is None:
            raise HTTPException(status_code=404, detail="待办事项不存在")
        return todo

    @app.post("/api/todos/update-overdue", tags=["召回管理"])
    async def update_overdue_todos():
        count = recall_service.update_overdue_todos()
        return {"updated": count}

    @app.get("/api/export/{batch_number}", response_class=PlainTextResponse, tags=["数据导出"])
    async def export_batch_traceability(batch_number: str):
        return export_service.export_batch_traceability(batch_number)

    return app
