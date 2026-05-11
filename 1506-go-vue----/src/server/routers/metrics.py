from typing import Optional, List
from datetime import date
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from ..dependencies import get_db
from ...core.schemas import MetricsResponse, WastageRankingResponse
from ...core.service import MetricsService


router = APIRouter(prefix="/metrics", tags=["metrics"])


@router.get("/summary", response_model=MetricsResponse)
def get_metrics(
    store_ids: Optional[List[int]] = Query(None),
    start_date: date = Query(...),
    end_date: date = Query(...),
    db: Session = Depends(get_db),
):
    return MetricsService(db).get_metrics(store_ids, start_date, end_date)


@router.get("/wastage-ranking", response_model=WastageRankingResponse)
def get_wastage_ranking(
    store_ids: Optional[List[int]] = Query(None),
    start_date: date = Query(...),
    end_date: date = Query(...),
    db: Session = Depends(get_db),
):
    return MetricsService(db).get_wastage_ranking(store_ids, start_date, end_date)
