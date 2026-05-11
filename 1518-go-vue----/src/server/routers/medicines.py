from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from core import schemas
from core.services import MedicineService
from server.dependencies import get_db_session

router = APIRouter(prefix="/medicines", tags=["medicines"])


@router.post("", response_model=schemas.Medicine, status_code=status.HTTP_201_CREATED)
def create_medicine(
    medicine_in: schemas.MedicineCreate,
    db: Session = Depends(get_db_session)
):
    return MedicineService.create(db, medicine_in)


@router.get("", response_model=List[schemas.Medicine])
def list_medicines(db: Session = Depends(get_db_session)):
    return MedicineService.get_all(db)


@router.get("/{medicine_id}", response_model=schemas.Medicine)
def get_medicine(
    medicine_id: int,
    db: Session = Depends(get_db_session)
):
    medicine = MedicineService.get_by_id(db, medicine_id)
    if not medicine:
        raise HTTPException(status_code=404, detail="Medicine not found")
    return medicine


@router.put("/{medicine_id}", response_model=schemas.Medicine)
def update_medicine(
    medicine_id: int,
    medicine_in: schemas.MedicineUpdate,
    db: Session = Depends(get_db_session)
):
    medicine = MedicineService.update(db, medicine_id, medicine_in)
    if not medicine:
        raise HTTPException(status_code=404, detail="Medicine not found")
    return medicine


@router.delete("/{medicine_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_medicine(
    medicine_id: int,
    db: Session = Depends(get_db_session)
):
    if not MedicineService.delete(db, medicine_id):
        raise HTTPException(status_code=404, detail="Medicine not found")
    return None


@router.post("/{medicine_id}/add-stock", response_model=schemas.Medicine)
def add_medicine_stock(
    medicine_id: int,
    amount: float,
    db: Session = Depends(get_db_session)
):
    if not MedicineService.add_stock(db, medicine_id, amount):
        raise HTTPException(status_code=404, detail="Medicine not found")
    return MedicineService.get_by_id(db, medicine_id)
