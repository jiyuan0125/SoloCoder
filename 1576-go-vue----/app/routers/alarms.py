from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from app.database import get_db
from app.models import WorkOrderStatus
from app.schemas import AlarmWorkOrderResponse
from app.services import get_alarm_work_orders

router = APIRouter(prefix="/api/alarms", tags=["alarms"])


@router.get("", response_model=List[AlarmWorkOrderResponse])
def list_alarm_work_orders(
    status: Optional[WorkOrderStatus] = None,
    db: Session = Depends(get_db)
):
    return get_alarm_work_orders(db, status)
