from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from core.models import EnterpriseCreate, EnterpriseResponse
from core import services

router = APIRouter(prefix="/enterprises", tags=["enterprises"])


@router.post("", response_model=EnterpriseResponse)
def create(enterprise: EnterpriseCreate, db: Session = Depends(get_db)):
    return services.create_enterprise(db, enterprise.name)


@router.get("", response_model=List[EnterpriseResponse])
def list_all(db: Session = Depends(get_db)):
    return services.list_enterprises(db)


@router.get("/{enterprise_id}", response_model=EnterpriseResponse)
def get_one(enterprise_id: int, db: Session = Depends(get_db)):
    enterprise = services.get_enterprise(db, enterprise_id)
    if not enterprise:
        raise HTTPException(status_code=404, detail="Enterprise not found")
    return enterprise
