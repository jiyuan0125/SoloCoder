from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.services import interlocking_service
from app.schemas import InterlockingCreate, InterlockingResponse, RouteRequest
from typing import List

router = APIRouter()


@router.post("/", response_model=InterlockingResponse)
def create_interlocking(interlocking: InterlockingCreate, db: Session = Depends(get_db)):
    existing = interlocking_service.get_interlocking(db, interlocking.device_id)
    if existing:
        raise HTTPException(status_code=400, detail="联锁设备已存在")
    return interlocking_service.create_interlocking(db, interlocking.device_id, interlocking.name)


@router.get("/", response_model=List[InterlockingResponse])
def list_interlockings(db: Session = Depends(get_db)):
    return interlocking_service.get_all_interlockings(db)


@router.get("/{device_id}", response_model=InterlockingResponse)
def get_interlocking(device_id: str, db: Session = Depends(get_db)):
    interlocking = interlocking_service.get_interlocking(db, device_id)
    if not interlocking:
        raise HTTPException(status_code=404, detail="联锁设备不存在")
    return interlocking


@router.post("/{device_id}/set-route")
def set_route(device_id: str, route_request: RouteRequest, db: Session = Depends(get_db)):
    result = interlocking_service.set_route(db, device_id, route_request.dict())
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result


@router.post("/{device_id}/cancel-route")
def cancel_route(device_id: str, db: Session = Depends(get_db)):
    result = interlocking_service.cancel_route(db, device_id)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result
