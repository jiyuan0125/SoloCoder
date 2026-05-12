from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Part, PurchaseAlert
from app.schemas import (
    PartCreate, PartUpdate, PartResponse, PurchaseAlertResponse, MessageResponse
)

router = APIRouter(prefix="/api/parts", tags=["配件管理"])


@router.post("/", response_model=PartResponse)
def create_part(part: PartCreate, db: Session = Depends(get_db)):
    existing = db.query(Part).filter(Part.part_code == part.part_code).first()
    if existing:
        raise HTTPException(status_code=400, detail="配件编码已存在")
    
    db_part = Part(
        part_code=part.part_code,
        part_name=part.part_name,
        stock=part.stock,
        warning_threshold=part.warning_threshold,
        unit=part.unit
    )
    db.add(db_part)
    db.commit()
    db.refresh(db_part)
    
    if db_part.stock < db_part.warning_threshold:
        alert = PurchaseAlert(
            part_id=db_part.id,
            current_stock=db_part.stock,
            threshold=db_part.warning_threshold,
            is_active=True
        )
        db.add(alert)
        db.commit()
    
    return db_part


@router.get("/", response_model=List[PartResponse])
def get_parts(
    part_code: str = None,
    part_name: str = None,
    db: Session = Depends(get_db)
):
    query = db.query(Part)
    if part_code:
        query = query.filter(Part.part_code.contains(part_code))
    if part_name:
        query = query.filter(Part.part_name.contains(part_name))
    return query.all()


@router.get("/{part_id}", response_model=PartResponse)
def get_part(part_id: int, db: Session = Depends(get_db)):
    part = db.query(Part).filter(Part.id == part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="配件不存在")
    return part


@router.get("/code/{part_code}", response_model=PartResponse)
def get_part_by_code(part_code: str, db: Session = Depends(get_db)):
    part = db.query(Part).filter(Part.part_code == part_code).first()
    if not part:
        raise HTTPException(status_code=404, detail="配件不存在")
    return part


@router.put("/{part_id}", response_model=PartResponse)
def update_part(
    part_id: int,
    part_update: PartUpdate,
    db: Session = Depends(get_db)
):
    part = db.query(Part).filter(Part.id == part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="配件不存在")
    
    if part_update.part_name is not None:
        part.part_name = part_update.part_name
    if part_update.stock is not None:
        part.stock = part_update.stock
    if part_update.warning_threshold is not None:
        part.warning_threshold = part_update.warning_threshold
    if part_update.unit is not None:
        part.unit = part_update.unit
    
    if part_update.stock is not None or part_update.warning_threshold is not None:
        existing_alert = db.query(PurchaseAlert).filter(
            PurchaseAlert.part_id == part.id,
            PurchaseAlert.is_active == True
        ).first()
        
        if part.stock < part.warning_threshold:
            if not existing_alert:
                alert = PurchaseAlert(
                    part_id=part.id,
                    current_stock=part.stock,
                    threshold=part.warning_threshold,
                    is_active=True
                )
                db.add(alert)
            else:
                existing_alert.current_stock = part.stock
                existing_alert.threshold = part.warning_threshold
        else:
            if existing_alert:
                existing_alert.is_active = False
    
    db.commit()
    db.refresh(part)
    return part


@router.post("/purchase-alerts/{alert_id}/resolve", response_model=MessageResponse)
def resolve_purchase_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = db.query(PurchaseAlert).filter(PurchaseAlert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="采购提醒不存在")
    
    alert.is_active = False
    db.commit()
    
    return MessageResponse(message="采购提醒已处理")


@router.get("/purchase-alerts/", response_model=List[PurchaseAlertResponse])
def get_purchase_alerts(active: bool = True, db: Session = Depends(get_db)):
    alerts = db.query(PurchaseAlert).filter(
        PurchaseAlert.is_active == active
    ).order_by(PurchaseAlert.created_at.desc()).all()
    
    result = []
    for alert in alerts:
        result.append(PurchaseAlertResponse(
            id=alert.id,
            part_code=alert.part.part_code,
            part_name=alert.part.part_name,
            current_stock=alert.current_stock,
            threshold=alert.threshold,
            is_active=alert.is_active,
            created_at=alert.created_at
        ))
    return result
