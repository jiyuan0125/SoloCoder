from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app import models, schemas
from app.utils import sort_devices_by_mileage, calculate_bridge_grade, get_inspection_frequency, get_next_month
from datetime import date

router = APIRouter(prefix="/devices", tags=["devices"])

@router.post("/", response_model=schemas.DeviceResponse)
def create_device(device: schemas.DeviceCreate, db: Session = Depends(get_db)):
    db_device = db.query(models.Device).filter(models.Device.device_code == device.device_code).first()
    if db_device:
        raise HTTPException(status_code=400, detail="设备编号已存在")
    
    device_dict = device.model_dump()
    if device.device_type == models.DeviceType.BRIDGE and device_dict.get('bridge_length'):
        device_dict['bridge_grade'] = calculate_bridge_grade(device_dict['bridge_length'])
    
    db_device = models.Device(**device_dict)
    db.add(db_device)
    db.commit()
    db.refresh(db_device)
    
    if db_device.device_type == models.DeviceType.BRIDGE:
        for inspection_type in [models.InspectionType.DAILY, models.InspectionType.FLOOD]:
            frequency = get_inspection_frequency(db_device.bridge_grade, inspection_type)
            plan = models.InspectionPlan(
                device_id=db_device.id,
                inspection_type=inspection_type,
                frequency_per_month=frequency,
                effective_month=date.today().replace(day=1)
            )
            db.add(plan)
        db.commit()
    
    return db_device

@router.get("/", response_model=List[schemas.DeviceResponse])
def get_devices(device_type: models.DeviceType = None, db: Session = Depends(get_db)):
    query = db.query(models.Device)
    if device_type:
        query = query.filter(models.Device.device_type == device_type)
    devices = query.all()
    return sort_devices_by_mileage(devices)

@router.get("/{device_code}", response_model=schemas.DeviceResponse)
def get_device(device_code: str, db: Session = Depends(get_db)):
    device = db.query(models.Device).filter(models.Device.device_code == device_code).first()
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    return device

@router.put("/{device_code}", response_model=schemas.DeviceResponse)
def update_device(device_code: str, device_update: schemas.DeviceUpdate, db: Session = Depends(get_db)):
    device = db.query(models.Device).filter(models.Device.device_code == device_code).first()
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    old_grade = device.bridge_grade
    update_dict = device_update.model_dump(exclude_unset=True)
    
    if 'bridge_length' in update_dict and update_dict['bridge_length']:
        update_dict['bridge_grade'] = calculate_bridge_grade(update_dict['bridge_length'])
    
    for key, value in update_dict.items():
        setattr(device, key, value)
    
    if old_grade and device.bridge_grade and old_grade != device.bridge_grade:
        active_plans = db.query(models.InspectionPlan).filter(
            models.InspectionPlan.device_id == device.id,
            models.InspectionPlan.is_active == 1
        ).all()
        
        for plan in active_plans:
            plan.is_active = 0
            
            new_frequency = get_inspection_frequency(device.bridge_grade, plan.inspection_type)
            new_plan = models.InspectionPlan(
                device_id=device.id,
                inspection_type=plan.inspection_type,
                frequency_per_month=new_frequency,
                effective_month=get_next_month(),
                is_active=1
            )
            db.add(new_plan)
    
    db.commit()
    db.refresh(device)
    return device

@router.put("/{device_code}/status", response_model=schemas.DeviceResponse)
def update_device_status(device_code: str, status_update: schemas.DeviceStatusUpdate, db: Session = Depends(get_db)):
    device = db.query(models.Device).filter(models.Device.device_code == device_code).first()
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    if device.status != status_update.status:
        history = models.DeviceStatusHistory(
            device_id=device.id,
            old_status=device.status,
            new_status=status_update.status,
            remark=status_update.remark
        )
        db.add(history)
        device.status = status_update.status
        db.commit()
        db.refresh(device)
    
    return device

@router.put("/{device_code}/grade", response_model=schemas.DeviceResponse)
def update_device_grade(device_code: str, grade: models.BridgeGrade, db: Session = Depends(get_db)):
    device = db.query(models.Device).filter(models.Device.device_code == device_code).first()
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    if device.device_type != models.DeviceType.BRIDGE:
        raise HTTPException(status_code=400, detail="只有桥梁设备有等级")
    
    if device.bridge_grade != grade:
        old_grade = device.bridge_grade
        device.bridge_grade = grade
        
        active_plans = db.query(models.InspectionPlan).filter(
            models.InspectionPlan.device_id == device.id,
            models.InspectionPlan.is_active == 1
        ).all()
        
        for plan in active_plans:
            plan.is_active = 0
            
            new_frequency = get_inspection_frequency(grade, plan.inspection_type)
            new_plan = models.InspectionPlan(
                device_id=device.id,
                inspection_type=plan.inspection_type,
                frequency_per_month=new_frequency,
                effective_month=get_next_month(),
                is_active=1
            )
            db.add(new_plan)
        
        db.commit()
        db.refresh(device)
    
    return device

@router.get("/{device_code}/status-history", response_model=List[schemas.DeviceStatusHistoryResponse])
def get_device_status_history(device_code: str, db: Session = Depends(get_db)):
    device = db.query(models.Device).filter(models.Device.device_code == device_code).first()
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    return device.status_history
