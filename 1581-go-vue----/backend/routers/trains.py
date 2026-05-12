from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from .. import schemas, crud
from ..database import get_db

router = APIRouter(prefix="/trains", tags=["trains"])


@router.get("/", response_model=List[schemas.TrainResponse])
def read_trains(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud.get_trains(db, skip=skip, limit=limit)


@router.get("/{train_id}", response_model=schemas.TrainResponse)
def read_train(train_id: int, db: Session = Depends(get_db)):
    db_train = crud.get_train(db, train_id=train_id)
    if db_train is None:
        raise HTTPException(status_code=404, detail="Train not found")
    return db_train


@router.post("/", response_model=schemas.TrainResponse, status_code=status.HTTP_201_CREATED)
def create_train(train: schemas.TrainCreate, db: Session = Depends(get_db)):
    db_train = crud.get_train_by_number(db, train_number=train.train_number)
    if db_train:
        raise HTTPException(status_code=400, detail="Train with this number already exists")
    return crud.create_train(db=db, train=train)


@router.put("/{train_id}", response_model=schemas.TrainResponse)
def update_train(train_id: int, train: schemas.TrainUpdate, db: Session = Depends(get_db)):
    db_train = crud.get_train(db, train_id=train_id)
    if db_train is None:
        raise HTTPException(status_code=404, detail="Train not found")
    return crud.update_train(db=db, train_id=train_id, train=train)
