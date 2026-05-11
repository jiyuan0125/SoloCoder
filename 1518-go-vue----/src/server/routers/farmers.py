from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from core import schemas
from core.services import FarmerService
from server.dependencies import get_db_session

router = APIRouter(prefix="/farmers", tags=["farmers"])


@router.post("", response_model=schemas.Farmer, status_code=status.HTTP_201_CREATED)
def create_farmer(
    farmer_in: schemas.FarmerCreate,
    db: Session = Depends(get_db_session)
):
    return FarmerService.create(db, farmer_in)


@router.get("", response_model=List[schemas.Farmer])
def list_farmers(db: Session = Depends(get_db_session)):
    return FarmerService.get_all(db)


@router.get("/{farmer_id}", response_model=schemas.Farmer)
def get_farmer(
    farmer_id: int,
    db: Session = Depends(get_db_session)
):
    farmer = FarmerService.get_by_id(db, farmer_id)
    if not farmer:
        raise HTTPException(status_code=404, detail="Farmer not found")
    return farmer


@router.put("/{farmer_id}", response_model=schemas.Farmer)
def update_farmer(
    farmer_id: int,
    farmer_in: schemas.FarmerUpdate,
    db: Session = Depends(get_db_session)
):
    farmer = FarmerService.update(db, farmer_id, farmer_in)
    if not farmer:
        raise HTTPException(status_code=404, detail="Farmer not found")
    return farmer


@router.delete("/{farmer_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_farmer(
    farmer_id: int,
    db: Session = Depends(get_db_session)
):
    if not FarmerService.delete(db, farmer_id):
        raise HTTPException(status_code=404, detail="Farmer not found")
    return None
