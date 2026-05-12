from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/agent-services", tags=["agent-services"])


@router.post("/", response_model=schemas.AgentService)
def create_service(service: schemas.AgentServiceCreate, db: Session = Depends(get_db)):
    return services.create_agent_service(db, service)


@router.get("/", response_model=List[schemas.AgentService])
def list_services(
    status: Optional[models.AgentServiceStatus] = None,
    ship_name: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(models.AgentService)
    if status:
        query = query.filter(models.AgentService.status == status)
    if ship_name:
        query = query.filter(models.AgentService.ship_name.ilike(f"%{ship_name}%"))
    return query.order_by(models.AgentService.created_at.desc()).all()


@router.get("/{service_id}", response_model=schemas.AgentServiceDetail)
def get_service(service_id: int, db: Session = Depends(get_db)):
    service = db.query(models.AgentService).filter(models.AgentService.id == service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    return service


@router.put("/{service_id}", response_model=schemas.AgentService)
def update_service(service_id: int, data: schemas.AgentServiceUpdate, db: Session = Depends(get_db)):
    service = db.query(models.AgentService).filter(models.AgentService.id == service_id).first()
    if not service:
        raise HTTPException(status_code=404, detail="代理服务不存在")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(service, key, value)
    db.commit()
    db.refresh(service)
    return service


@router.post("/{service_id}/transition", response_model=schemas.AgentService)
def transition_service(
    service_id: int,
    transition: schemas.StatusTransition,
    db: Session = Depends(get_db)
):
    return services.transition_status(db, service_id, transition.target_status)


@router.post("/{service_id}/confirm-departure", response_model=schemas.AgentService)
def confirm_departure(service_id: int, db: Session = Depends(get_db)):
    return services.confirm_departure(db, service_id)
