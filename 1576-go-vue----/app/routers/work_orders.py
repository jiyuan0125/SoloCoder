from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from datetime import datetime
from typing import List, Optional
from app.database import get_db
from app.models import WorkOrderStatus, MaintenanceType
from app.schemas import WorkOrderCreate, WorkOrderUpdate, WorkOrderResponse
from app.services import (
    create_work_order, get_work_order, get_work_orders,
    update_work_order_status, get_today_work_orders,
    calculate_overdue
)

router = APIRouter(prefix="/api/work-orders", tags=["work-orders"])


@router.post("", response_model=WorkOrderResponse)
def create_new_work_order(order_data: WorkOrderCreate, db: Session = Depends(get_db)):
    if order_data.maintenance_type != MaintenanceType.FAULT_REPAIR:
        if not order_data.plan_start_time or not order_data.plan_end_time:
            raise HTTPException(
                status_code=400,
                detail="非故障抢修类型的工单必须指定计划开始和结束时间"
            )
    
    return create_work_order(db, order_data)


@router.get("", response_model=List[WorkOrderResponse])
def list_work_orders(
    section: Optional[str] = None,
    status: Optional[WorkOrderStatus] = None,
    db: Session = Depends(get_db)
):
    orders = get_work_orders(db, section, status)
    now = datetime.now()
    result = []
    for order in orders:
        order_dict = order.__dict__.copy()
        order_dict['is_overdue'] = calculate_overdue(order, now)
        result.append(order_dict)
    return result


@router.get("/today", response_model=List[WorkOrderResponse])
def list_today_work_orders(db: Session = Depends(get_db)):
    orders = get_today_work_orders(db)
    now = datetime.now()
    result = []
    for order in orders:
        order_dict = order.__dict__.copy()
        order_dict['is_overdue'] = calculate_overdue(order, now)
        result.append(order_dict)
    return result


@router.get("/{order_id}", response_model=WorkOrderResponse)
def get_single_work_order(order_id: int, db: Session = Depends(get_db)):
    order = get_work_order(db, order_id)
    if not order:
        raise HTTPException(status_code=404, detail="工单不存在")
    order_dict = order.__dict__.copy()
    order_dict['is_overdue'] = calculate_overdue(order, datetime.now())
    return order_dict


@router.put("/{order_id}", response_model=WorkOrderResponse)
def update_work_order(order_id: int, update_data: WorkOrderUpdate, db: Session = Depends(get_db)):
    order = update_work_order_status(db, order_id, update_data)
    if not order:
        raise HTTPException(status_code=404, detail="工单不存在")
    order_dict = order.__dict__.copy()
    order_dict['is_overdue'] = calculate_overdue(order, datetime.now())
    return order_dict
