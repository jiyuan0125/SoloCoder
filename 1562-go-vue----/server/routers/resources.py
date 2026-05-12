from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from server.models import CleaningTeam, FuelTruck, ShuttleBus, FuelTruckStatus, BusStatus
from server.schemas import (
    CleaningTeamCreate, CleaningTeamResponse,
    FuelTruckCreate, FuelTruckResponse,
    ShuttleBusCreate, ShuttleBusResponse
)

router = APIRouter(prefix="/resources", tags=["resources"])


@router.post("/cleaning-teams", response_model=CleaningTeamResponse, status_code=status.HTTP_201_CREATED)
def create_cleaning_team(team: CleaningTeamCreate, db: Session = Depends(get_db)):
    db_team = CleaningTeam(**team.model_dump())
    db.add(db_team)
    db.commit()
    db.refresh(db_team)
    return db_team


@router.get("/cleaning-teams", response_model=List[CleaningTeamResponse])
def list_cleaning_teams(db: Session = Depends(get_db)):
    return db.query(CleaningTeam).all()


@router.get("/cleaning-teams/{team_id}", response_model=CleaningTeamResponse)
def get_cleaning_team(team_id: int, db: Session = Depends(get_db)):
    team = db.query(CleaningTeam).filter(CleaningTeam.id == team_id).first()
    if not team:
        raise HTTPException(status_code=404, detail="清洁班组不存在")
    return team


@router.post("/fuel-trucks", response_model=FuelTruckResponse, status_code=status.HTTP_201_CREATED)
def create_fuel_truck(truck: FuelTruckCreate, db: Session = Depends(get_db)):
    db_truck = FuelTruck(**truck.model_dump())
    db.add(db_truck)
    db.commit()
    db.refresh(db_truck)
    return db_truck


@router.get("/fuel-trucks", response_model=List[FuelTruckResponse])
def list_fuel_trucks(db: Session = Depends(get_db)):
    return db.query(FuelTruck).all()


@router.get("/fuel-trucks/{truck_id}", response_model=FuelTruckResponse)
def get_fuel_truck(truck_id: int, db: Session = Depends(get_db)):
    truck = db.query(FuelTruck).filter(FuelTruck.id == truck_id).first()
    if not truck:
        raise HTTPException(status_code=404, detail="加油车不存在")
    return truck


@router.patch("/fuel-trucks/{truck_id}/refuel")
def refuel_truck(truck_id: int, db: Session = Depends(get_db)):
    truck = db.query(FuelTruck).filter(FuelTruck.id == truck_id).first()
    if not truck:
        raise HTTPException(status_code=404, detail="加油车不存在")

    if truck.status == FuelTruckStatus.IN_USE:
        raise HTTPException(status_code=400, detail="加油车正在使用中，无法补油")

    truck.current_fuel = truck.capacity
    db.commit()
    db.refresh(truck)

    return {"message": "补油完成", "truck": FuelTruckResponse.model_validate(truck)}


@router.post("/shuttle-buses", response_model=ShuttleBusResponse, status_code=status.HTTP_201_CREATED)
def create_shuttle_bus(bus: ShuttleBusCreate, db: Session = Depends(get_db)):
    db_bus = ShuttleBus(**bus.model_dump())
    db.add(db_bus)
    db.commit()
    db.refresh(db_bus)
    return db_bus


@router.get("/shuttle-buses", response_model=List[ShuttleBusResponse])
def list_shuttle_buses(db: Session = Depends(get_db)):
    return db.query(ShuttleBus).all()


@router.get("/shuttle-buses/{bus_id}", response_model=ShuttleBusResponse)
def get_shuttle_bus(bus_id: int, db: Session = Depends(get_db)):
    bus = db.query(ShuttleBus).filter(ShuttleBus.id == bus_id).first()
    if not bus:
        raise HTTPException(status_code=404, detail="摆渡车不存在")
    return bus


@router.patch("/shuttle-buses/{bus_id}/status/{new_status}")
def update_bus_status(bus_id: int, new_status: BusStatus, db: Session = Depends(get_db)):
    bus = db.query(ShuttleBus).filter(ShuttleBus.id == bus_id).first()
    if not bus:
        raise HTTPException(status_code=404, detail="摆渡车不存在")

    bus.status = new_status
    db.commit()
    db.refresh(bus)

    return {"message": "状态更新成功", "bus": ShuttleBusResponse.model_validate(bus)}


@router.patch("/fuel-trucks/{truck_id}/status/{new_status}")
def update_truck_status(truck_id: int, new_status: FuelTruckStatus, db: Session = Depends(get_db)):
    truck = db.query(FuelTruck).filter(FuelTruck.id == truck_id).first()
    if not truck:
        raise HTTPException(status_code=404, detail="加油车不存在")

    truck.status = new_status
    db.commit()
    db.refresh(truck)

    return {"message": "状态更新成功", "truck": FuelTruckResponse.model_validate(truck)}
