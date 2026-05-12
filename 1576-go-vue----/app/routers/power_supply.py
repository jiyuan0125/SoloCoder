from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.schemas import PowerSupplyDataCreate, PowerSupplyDataResponse, SectionStatisticsResponse
from app.services import (
    create_power_supply_data, get_latest_power_supply_data,
    get_power_supply_history, get_or_create_section_statistics
)

router = APIRouter(prefix="/api/power-supply", tags=["power-supply"])


@router.post("/data", response_model=PowerSupplyDataResponse)
def add_power_supply_data(data: PowerSupplyDataCreate, db: Session = Depends(get_db)):
    return create_power_supply_data(db, data)


@router.get("/data/{section}/latest", response_model=PowerSupplyDataResponse)
def get_latest_data(section: str, db: Session = Depends(get_db)):
    data = get_latest_power_supply_data(db, section)
    if not data:
        raise HTTPException(status_code=404, detail="未找到该区段的供电数据")
    return data


@router.get("/data/{section}/history", response_model=List[PowerSupplyDataResponse])
def get_data_history(section: str, limit: int = 100, db: Session = Depends(get_db)):
    return get_power_supply_history(db, section, limit)


@router.get("/statistics/{section}", response_model=SectionStatisticsResponse)
def get_statistics(section: str, db: Session = Depends(get_db)):
    return get_or_create_section_statistics(db, section)
