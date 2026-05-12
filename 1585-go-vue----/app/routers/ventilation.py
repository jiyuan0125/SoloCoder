from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models import ControlMode
from app.schemas import (
    VentilationDeviceCreate, VentilationDeviceResponse, VentilationModeUpdate,
    DeviceOperationResponse
)
from app.services.ventilation_service import VentilationService

router = APIRouter(prefix="/ventilation", tags=["ventilation"])


@router.get("/", response_model=List[VentilationDeviceResponse])
def list_devices(db: Session = Depends(get_db)):
    return VentilationService.get_all_devices(db)


@router.get("/{device_id}", response_model=VentilationDeviceResponse)
def get_device(device_id: int, db: Session = Depends(get_db)):
    device = VentilationService.get_device(db, device_id)
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    return device


@router.post("/", response_model=VentilationDeviceResponse)
def create_device(device: VentilationDeviceCreate, db: Session = Depends(get_db)):
    return VentilationService.create_device(db, device.model_dump())


@router.post("/{device_id}/start", response_model=DeviceOperationResponse)
def start_device(device_id: int, db: Session = Depends(get_db)):
    device, message = VentilationService.start_device(db, device_id)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/{device_id}/stop", response_model=DeviceOperationResponse)
def stop_device(device_id: int, db: Session = Depends(get_db)):
    device, message = VentilationService.stop_device(db, device_id)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/{device_id}/mode", response_model=DeviceOperationResponse)
def set_mode(device_id: int, mode_update: VentilationModeUpdate, db: Session = Depends(get_db)):
    device, message = VentilationService.set_mode(db, device_id, mode_update.mode)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/{device_id}/sensor", response_model=VentilationDeviceResponse)
def update_sensor(
    device_id: int,
    co_level: float = None,
    visibility: float = None,
    db: Session = Depends(get_db)
):
    device = VentilationService.update_sensor_data(db, device_id, co_level, visibility)
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    if device.mode == ControlMode.AUTO.value:
        VentilationService.process_auto_control(db, device_id)
        db.refresh(device)
    
    return device
