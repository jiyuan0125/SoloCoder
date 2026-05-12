from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.services import switch_service
from app.schemas import SwitchCreate, SwitchResponse
from pydantic import BaseModel
from typing import List

router = APIRouter()


class SwitchPositionRequest(BaseModel):
    position: str


class SwitchLockRequest(BaseModel):
    locked: bool


@router.post("/", response_model=SwitchResponse)
def create_switch(switch: SwitchCreate, db: Session = Depends(get_db)):
    existing = switch_service.get_switch(db, switch.device_id)
    if existing:
        raise HTTPException(status_code=400, detail="道岔已存在")
    return switch_service.create_switch(db, switch.device_id, switch.name)


@router.get("/", response_model=List[SwitchResponse])
def list_switches(db: Session = Depends(get_db)):
    return switch_service.get_all_switches(db)


@router.get("/{device_id}", response_model=SwitchResponse)
def get_switch(device_id: str, db: Session = Depends(get_db)):
    switch = switch_service.get_switch(db, device_id)
    if not switch:
        raise HTTPException(status_code=404, detail="道岔不存在")
    return switch


@router.post("/{device_id}/switch")
def switch_position(device_id: str, request: SwitchPositionRequest, db: Session = Depends(get_db)):
    result = switch_service.switch_position(db, device_id, request.position)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result


@router.post("/{device_id}/lock")
def lock_switch(device_id: str, request: SwitchLockRequest, db: Session = Depends(get_db)):
    result = switch_service.lock_switch(db, device_id, request.locked)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result
