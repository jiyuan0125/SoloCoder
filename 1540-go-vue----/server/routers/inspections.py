from datetime import datetime, date, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session, selectinload
from sqlalchemy import select

from ..database import get_db
from ..config import settings
from ..models import (
    RadiationSource, SourceStatus, InspectionRecord, EmergencyTodo, TodoStatus
)
from ..schemas import InspectionCreate, Inspection as InspectionSchema, Todo, TodoUpdate
from ..utils import log_operation

router = APIRouter(tags=["巡检与应急管理"])


@router.get("/inspections", response_model=List[InspectionSchema])
def list_inspections(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    source_id: Optional[int] = None,
    abnormal_only: bool = Query(False),
    db: Session = Depends(get_db),
):
    stmt = select(InspectionRecord)
    if source_id:
        stmt = stmt.where(InspectionRecord.source_id == source_id)
    if abnormal_only:
        stmt = stmt.where(InspectionRecord.is_abnormal == 1)
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(InspectionRecord.inspection_time.desc())
    return db.execute(stmt).scalars().all()


@router.post("/inspections", response_model=InspectionSchema)
def create_inspection(inspection_in: InspectionCreate, db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == inspection_in.source_id)
    source = db.execute(stmt).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="放射源不存在")
    
    if source.status == SourceStatus.RETIRED:
        raise HTTPException(status_code=400, detail="已退役的放射源无需巡检")
    
    is_abnormal = 1 if inspection_in.radiation_dose > settings.RADIATION_LIMIT else 0
    result = "异常" if is_abnormal else "正常"
    
    inspection = InspectionRecord(
        source_id=inspection_in.source_id,
        inspector=inspection_in.inspector,
        radiation_dose=inspection_in.radiation_dose,
        is_abnormal=is_abnormal,
        result=result,
        remarks=inspection_in.remarks,
    )
    db.add(inspection)
    db.flush()
    
    if is_abnormal:
        todo = EmergencyTodo(
            inspection_id=inspection.id,
            source_id=source.id,
            source_code=source.source_code,
            dose_value=inspection_in.radiation_dose,
            status=TodoStatus.PENDING,
        )
        db.add(todo)
    
    db.commit()
    db.refresh(inspection)
    
    log_operation(db, "巡检记录", inspection_in.inspector, "InspectionRecord", inspection.id,
                  f"放射源巡检: {source.source_code}, 剂量={inspection_in.radiation_dose}, 结果={result}")
    
    return inspection


@router.get("/inspections/{inspection_id}", response_model=InspectionSchema)
def get_inspection(inspection_id: int, db: Session = Depends(get_db)):
    stmt = select(InspectionRecord).where(InspectionRecord.id == inspection_id)
    inspection = db.execute(stmt).scalar_one_or_none()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检记录不存在")
    return inspection


@router.get("/inspection-frequency")
def get_inspection_frequency():
    return {
        "一类": "每天1次",
        "二类": "每天1次",
        "三类": "每周1次",
        "四类": "每月1次",
        "五类": "每月1次",
    }


@router.get("/todos", response_model=List[Todo])
def list_todos(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    status: Optional[TodoStatus] = None,
    db: Session = Depends(get_db),
):
    stmt = select(EmergencyTodo)
    if status:
        stmt = stmt.where(EmergencyTodo.status == status)
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(EmergencyTodo.created_at.desc())
    return db.execute(stmt).scalars().all()


@router.get("/todos/{todo_id}", response_model=Todo)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    stmt = select(EmergencyTodo).where(EmergencyTodo.id == todo_id)
    todo = db.execute(stmt).scalar_one_or_none()
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@router.put("/todos/{todo_id}", response_model=Todo)
def update_todo(todo_id: int, todo_in: TodoUpdate, db: Session = Depends(get_db)):
    stmt = select(EmergencyTodo).where(EmergencyTodo.id == todo_id)
    todo = db.execute(stmt).scalar_one_or_none()
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    
    todo.status = todo_in.status
    todo.handler = todo_in.handler
    todo.handle_time = datetime.utcnow()
    todo.handle_result = todo_in.handle_result
    
    if todo_in.status == TodoStatus.RESOLVED:
        ins_stmt = select(InspectionRecord).where(InspectionRecord.id == todo.inspection_id)
        inspection = db.execute(ins_stmt).scalar_one_or_none()
        if inspection:
            inspection.result = "已处理"
            if inspection.remarks:
                inspection.remarks = inspection.remarks + f"\n处理结果: {todo_in.handle_result}"
            else:
                inspection.remarks = f"处理结果: {todo_in.handle_result}"
    
    db.commit()
    db.refresh(todo)
    
    log_operation(db, "处理应急待办", todo_in.handler, "EmergencyTodo", todo.id,
                  f"待办处理状态: {todo_in.status.value}, 处理结果: {todo_in.handle_result}")
    
    return todo
