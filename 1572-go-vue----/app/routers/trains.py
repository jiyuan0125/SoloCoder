from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Optional
from app.database import get_db
from app.models import TrainStatus
from app.schemas.models import TrainCreate, TrainUpdate, TrainResponse
from app.services.train_service import TrainService

router = APIRouter()


@router.post("/", response_model=TrainResponse, status_code=status.HTTP_201_CREATED)
def create_train(train: TrainCreate, db: Session = Depends(get_db)):
    try:
        return TrainService.create_train(db, train)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/", response_model=List[TrainResponse])
def list_trains(status: Optional[TrainStatus] = None, db: Session = Depends(get_db)):
    return TrainService.list_trains(db, status)


@router.get("/{train_id}", response_model=TrainResponse)
def get_train(train_id: int, db: Session = Depends(get_db)):
    train = TrainService.get_train(db, train_id)
    if not train:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="列车不存在")
    return train


@router.patch("/{train_id}", response_model=TrainResponse)
def update_train(train_id: int, train_update: TrainUpdate, db: Session = Depends(get_db)):
    train = TrainService.update_train(db, train_id, train_update)
    if not train:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="列车不存在")
    return train


@router.delete("/{train_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_train(train_id: int, db: Session = Depends(get_db)):
    if not TrainService.delete_train(db, train_id):
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="列车不存在")
