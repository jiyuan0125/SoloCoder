from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import LicenseCreate, LicenseUpdate, LicenseResponse
from server import services

router = APIRouter(prefix="/licenses", tags=["licenses"])


@router.get("", response_model=List[LicenseResponse])
def list_licenses(
    skip: int = 0,
    limit: int = 100,
    ship_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    return services.get_licenses(db, skip=skip, limit=limit, ship_id=ship_id)


@router.post("", response_model=LicenseResponse)
def create_license(license_data: LicenseCreate, db: Session = Depends(get_db)):
    db_license = services.create_license(db, license_data)
    db.commit()
    return db_license


@router.get("/{license_id}", response_model=LicenseResponse)
def get_license(license_id: int, db: Session = Depends(get_db)):
    db_license = services.get_license(db, license_id)
    if not db_license:
        raise HTTPException(status_code=404, detail="许可不存在")
    return db_license


@router.put("/{license_id}", response_model=LicenseResponse)
def update_license(
    license_id: int,
    license_update: LicenseUpdate,
    db: Session = Depends(get_db)
):
    db_license = services.update_license(db, license_id, license_update)
    if not db_license:
        raise HTTPException(status_code=404, detail="许可不存在")
    db.commit()
    return db_license
