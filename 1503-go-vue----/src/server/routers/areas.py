from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from src.core import schemas, services
from src.core.database import get_db

router = APIRouter(prefix="/areas", tags=["areas"])


@router.post("/", response_model=schemas.AreaRead)
def create_area(area: schemas.AreaCreate, db: Session = Depends(get_db)):
    return services.create_area(db, area)


@router.get("/", response_model=List[schemas.AreaRead])
def list_areas(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_areas(db, skip=skip, limit=limit)


@router.get("/{area_id}", response_model=schemas.AreaDetail)
def get_area(area_id: int, db: Session = Depends(get_db)):
    area = services.get_area_by_id(db, area_id)
    if not area:
        raise HTTPException(status_code=404, detail="区域不存在")
    return area
