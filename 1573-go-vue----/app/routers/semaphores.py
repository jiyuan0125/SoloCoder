from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.services import semaphore_service
from app.schemas import SemaphoreCreate, SemaphoreResponse
from typing import List

router = APIRouter()


@router.post("/", response_model=SemaphoreResponse)
def create_semaphore(semaphore: SemaphoreCreate, db: Session = Depends(get_db)):
    existing = semaphore_service.get_semaphore(db, semaphore.device_id)
    if existing:
        raise HTTPException(status_code=400, detail="信号机已存在")
    return semaphore_service.create_semaphore(db, semaphore.device_id, semaphore.name)


@router.get("/", response_model=List[SemaphoreResponse])
def list_semaphores(db: Session = Depends(get_db)):
    return semaphore_service.get_all_semaphores(db)


@router.get("/{device_id}", response_model=SemaphoreResponse)
def get_semaphore(device_id: str, db: Session = Depends(get_db)):
    semaphore = semaphore_service.get_semaphore(db, device_id)
    if not semaphore:
        raise HTTPException(status_code=404, detail="信号机不存在")
    return semaphore


@router.post("/{device_id}/open")
def open_semaphore(device_id: str, db: Session = Depends(get_db)):
    result = semaphore_service.open_semaphore(db, device_id)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result


@router.post("/{device_id}/close")
def close_semaphore(device_id: str, db: Session = Depends(get_db)):
    result = semaphore_service.close_semaphore(db, device_id)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result
