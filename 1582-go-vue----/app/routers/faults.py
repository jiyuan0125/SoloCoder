from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks, status
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Fault, Priority, FaultStatus
from ..schemas import FaultCreate, FaultResponse, FaultUpdate

router = APIRouter(prefix="/faults", tags=["闸机故障"])


def escalate_faults_background(db: Session):
    now = datetime.utcnow()
    twenty_four_hours_ago = now - timedelta(hours=24)

    faults = db.query(Fault).filter(
        Fault.reported_at < twenty_four_hours_ago,
        Fault.status == FaultStatus.PENDING,
        Fault.escalated == False
    ).all()

    for fault in faults:
        fault.priority = Priority.HIGH
        fault.escalated = True

    if faults:
        db.commit()


@router.post("/", response_model=FaultResponse, status_code=status.HTTP_201_CREATED)
def create_fault(
    fault_data: FaultCreate,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db)
):
    fault = Fault(
        gate_id=fault_data.gate_id,
        station=fault_data.station,
        description=fault_data.description,
        priority=fault_data.priority,
        status=FaultStatus.PENDING
    )

    db.add(fault)
    db.commit()
    db.refresh(fault)

    background_tasks.add_task(escalate_faults_background, db)

    return fault


@router.get("/", response_model=List[FaultResponse])
def list_faults(
    status_filter: Optional[FaultStatus] = None,
    priority_filter: Optional[Priority] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Fault)
    if status_filter:
        query = query.filter(Fault.status == status_filter)
    if priority_filter:
        query = query.filter(Fault.priority == priority_filter)

    faults = query.order_by(Fault.id.desc()).offset(skip).limit(limit).all()
    return faults


@router.get("/{fault_id}", response_model=FaultResponse)
def get_fault(fault_id: int, db: Session = Depends(get_db)):
    fault = db.query(Fault).filter(Fault.id == fault_id).first()
    if not fault:
        raise HTTPException(status_code=404, detail="故障记录不存在")
    return fault


@router.put("/{fault_id}", response_model=FaultResponse)
def update_fault(
    fault_id: int,
    update_data: FaultUpdate,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db)
):
    fault = db.query(Fault).filter(Fault.id == fault_id).first()
    if not fault:
        raise HTTPException(status_code=404, detail="故障记录不存在")

    fault.status = update_data.status
    if update_data.status == FaultStatus.RESOLVED:
        fault.resolved_at = datetime.utcnow()

    db.commit()
    db.refresh(fault)

    background_tasks.add_task(escalate_faults_background, db)

    return fault


@router.post("/check-escalation")
def check_and_escalate(db: Session = Depends(get_db)):
    now = datetime.utcnow()
    twenty_four_hours_ago = now - timedelta(hours=24)

    faults = db.query(Fault).filter(
        Fault.reported_at < twenty_four_hours_ago,
        Fault.status == FaultStatus.PENDING,
        Fault.escalated == False
    ).all()

    count = 0
    for fault in faults:
        fault.priority = Priority.HIGH
        fault.escalated = True
        count += 1

    if faults:
        db.commit()

    return {
        "checked_at": now.isoformat(),
        "escalated_count": count,
        "message": f"已将{count}个超时未处理的故障升级为高优先级"
    }
