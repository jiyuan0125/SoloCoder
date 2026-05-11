from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session

from core import schemas
from core.services import VisitRecordService
from server.dependencies import get_db_session

router = APIRouter(prefix="/visits", tags=["visits"])


@router.post("", response_model=schemas.VisitRecord, status_code=status.HTTP_201_CREATED)
def create_visit(
    visit_in: schemas.VisitRecordCreate,
    db: Session = Depends(get_db_session)
):
    visit = VisitRecordService.create(db, visit_in)
    if not visit:
        raise HTTPException(
            status_code=400,
            detail="Failed to create visit. May be duplicate for same farmer on same day, or insufficient medicine stock."
        )
    return visit


@router.get("", response_model=List[schemas.VisitRecord])
def list_visits(
    farmer_id: Optional[int] = Query(None),
    visit_date: Optional[date] = Query(None),
    db: Session = Depends(get_db_session)
):
    if farmer_id:
        return VisitRecordService.get_by_farmer(db, farmer_id)
    if visit_date:
        return VisitRecordService.get_by_date(db, visit_date)
    return VisitRecordService.get_all(db)


@router.get("/{visit_id}", response_model=schemas.VisitRecord)
def get_visit(
    visit_id: int,
    db: Session = Depends(get_db_session)
):
    visit = VisitRecordService.get_by_id(db, visit_id)
    if not visit:
        raise HTTPException(status_code=404, detail="Visit record not found")
    return visit
