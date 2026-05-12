from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/material-deliveries", tags=["materials"])


@router.post("/", response_model=schemas.MaterialDelivery)
def create_delivery(delivery: schemas.MaterialDeliveryCreate, db: Session = Depends(get_db)):
    return services.create_material_delivery(db, delivery)


@router.get("/", response_model=List[schemas.MaterialDelivery])
def list_deliveries(
    agent_service_id: int = None,
    status: models.DeliveryStatus = None,
    material_type: models.MaterialType = None,
    db: Session = Depends(get_db)
):
    query = db.query(models.MaterialDelivery)
    if agent_service_id:
        query = query.filter(models.MaterialDelivery.agent_service_id == agent_service_id)
    if status:
        query = query.filter(models.MaterialDelivery.status == status)
    if material_type:
        query = query.filter(models.MaterialDelivery.material_type == material_type)
    return query.order_by(models.MaterialDelivery.created_at.desc()).all()


@router.get("/{delivery_id}", response_model=schemas.MaterialDelivery)
def get_delivery(delivery_id: int, db: Session = Depends(get_db)):
    delivery = db.query(models.MaterialDelivery).filter(models.MaterialDelivery.id == delivery_id).first()
    if not delivery:
        raise HTTPException(status_code=404, detail="物料配送不存在")
    return delivery


@router.post("/{delivery_id}/complete", response_model=schemas.MaterialDelivery)
def complete_delivery(delivery_id: int, db: Session = Depends(get_db)):
    return services.complete_material_delivery(db, delivery_id)
