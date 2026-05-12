from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/settlements", tags=["settlements"])


@router.post("/", response_model=schemas.Settlement)
def create_settlement(settlement: schemas.SettlementCreate, db: Session = Depends(get_db)):
    return services.create_settlement(db, settlement)


@router.get("/", response_model=List[schemas.Settlement])
def list_settlements(agent_service_id: int = None, db: Session = Depends(get_db)):
    query = db.query(models.Settlement)
    if agent_service_id:
        query = query.filter(models.Settlement.agent_service_id == agent_service_id)
    return query.order_by(models.Settlement.created_at.desc()).all()


@router.get("/{settlement_id}", response_model=schemas.Settlement)
def get_settlement(settlement_id: int, db: Session = Depends(get_db)):
    settlement = db.query(models.Settlement).filter(models.Settlement.id == settlement_id).first()
    if not settlement:
        raise HTTPException(status_code=404, detail="结算记录不存在")
    return settlement
