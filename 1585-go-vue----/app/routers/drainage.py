from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models import ControlMode
from app.schemas import (
    DrainageDeviceCreate, DrainageDeviceResponse, DrainageLevelUpdate,
    DeviceOperationResponse, MaintenanceTaskResponse
)
from app.services.drainage_service import DrainageService

router = APIRouter(prefix="/drainage", tags=["drainage"])


@router.get("/", response_model=List[DrainageDeviceResponse])
def list_devices(db: Session = Depends(get_db)):
    return DrainageService.get_all_devices(db)


@router.get("/{device_id}", response_model=DrainageDeviceResponse)
def get_device(device_id: int, db: Session = Depends(get_db)):
    device = DrainageService.get_device(db, device_id)
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    return device


@router.get("/{device_id}/maintenance", response_model=List[MaintenanceTaskResponse])
def get_pending_maintenance(device_id: int, db: Session = Depends(get_db)):
    return DrainageService.get_pending_maintenance_tasks(db, device_id)


@router.post("/", response_model=DrainageDeviceResponse)
def create_device(device: DrainageDeviceCreate, db: Session = Depends(get_db)):
    return DrainageService.create_device(db, device.model_dump())


@router.post("/{device_id}/start", response_model=DeviceOperationResponse)
def start_device(device_id: int, db: Session = Depends(get_db)):
    device, message = DrainageService.start_device(db, device_id)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/{device_id}/stop", response_model=DeviceOperationResponse)
def stop_device(device_id: int, db: Session = Depends(get_db)):
    device, message = DrainageService.stop_device(db, device_id)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/{device_id}/level", response_model=DrainageDeviceResponse)
def update_level(
    device_id: int,
    level_update: DrainageLevelUpdate,
    db: Session = Depends(get_db)
):
    device = DrainageService.update_water_level(db, device_id, level_update.level)
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    if device.mode == ControlMode.AUTO.value:
        DrainageService.process_auto_control(db, device_id)
        db.refresh(device)
    
    return device


@router.post("/maintenance/{task_id}/complete", response_model=DeviceOperationResponse)
def complete_maintenance(task_id: int, db: Session = Depends(get_db)):
    task, message = DrainageService.complete_maintenance_task(db, task_id)
    if not task:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}
