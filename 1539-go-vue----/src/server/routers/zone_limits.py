from fastapi import APIRouter, HTTPException
from typing import List

from src.core import storage
from src.core.models import ZoneLimit
from src.server.schemas import (
    ZoneLimitCreate,
    ZoneLimitUpdate,
    ZoneLimitResponse,
)

router = APIRouter(prefix="/zone-limits", tags=["zone-limits"])


@router.post("/", response_model=ZoneLimitResponse, status_code=201)
def create_zone_limit(zone_limit: ZoneLimitCreate):
    existing = storage.query(ZoneLimit, zone_type=zone_limit.zone_type)
    if existing:
        raise HTTPException(status_code=400, detail=f"Zone limit for {zone_limit.zone_type.value} already exists")
    
    db_zone_limit = ZoneLimit(
        id=0,
        zone_type=zone_limit.zone_type,
        daytime_limit=zone_limit.daytime_limit,
        nighttime_limit=zone_limit.nighttime_limit,
    )
    return storage.create(db_zone_limit)


@router.get("/", response_model=List[ZoneLimitResponse])
def get_zone_limits():
    return storage.get_all(ZoneLimit)


@router.get("/{zone_limit_id}", response_model=ZoneLimitResponse)
def get_zone_limit(zone_limit_id: int):
    zone_limit = storage.get_by_id(ZoneLimit, zone_limit_id)
    if not zone_limit:
        raise HTTPException(status_code=404, detail="Zone limit not found")
    return zone_limit


@router.put("/{zone_limit_id}", response_model=ZoneLimitResponse)
def update_zone_limit(zone_limit_id: int, zone_limit: ZoneLimitUpdate):
    db_zone_limit = storage.get_by_id(ZoneLimit, zone_limit_id)
    if not db_zone_limit:
        raise HTTPException(status_code=404, detail="Zone limit not found")
    
    update_data = zone_limit.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_zone_limit, key, value)
    
    return storage.update(db_zone_limit)


@router.delete("/{zone_limit_id}", status_code=204)
def delete_zone_limit(zone_limit_id: int):
    if not storage.delete(ZoneLimit, zone_limit_id):
        raise HTTPException(status_code=404, detail="Zone limit not found")
