from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models import ControlMode
from app.schemas import (
    LightingDeviceCreate, LightingDeviceResponse, LightingBrightnessUpdate,
    LightingGroupUpdate, DeviceOperationResponse
)
from app.services.lighting_service import LightingService

router = APIRouter(prefix="/lighting", tags=["lighting"])


@router.get("/", response_model=List[LightingDeviceResponse])
def list_devices(db: Session = Depends(get_db)):
    return LightingService.get_all_devices(db)


@router.get("/{device_id}", response_model=LightingDeviceResponse)
def get_device(device_id: int, db: Session = Depends(get_db)):
    device = LightingService.get_device(db, device_id)
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    return device


@router.post("/", response_model=LightingDeviceResponse)
def create_device(device: LightingDeviceCreate, db: Session = Depends(get_db)):
    return LightingService.create_device(db, device.model_dump())


@router.post("/{device_id}/brightness", response_model=DeviceOperationResponse)
def set_brightness(
    device_id: int,
    brightness_update: LightingBrightnessUpdate,
    db: Session = Depends(get_db)
):
    device, message = LightingService.set_brightness(db, device_id, brightness_update.brightness)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/{device_id}/group", response_model=DeviceOperationResponse)
def set_group(
    device_id: int,
    group_update: LightingGroupUpdate,
    db: Session = Depends(get_db)
):
    device, message = LightingService.set_group(db, device_id, group_update.group)
    if not device:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/auto-update")
def auto_update_brightness(db: Session = Depends(get_db)):
    LightingService.update_auto_brightness(db)
    return {"success": True, "message": "自动亮度调整完成"}
