from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from datetime import datetime
from app.database import get_db
from app.models import (
    Cabin, CabinStatus, TripRecord, MaintenanceTask, MaintenanceStatus
)
from app.schemas import (
    DispatchRequest, DispatchResponse, ArriveResponse,
    SensorWeightRequest, SensorWeightResponse,
    SensorStatusRequest, SensorStatusResponse
)
from app.services.queue_manager import QueueManager
from app.services.weight_manager import WeightManager
from app.services.maintenance_checker import MaintenanceChecker

router = APIRouter(prefix="/cabin", tags=["缆车运营"])

weight_manager = WeightManager()


def get_cabin_or_404(db: Session, cabin_number: str) -> Cabin:
    cabin = db.query(Cabin).filter(Cabin.cabin_number == cabin_number).first()
    if not cabin:
        raise HTTPException(status_code=404, detail=f"轿厢 {cabin_number} 不存在")
    return cabin


@router.post("/{cabin_number}/dispatch", response_model=DispatchResponse)
def dispatch_cabin(
    cabin_number: str,
    request: DispatchRequest,
    db: Session = Depends(get_db)
):
    cabin = get_cabin_or_404(db, cabin_number)
    maintenance_checker = MaintenanceChecker(db)
    
    can_operate, reasons = maintenance_checker.can_operate(cabin.id)
    if not can_operate:
        raise HTTPException(
            status_code=400,
            detail=f"无法发车：{'; '.join(reasons)}"
        )
    
    if cabin.current_status not in [CabinStatus.IDLE, CabinStatus.DOWNWARD]:
        raise HTTPException(
            status_code=400,
            detail=f"轿厢 {cabin_number} 当前状态为 {cabin.current_status.value}，无法发车"
        )
    
    if request.is_first_trip and not request.has_passengers:
        weight = 0.0
        passenger_count = None
        weight_is_estimated = False
        has_passengers = False
    else:
        queue_manager = QueueManager(db)
        items, passenger_count, weight = queue_manager.get_next_queue_items(cabin.max_weight)
        
        if not cabin.sensor_status_ok:
            weight_is_estimated = True
            weight = weight_manager.calculate_weight(passenger_count)
        else:
            weight_is_estimated = False
            if cabin.current_weight > 0:
                weight = cabin.current_weight
        
        if passenger_count > 0:
            queue_manager.mark_items_served(items, cabin.id)
    
    weight_info = weight_manager.get_weight_info(weight, cabin.max_weight)
    
    if not weight_info["can_dispatch"]:
        raise HTTPException(
            status_code=400,
            detail=f"载重量超标（{weight:.1f}kg / {cabin.max_weight}kg），禁止发车"
        )
    
    now = datetime.utcnow()
    cabin.current_status = CabinStatus.UPWARD
    cabin.current_weight = weight
    cabin.passenger_count = passenger_count if passenger_count else 0
    cabin.last_dispatch_time = now
    cabin.current_direction = "upward"
    
    trip_record = TripRecord(
        cabin_id=cabin.id,
        dispatch_time=now,
        direction="upward",
        weight=weight,
        weight_is_estimated=weight_is_estimated,
        passenger_count=passenger_count,
        is_first_trip=request.is_first_trip,
        has_passengers=has_passengers if request.is_first_trip else (passenger_count > 0)
    )
    
    db.add(trip_record)
    db.commit()
    db.refresh(cabin)
    db.refresh(trip_record)
    
    message = "发车成功"
    if weight_info["status"] == "warning":
        message = f"发车成功（载重量警告：已达 {weight_info['weight_ratio']*100:.1f}%）"
    if weight_is_estimated:
        message += " [使用估算载重量]"
    
    return DispatchResponse(
        success=True,
        message=message,
        cabin_number=cabin_number,
        dispatch_time=now,
        weight=weight,
        weight_is_estimated=weight_is_estimated,
        passenger_count=passenger_count,
        weight_status=weight_info["status"],
        direction="upward"
    )


@router.post("/{cabin_number}/arrive", response_model=ArriveResponse)
def arrive_cabin(
    cabin_number: str,
    db: Session = Depends(get_db)
):
    cabin = get_cabin_or_404(db, cabin_number)
    
    if cabin.current_status != CabinStatus.UPWARD:
        raise HTTPException(
            status_code=400,
            detail=f"轿厢 {cabin_number} 当前不在上行状态"
        )
    
    now = datetime.utcnow()
    cabin.current_status = CabinStatus.IDLE
    cabin.last_arrive_time = now
    cabin.current_weight = 0.0
    cabin.passenger_count = 0
    
    trip_record = db.query(TripRecord).filter(
        TripRecord.cabin_id == cabin.id,
        TripRecord.arrive_time == None
    ).order_by(TripRecord.dispatch_time.desc()).first()
    
    if trip_record:
        trip_record.arrive_time = now
    
    db.commit()
    
    return ArriveResponse(
        success=True,
        message=f"轿厢 {cabin_number} 到达",
        cabin_number=cabin_number,
        arrive_time=now
    )
