from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional

from server.database import get_db
from server import crud, schemas

router = APIRouter(prefix="/api/warnings", tags=["预警管理"])


@router.get("/", response_model=List[schemas.WarningRecordResponse], summary="获取预警列表")
def get_warnings(
    skip: int = 0,
    limit: int = 100,
    canal_id: Optional[int] = None,
    year: Optional[int] = None,
    is_resolved: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    warnings = crud.get_warnings(
        db, skip=skip, limit=limit,
        canal_id=canal_id, year=year, is_resolved=is_resolved
    )
    return warnings


@router.post("/{warning_id}/resolve", response_model=schemas.WarningRecordResponse, summary="标记预警已解决")
def resolve_warning(warning_id: int, db: Session = Depends(get_db)):
    warning = crud.resolve_warning(db, warning_id=warning_id)
    if warning is None:
        raise HTTPException(status_code=404, detail="预警记录不存在")
    return warning
