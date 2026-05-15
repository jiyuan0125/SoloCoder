from fastapi import APIRouter, Depends, HTTPException, Header, Query
from sqlalchemy.orm import Session
from sqlalchemy import and_
from typing import Optional, List
from datetime import datetime
import json

from app.core.database import get_db
from app.models.models import TaskConfig, ExecutionAudit, OperationAudit, TaskStatus, AuditAction
from app.schemas.schemas import (
    TaskConfigCreate, TaskConfigUpdate, TaskConfigResponse,
    ExecutionAuditResponse, OperationAuditResponse, AuditQueryParams
)
from app.core.scheduler import add_task_to_scheduler, remove_task_from_scheduler

router = APIRouter()


def get_auth_info(x_user: str = Header(..., alias="X-User"), x_source: str = Header(..., alias="X-Source")):
    return {"user": x_user, "source": x_source}


def record_operation_audit(db: Session, task_id: Optional[int], task_name: Optional[str],
                           action: AuditAction, operator: str, operation_source: str, details: dict):
    audit = OperationAudit(
        task_id=task_id,
        task_name=task_name,
        action=action,
        operator=operator,
        operation_source=operation_source,
        details=json.dumps(details, ensure_ascii=False)
    )
    db.add(audit)
    db.commit()


@router.post("/tasks", response_model=TaskConfigResponse)
def create_task(
    task: TaskConfigCreate,
    auth: dict = Depends(get_auth_info),
    db: Session = Depends(get_db)
):
    existing_task = db.query(TaskConfig).filter(TaskConfig.name == task.name).first()
    if existing_task:
        raise HTTPException(status_code=400, detail="Task name already exists")

    db_task = TaskConfig(
        name=task.name,
        cron_expression=task.cron_expression,
        callback_url=task.callback_url,
        timeout=task.timeout,
        max_retry=task.max_retry,
        is_active=task.is_active,
        created_by=auth["user"],
        operation_source=auth["source"]
    )
    db.add(db_task)
    db.commit()
    db.refresh(db_task)

    record_operation_audit(
        db, db_task.id, db_task.name, AuditAction.CREATE,
        auth["user"], auth["source"], task.model_dump()
    )

    if db_task.is_active:
        add_task_to_scheduler(db_task)

    return db_task


@router.get("/tasks", response_model=List[TaskConfigResponse])
def list_tasks(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    tasks = db.query(TaskConfig).offset(skip).limit(limit).all()
    return tasks


@router.get("/tasks/{task_id}", response_model=TaskConfigResponse)
def get_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(TaskConfig).filter(TaskConfig.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task


@router.put("/tasks/{task_id}", response_model=TaskConfigResponse)
def update_task(
    task_id: int,
    task_update: TaskConfigUpdate,
    auth: dict = Depends(get_auth_info),
    db: Session = Depends(get_db)
):
    db_task = db.query(TaskConfig).filter(TaskConfig.id == task_id).first()
    if not db_task:
        raise HTTPException(status_code=404, detail="Task not found")

    old_data = {
        "name": db_task.name,
        "cron_expression": db_task.cron_expression,
        "callback_url": db_task.callback_url,
        "timeout": db_task.timeout,
        "max_retry": db_task.max_retry,
        "is_active": db_task.is_active
    }

    if task_update.name is not None:
        existing_task = db.query(TaskConfig).filter(
            TaskConfig.name == task_update.name,
            TaskConfig.id != task_id
        ).first()
        if existing_task:
            raise HTTPException(status_code=400, detail="Task name already exists")
        db_task.name = task_update.name

    update_data = task_update.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(db_task, field, value)

    db.commit()
    db.refresh(db_task)

    record_operation_audit(
        db, db_task.id, db_task.name, AuditAction.UPDATE,
        auth["user"], auth["source"], {"old": old_data, "new": update_data}
    )

    add_task_to_scheduler(db_task)

    return db_task


@router.delete("/tasks/{task_id}")
def delete_task(
    task_id: int,
    auth: dict = Depends(get_auth_info),
    db: Session = Depends(get_db)
):
    db_task = db.query(TaskConfig).filter(TaskConfig.id == task_id).first()
    if not db_task:
        raise HTTPException(status_code=404, detail="Task not found")

    task_name = db_task.name
    remove_task_from_scheduler(task_id)

    db.delete(db_task)
    db.commit()

    record_operation_audit(
        db, task_id, task_name, AuditAction.DELETE,
        auth["user"], auth["source"], {"name": task_name}
    )

    return {"message": "Task deleted successfully"}


@router.get("/audits/executions", response_model=List[ExecutionAuditResponse])
def list_execution_audits(
    task_name: Optional[str] = Query(None),
    start_time: Optional[datetime] = Query(None),
    end_time: Optional[datetime] = Query(None),
    status: Optional[TaskStatus] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(ExecutionAudit)

    filters = []
    if task_name:
        filters.append(ExecutionAudit.task_name == task_name)
    if start_time:
        filters.append(ExecutionAudit.trigger_time >= start_time)
    if end_time:
        filters.append(ExecutionAudit.trigger_time <= end_time)
    if status:
        filters.append(ExecutionAudit.status == status)

    if filters:
        query = query.filter(and_(*filters))

    audits = query.order_by(ExecutionAudit.trigger_time.desc()).offset(skip).limit(limit).all()
    return audits


@router.get("/audits/executions/{audit_id}", response_model=ExecutionAuditResponse)
def get_execution_audit(audit_id: int, db: Session = Depends(get_db)):
    audit = db.query(ExecutionAudit).filter(ExecutionAudit.id == audit_id).first()
    if not audit:
        raise HTTPException(status_code=404, detail="Audit log not found")
    return audit


@router.get("/audits/operations", response_model=List[OperationAuditResponse])
def list_operation_audits(
    task_name: Optional[str] = Query(None),
    start_time: Optional[datetime] = Query(None),
    end_time: Optional[datetime] = Query(None),
    action: Optional[AuditAction] = Query(None),
    operator: Optional[str] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(OperationAudit)

    filters = []
    if task_name:
        filters.append(OperationAudit.task_name == task_name)
    if start_time:
        filters.append(OperationAudit.operation_time >= start_time)
    if end_time:
        filters.append(OperationAudit.operation_time <= end_time)
    if action:
        filters.append(OperationAudit.action == action)
    if operator:
        filters.append(OperationAudit.operator == operator)

    if filters:
        query = query.filter(and_(*filters))

    audits = query.order_by(OperationAudit.operation_time.desc()).offset(skip).limit(limit).all()
    return audits
