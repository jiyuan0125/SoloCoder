import os
from typing import List, Optional
from pydantic import BaseModel
from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from src.core.models import SolutionType
from src.core.schemas import (
    AuditCreate,
    AuditResponse,
    SolutionCreate,
    SolutionResponse,
    DimensionCreate,
    SolutionFilterResponse,
    SolutionImplementRequest,
    AcceptanceRequest,
    AuditLogResponse,
    AuditSummary,
)
from src.core.services import AuditService, AuditLogService
from src.server.database import init_db, get_db

app = FastAPI(
    title="清洁生产审核管理系统",
    description="企业清洁生产审核流程管理后端服务",
    version="1.0.0",
)


@app.on_event("startup")
def startup_event():
    init_db()


@app.get("/health")
def health_check():
    return {"status": "ok"}


@app.post("/audits", response_model=AuditResponse, status_code=201)
def create_audit(
    data: AuditCreate,
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None, description="操作用户ID"),
    user_name: Optional[str] = Query(None, description="操作用户名称"),
):
    service = AuditService(db)
    audit = service.create_audit(
        company_name=data.company_name,
        company_id=data.company_id,
        start_date=data.start_date,
        user_id=user_id,
        user_name=user_name,
    )
    return audit


@app.get("/audits", response_model=List[AuditResponse])
def list_audits(db: Session = Depends(get_db)):
    service = AuditService(db)
    return service.get_all_audits()


@app.get("/audits/{audit_id}", response_model=AuditResponse)
def get_audit(audit_id: int, db: Session = Depends(get_db)):
    service = AuditService(db)
    audit = service.get_audit(audit_id)
    if not audit:
        raise HTTPException(status_code=404, detail="审核项目不存在")
    return audit


class StageAdvanceRequest(BaseModel):
    notes: Optional[str] = None


@app.post("/audits/{audit_id}/advance", response_model=dict)
def advance_stage(
    audit_id: int,
    request: Optional[StageAdvanceRequest] = None,
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None),
    user_name: Optional[str] = Query(None),
):
    notes = request.notes if request else None
    service = AuditService(db)
    stage = service.advance_stage(audit_id, notes, user_id, user_name)
    if not stage:
        raise HTTPException(status_code=400, detail="无法推进阶段，可能所有阶段已完成")
    return {
        "stage_id": stage.id,
        "stage_type": stage.stage_type.value,
        "completed": True,
    }


@app.post("/audits/{audit_id}/solutions", response_model=SolutionResponse, status_code=201)
def create_solution(
    audit_id: int,
    data: SolutionCreate,
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None),
    user_name: Optional[str] = Query(None),
):
    service = AuditService(db)
    solution = service.create_solution(
        audit_id=audit_id,
        name=data.name,
        solution_type=data.solution_type,
        description=data.description,
        expected_energy_saving=data.expected_energy_saving,
        expected_investment=data.expected_investment,
        user_id=user_id,
        user_name=user_name,
    )
    if not solution:
        raise HTTPException(status_code=404, detail="审核项目不存在")
    return solution


@app.get("/solutions/{solution_id}", response_model=SolutionResponse)
def get_solution(solution_id: int, db: Session = Depends(get_db)):
    from src.core.repositories import SolutionRepository
    repo = SolutionRepository(db)
    solution = repo.get_by_id(solution_id)
    if not solution:
        raise HTTPException(status_code=404, detail="方案不存在")
    return solution


@app.post("/solutions/{solution_id}/dimensions", response_model=SolutionResponse)
def add_dimensions(
    solution_id: int,
    dimensions: List[DimensionCreate],
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None),
    user_name: Optional[str] = Query(None),
):
    service = AuditService(db)
    solution = service.add_dimension_scores(
        solution_id=solution_id,
        dimensions=[d.model_dump() for d in dimensions],
        user_id=user_id,
        user_name=user_name,
    )
    if not solution:
        raise HTTPException(status_code=404, detail="方案不存在")
    return solution


@app.post("/solutions/{solution_id}/filter", response_model=SolutionFilterResponse)
def filter_solution(
    solution_id: int,
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None),
    user_name: Optional[str] = Query(None),
):
    service = AuditService(db)
    result = service.filter_solution(solution_id, user_id, user_name)
    if not result:
        raise HTTPException(status_code=404, detail="方案不存在")
    return result


@app.post("/solutions/{solution_id}/implement", response_model=SolutionResponse)
def implement_solution(
    solution_id: int,
    data: SolutionImplementRequest,
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None),
    user_name: Optional[str] = Query(None),
):
    service = AuditService(db)
    solution = service.implement_solution(
        solution_id=solution_id,
        actual_energy_saving=data.actual_energy_saving,
        actual_investment=data.actual_investment,
        user_id=user_id,
        user_name=user_name,
    )
    if not solution:
        raise HTTPException(status_code=404, detail="方案不存在或未通过筛选")
    return solution


@app.post("/audits/{audit_id}/acceptance", response_model=AuditResponse)
def perform_acceptance(
    audit_id: int,
    data: AcceptanceRequest,
    db: Session = Depends(get_db),
    user_id: Optional[str] = Query(None),
    user_name: Optional[str] = Query(None),
):
    service = AuditService(db)
    audit = service.perform_acceptance(
        audit_id=audit_id,
        score=data.score,
        notes=data.notes,
        user_id=user_id,
        user_name=user_name,
    )
    if not audit:
        raise HTTPException(status_code=404, detail="审核项目不存在")
    return audit


@app.get("/audits/{audit_id}/summary", response_model=AuditSummary)
def get_audit_summary(audit_id: int, db: Session = Depends(get_db)):
    service = AuditService(db)
    summary = service.get_summary(audit_id)
    if not summary:
        raise HTTPException(status_code=404, detail="审核项目不存在")
    return summary


@app.get("/logs", response_model=List[AuditLogResponse])
def list_logs(
    audit_id: Optional[int] = None,
    db: Session = Depends(get_db),
):
    service = AuditLogService(db)
    return service.get_logs(audit_id=audit_id)


@app.exception_handler(Exception)
async def exception_handler(request, exc):
    return JSONResponse(
        status_code=500,
        content={"detail": str(exc)},
    )
