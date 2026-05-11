from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import (
    AgentService,
    AgentServiceCreate,
    AgentServiceUpdate,
    AgentServiceDetail,
    Message,
)
from server import services

router = APIRouter(prefix="/api/agent-services", tags=["agent-services"])


@router.post("", response_model=AgentService)
def create_service(service_in: AgentServiceCreate, db: Session = Depends(get_db)):
    return services.create_agent_service(db, service_in)


@router.get("", response_model=List[AgentService])
def list_services(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_agent_services(db, skip=skip, limit=limit)


@router.get("/{service_id}", response_model=AgentServiceDetail)
def get_service(service_id: int, db: Session = Depends(get_db)):
    service = services.get_agent_service(db, service_id)
    if not service:
        raise HTTPException(status_code=404, detail="Agent service not found")
    return service


@router.patch("/{service_id}", response_model=AgentService)
def update_service(
    service_id: int,
    service_in: AgentServiceUpdate,
    db: Session = Depends(get_db),
):
    service = services.update_agent_service(db, service_id, service_in)
    if not service:
        raise HTTPException(status_code=404, detail="Agent service not found")
    return service


@router.post("/{service_id}/advance-stage", response_model=AgentService)
def advance_service_stage(
    service_id: int,
    notes: Optional[str] = None,
    db: Session = Depends(get_db),
):
    return services.advance_stage(db, service_id, notes)


@router.post("/{service_id}/confirm-departure", response_model=AgentService)
def confirm_departure(
    service_id: int,
    notes: Optional[str] = None,
    db: Session = Depends(get_db),
):
    return services.confirm_departure(db, service_id, notes)


@router.get("/{service_id}/departure-check", response_model=Message)
def check_departure_preconditions(service_id: int, db: Session = Depends(get_db)):
    return services.validate_departure_preconditions(db, service_id)
