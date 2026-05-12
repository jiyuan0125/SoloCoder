from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import RectificationCreate, RectificationUpdate, RectificationResponse
from server import services

router = APIRouter(prefix="/rectifications", tags=["rectifications"])


@router.get("", response_model=List[RectificationResponse])
def list_rectifications(
    skip: int = 0,
    limit: int = 100,
    ship_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    return services.get_rectifications(db, skip=skip, limit=limit, ship_id=ship_id)


@router.post("", response_model=RectificationResponse)
def create_rectification(rect: RectificationCreate, db: Session = Depends(get_db)):
    db_rect = services.create_rectification(db, rect)
    db.commit()
    return db_rect


@router.get("/{rect_id}", response_model=RectificationResponse)
def get_rectification(rect_id: int, db: Session = Depends(get_db)):
    db_rect = services.get_rectification(db, rect_id)
    if not db_rect:
        raise HTTPException(status_code=404, detail="整改记录不存在")
    return db_rect


@router.put("/{rect_id}", response_model=RectificationResponse)
def update_rectification(
    rect_id: int,
    rect_update: RectificationUpdate,
    db: Session = Depends(get_db)
):
    try:
        db_rect = services.update_rectification(db, rect_id, rect_update)
        if not db_rect:
            raise HTTPException(status_code=404, detail="整改记录不存在")
        db.commit()
        return db_rect
    except ValueError as e:
        db.rollback()
        raise HTTPException(status_code=400, detail=str(e))
