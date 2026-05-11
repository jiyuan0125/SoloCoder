from typing import List, Optional
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from ..dependencies import get_db
from ...core.schemas import LowStockAlertResponse
from ...core.service import LowStockAlertService


router = APIRouter(prefix="/alerts", tags=["alerts"])


@router.get("", response_model=List[LowStockAlertResponse])
def list_alerts(
    store_id: Optional[int] = Query(None),
    active_only: bool = Query(True),
    db: Session = Depends(get_db),
):
    service = LowStockAlertService(db)
    if store_id:
        return service.list_by_store(store_id, active_only)
    return service.list_all(active_only)
