from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import InspectionCreate, InspectionUpdate, InspectionResponse
from server import services

router = APIRouter(prefix="/inspections", tags=["inspections"])


@router.get("", response_model=List[InspectionResponse])
def list_inspections(
    skip: int = 0,
    limit: int = 100,
    ship_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    return services.get_inspections(db, skip=skip, limit=limit, ship_id=ship_id)


@router.post("", response_model=InspectionResponse)
def create_inspection(inspection: InspectionCreate, db: Session = Depends(get_db)):
    db_inspection = services.create_inspection(db, inspection)
    db.commit()
    return db_inspection


@router.get("/{inspection_id}", response_model=InspectionResponse)
def get_inspection(inspection_id: int, db: Session = Depends(get_db)):
    db_inspection = services.get_inspection(db, inspection_id)
    if not db_inspection:
        raise HTTPException(status_code=404, detail="检查记录不存在")
    return db_inspection


@router.put("/{inspection_id}", response_model=InspectionResponse)
def update_inspection(
    inspection_id: int,
    inspection_update: InspectionUpdate,
    db: Session = Depends(get_db)
):
    db_inspection = services.update_inspection(db, inspection_id, inspection_update)
    if not db_inspection:
        raise HTTPException(status_code=404, detail="检查记录不存在")
    db.commit()
    return db_inspection
