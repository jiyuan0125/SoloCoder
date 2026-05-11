from fastapi import APIRouter, HTTPException
from typing import List

from src.core import storage
from src.core.models import ConstructionPermit
from src.server.schemas import (
    ConstructionPermitCreate,
    ConstructionPermitUpdate,
    ConstructionPermitResponse,
)

router = APIRouter(prefix="/construction-permits", tags=["construction-permits"])


@router.post("/", response_model=ConstructionPermitResponse, status_code=201)
def create_permit(permit: ConstructionPermitCreate):
    db_permit = ConstructionPermit(
        id=0,
        point_ids=permit.point_ids,
        permit_number=permit.permit_number,
        start_date=permit.start_date,
        end_date=permit.end_date,
        description=permit.description,
        is_active=permit.is_active,
    )
    return storage.create(db_permit)


@router.get("/", response_model=List[ConstructionPermitResponse])
def get_permits():
    return storage.get_all(ConstructionPermit)


@router.get("/{permit_id}", response_model=ConstructionPermitResponse)
def get_permit(permit_id: int):
    permit = storage.get_by_id(ConstructionPermit, permit_id)
    if not permit:
        raise HTTPException(status_code=404, detail="Construction permit not found")
    return permit


@router.put("/{permit_id}", response_model=ConstructionPermitResponse)
def update_permit(permit_id: int, permit: ConstructionPermitUpdate):
    db_permit = storage.get_by_id(ConstructionPermit, permit_id)
    if not db_permit:
        raise HTTPException(status_code=404, detail="Construction permit not found")
    
    update_data = permit.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_permit, key, value)
    
    return storage.update(db_permit)


@router.delete("/{permit_id}", status_code=204)
def delete_permit(permit_id: int):
    if not storage.delete(ConstructionPermit, permit_id):
        raise HTTPException(status_code=404, detail="Construction permit not found")
