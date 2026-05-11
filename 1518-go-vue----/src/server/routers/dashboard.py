from datetime import date
from typing import Optional
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from core import schemas
from core.services import DashboardService
from server.dependencies import get_db_session

router = APIRouter(prefix="/dashboard", tags=["dashboard"])


@router.get("/stats", response_model=schemas.DashboardStats)
def get_monthly_stats(
    target_date: Optional[date] = Query(None),
    db: Session = Depends(get_db_session)
):
    return DashboardService.get_monthly_stats(db, target_date)
