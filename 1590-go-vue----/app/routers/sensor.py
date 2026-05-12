from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from datetime import datetime
from app.database import get_db
from app.models import Cabin
from app.schemas import (
    SensorWeightRequest, SensorWeightResponse,
    SensorStatusRequest, SensorStatusResponse
)
from app.services.weight_manager import WeightManager
from app.routers.cabin import get_cabin_or_404

router = APIRouter(prefix="/cabin", tags=["传感器管理"])

weight_manager = WeightManager()


@router.post("/{cabin_number}/sensor/weight", response_model=SensorWeightResponse)
def update_sensor_weight(
    cabin_number: str,
    request: SensorWeightRequest,
    db: Session = Depends(get_db)
):
    cabin = get_cabin_or_404(db, cabin_number)
    
    if request.weight < 0:
        raise HTTPException(status_code=400, detail="载重量不能为负数")
    
    cabin.current_weight = request.weight
    db.commit()
    db.refresh(cabin)
    
    weight_info = weight_manager.get_weight_info(request.weight, cabin.max_weight)
    
    return SensorWeightResponse(
        cabin_number=cabin_number,
        current_weight=request.weight,
        max_weight=cabin.max_weight,
        weight_ratio=weight_info["weight_ratio"],
        status=weight_info["status"],
        can_dispatch=weight_info["can_dispatch"]
    )


@router.post("/{cabin_number}/sensor/status", response_model=SensorStatusResponse)
def update_sensor_status(
    cabin_number: str,
    request: SensorStatusRequest,
    db: Session = Depends(get_db)
):
    cabin = get_cabin_or_404(db, cabin_number)
    
    old_status = cabin.sensor_status_ok
    cabin.sensor_status_ok = request.status_ok
    db.commit()
    db.refresh(cabin)
    
    if request.status_ok:
        if not old_status:
            message = "传感器已恢复正常"
        else:
            message = "传感器状态正常"
    else:
        message = "传感器故障，将使用估算载重量"
    
    return SensorStatusResponse(
        cabin_number=cabin_number,
        sensor_status_ok=request.status_ok,
        message=message
    )
