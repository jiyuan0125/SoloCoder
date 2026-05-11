from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from core.models import MonthlyStatResponse
from core import services

router = APIRouter(prefix="/statistics", tags=["statistics"])


@router.get("/monthly/{year}/{month}", response_model=List[MonthlyStatResponse])
def monthly_stats(year: int, month: int, db: Session = Depends(get_db)):
    return services.get_all_monthly_metrics(db, year, month)


@router.get("/monthly/{year}/{month}/{facility_id}", response_model=MonthlyStatResponse)
def facility_monthly_stats(year: int, month: int, facility_id: int, db: Session = Depends(get_db)):
    metric = services.calculate_monthly_metrics(db, facility_id, year, month)
    if not metric:
        from fastapi import HTTPException
        raise HTTPException(status_code=404, detail="Facility not found")
    return metric
