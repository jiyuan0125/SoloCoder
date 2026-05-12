from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime
from app.database import get_db
from app.models import AlertType, AlertStatus
from app.schemas.models import DelayAdjustmentResponse, RouteOptimizationResponse, AlertResponse
from app.services.delay_service import DelayService
from app.services.optimization_service import OptimizationService, WorkHoursService

router = APIRouter()


@router.post("/routes/{route_id}/check-delay")
def check_delay(route_id: int, actual_departure: Optional[datetime] = None,
                actual_arrival: Optional[datetime] = None, db: Session = Depends(get_db)):
    try:
        adjustment = DelayService.check_and_handle_delay(db, route_id, actual_departure, actual_arrival)
        return {
            "status": "success",
            "adjustment_created": adjustment is not None,
            "adjustment": adjustment
        }
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/routes/{route_id}/recover-delay")
def recover_delay(route_id: int, db: Session = Depends(get_db)):
    try:
        DelayService.recover_from_delay(db, route_id)
        return {"status": "success", "message": "晚点已恢复，月度工时将重新核算"}
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/delay-adjustments", response_model=List[DelayAdjustmentResponse])
def list_delay_adjustments(route_id: Optional[int] = None, limit: int = 100, db: Session = Depends(get_db)):
    return DelayService.get_delay_adjustments(db, route_id, limit)


@router.get("/train-utilization")
def get_train_utilization(train_id: int, days: int = 7, db: Session = Depends(get_db)):
    from datetime import date, timedelta
    end_date = date.today()
    start_date = end_date - timedelta(days=days - 1)
    return OptimizationService.calculate_train_utilization(db, train_id, start_date, end_date)


@router.post("/generate-optimizations")
def generate_optimizations(days: int = 7, db: Session = Depends(get_db)):
    suggestions = OptimizationService.generate_optimization_suggestions(db, days)
    return {
        "status": "success",
        "suggestions_count": len(suggestions),
        "suggestions": suggestions
    }


@router.get("/optimizations", response_model=List[RouteOptimizationResponse])
def list_optimizations(train_id: Optional[int] = None, is_implemented: Optional[bool] = None, 
                       db: Session = Depends(get_db)):
    return OptimizationService.get_optimizations(db, train_id, is_implemented)


@router.post("/optimizations/{opt_id}/implement")
def implement_optimization(opt_id: int, db: Session = Depends(get_db)):
    try:
        return OptimizationService.mark_implemented(db, opt_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/alerts", response_model=List[AlertResponse])
def list_alerts(status: Optional[AlertStatus] = None, alert_type: Optional[AlertType] = None,
                limit: int = 100, db: Session = Depends(get_db)):
    return WorkHoursService.get_alerts(db, status, alert_type, limit)


@router.post("/alerts/{alert_id}/acknowledge", response_model=AlertResponse)
def acknowledge_alert(alert_id: int, db: Session = Depends(get_db)):
    try:
        return WorkHoursService.acknowledge_alert(db, alert_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/alerts/{alert_id}/resolve", response_model=AlertResponse)
def resolve_alert(alert_id: int, db: Session = Depends(get_db)):
    try:
        return WorkHoursService.resolve_alert(db, alert_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))
