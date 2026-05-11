import os
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import PlainTextResponse
from typing import List, Optional

from core import (
    InMemoryRepository,
    TraceabilityService,
    SupplyChainCreate,
    SupplyChainRecord,
    InspectionCreate,
    InspectionRecord,
    RecallCreate,
    RecallRecord,
    TodoItem,
    TodoUpdate,
    ValidationError,
)

app = FastAPI(
    title="食品安全溯源管理系统",
    description="供应链管理、检测管理、召回管理和数据导出",
    version="1.0.0",
)

repository = InMemoryRepository()
service = TraceabilityService(repository)


@app.get("/", tags=["Health"])
async def health_check():
    return {"status": "ok", "service": "食品安全溯源管理系统"}


@app.post("/api/supply-chain", tags=["供应链管理"], response_model=SupplyChainRecord)
def create_supply_chain_record(record: SupplyChainCreate):
    try:
        return service.create_supply_chain_record(record)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/supply-chain/{batch_number}", tags=["供应链管理"], response_model=List[SupplyChainRecord])
def get_supply_chain_timeline(batch_number: str):
    return service.get_supply_chain_timeline(batch_number)


@app.post("/api/inspection", tags=["检测管理"], response_model=InspectionRecord)
def create_inspection_record(record: InspectionCreate):
    try:
        return service.create_inspection_record(record)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/inspection/{batch_number}", tags=["检测管理"], response_model=List[InspectionRecord])
def get_inspection_records(batch_number: str):
    return service.get_inspection_records(batch_number)


@app.post("/api/recall", tags=["召回管理"])
def initiate_recall(data: RecallCreate):
    try:
        result = service.initiate_recall(data)
        return {
            "recall": result["recall"],
            "todos": result["todos"],
        }
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/recall", tags=["召回管理"], response_model=List[RecallRecord])
def list_recalls():
    return service.get_all_recalls()


@app.get("/api/recall/{recall_id}", tags=["召回管理"], response_model=Optional[RecallRecord])
def get_recall(recall_id: str):
    recall = service.get_recall(recall_id)
    if not recall:
        raise HTTPException(status_code=404, detail="召回记录不存在")
    return recall


@app.get("/api/recall/{recall_id}/todos", tags=["召回管理"], response_model=List[TodoItem])
def get_recall_todos(recall_id: str):
    return service.get_recall_todos(recall_id)


@app.post("/api/recall/{recall_id}/complete", tags=["召回管理"], response_model=RecallRecord)
def complete_recall(recall_id: str):
    try:
        result = service.complete_recall(recall_id)
        if not result:
            raise HTTPException(status_code=404, detail="召回记录不存在")
        return result
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/todos", tags=["待办管理"], response_model=List[TodoItem])
def list_todos():
    return service.get_all_todos()


@app.put("/api/todos/{todo_id}", tags=["待办管理"], response_model=TodoItem)
def update_todo(todo_id: str, data: TodoUpdate):
    result = service.update_todo_status(todo_id, data)
    if not result:
        raise HTTPException(status_code=404, detail="待办任务不存在")
    return result


@app.get("/api/export/{batch_number}", tags=["数据导出"])
def export_traceability(batch_number: str):
    return service.export_traceability(batch_number)


@app.get("/api/export/{batch_number}/text", tags=["数据导出"], response_class=PlainTextResponse)
def export_to_text(batch_number: str):
    return service.export_to_text(batch_number)
