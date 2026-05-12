from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.core.database import get_db
from server.schemas import (
    MaintenancePlanCreate,
    MaintenancePlanResponse,
    MaintenancePlanUpdate,
    WorkOrderAccept,
    WorkOrderAssign,
    WorkOrderCreate,
    WorkOrderExecute,
    WorkOrderResponse,
)
from server.services import MaintenanceService

router = APIRouter(tags=["维护管理"])


@router.post("/work-orders", response_model=WorkOrderResponse)
def create_work_order(data: WorkOrderCreate, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    return service.create_work_order(
        lighthouse_id=data.lighthouse_id,
        order_type=data.order_type,
        priority=data.priority,
        title=data.title,
        description=data.description,
    )


@router.get("/work-orders", response_model=List[WorkOrderResponse])
def list_work_orders(lighthouse_id: Optional[int] = None, status: Optional[str] = None, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    return service.get_work_orders(lighthouse_id, status)


@router.get("/work-orders/{order_id}", response_model=WorkOrderResponse)
def get_work_order(order_id: int, db: Session = Depends(get_db)):
    from server.models import WorkOrder

    order = db.query(WorkOrder).filter(WorkOrder.id == order_id).first()
    if not order:
        raise HTTPException(status_code=404, detail="工单不存在")
    return order


@router.post("/work-orders/{order_id}/assign", response_model=WorkOrderResponse)
def assign_work_order(order_id: int, data: WorkOrderAssign, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    try:
        order = service.assign_work_order(order_id, data.assigned_to)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if not order:
        raise HTTPException(status_code=404, detail="工单不存在")
    return order


@router.post("/work-orders/{order_id}/execute", response_model=WorkOrderResponse)
def execute_work_order(order_id: int, data: WorkOrderExecute, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    try:
        order = service.execute_work_order(order_id, data.executed_by, data.execution_notes)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if not order:
        raise HTTPException(status_code=404, detail="工单不存在")
    return order


@router.post("/work-orders/{order_id}/accept", response_model=WorkOrderResponse)
def accept_work_order(order_id: int, data: WorkOrderAccept, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    try:
        order = service.accept_work_order(order_id, data.accepted_by, data.acceptance_result)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if not order:
        raise HTTPException(status_code=404, detail="工单不存在")
    return order


@router.post("/maintenance-plans", response_model=MaintenancePlanResponse)
def create_maintenance_plan(data: MaintenancePlanCreate, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    return service.create_maintenance_plan(
        lighthouse_id=data.lighthouse_id,
        plan_type=data.plan_type,
        frequency_days=data.frequency_days,
        title=data.title,
        description=data.description,
        is_active=data.is_active,
    )


@router.get("/maintenance-plans", response_model=List[MaintenancePlanResponse])
def list_maintenance_plans(lighthouse_id: Optional[int] = None, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    return service.get_maintenance_plans(lighthouse_id)


@router.put("/maintenance-plans/{plan_id}", response_model=MaintenancePlanResponse)
def update_maintenance_plan(plan_id: int, data: MaintenancePlanUpdate, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    plan = service.update_maintenance_plan(plan_id, data.model_dump(exclude_unset=True))
    if not plan:
        raise HTTPException(status_code=404, detail="维护计划不存在")
    return plan


@router.get("/maintenance-plans/due")
def list_due_plans(db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    return service.check_due_plans()


@router.post("/maintenance-plans/{plan_id}/execute", response_model=WorkOrderResponse)
def execute_maintenance_plan(plan_id: int, db: Session = Depends(get_db)):
    service = MaintenanceService(db)
    try:
        order = service.execute_plan(plan_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if not order:
        raise HTTPException(status_code=404, detail="维护计划不存在")
    return order
