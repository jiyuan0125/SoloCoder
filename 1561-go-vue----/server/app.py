import os
from typing import List
from fastapi import FastAPI, Depends, HTTPException, status
from sqlalchemy.orm import Session
from .database import engine, Base, get_db
from . import models, schemas, services


Base.metadata.create_all(bind=engine)

app = FastAPI(title="航班运行保障系统", version="1.0.0")


@app.get("/flights/", response_model=List[schemas.FlightResponse])
def list_flights(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_flights(db, skip=skip, limit=limit)


@app.get("/flights/{flight_id}", response_model=schemas.FlightResponse)
def read_flight(flight_id: int, db: Session = Depends(get_db)):
    flight = services.get_flight(db, flight_id=flight_id)
    if flight is None:
        raise HTTPException(status_code=404, detail="Flight not found")
    return flight


@app.post("/flights/", response_model=schemas.FlightResponse, status_code=status.HTTP_201_CREATED)
def create_flight(flight: schemas.FlightCreate, db: Session = Depends(get_db)):
    db_flight = services.get_flight_by_number(db, flight_number=flight.flight_number)
    if db_flight:
        raise HTTPException(status_code=400, detail="Flight number already registered")
    return services.create_flight(db=db, flight=flight)


@app.put("/flights/{flight_id}", response_model=schemas.FlightResponse)
def update_flight(flight_id: int, flight_update: schemas.FlightUpdate, db: Session = Depends(get_db)):
    try:
        flight = services.update_flight(db, flight_id=flight_id, flight_update=flight_update)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if flight is None:
        raise HTTPException(status_code=404, detail="Flight not found")
    return flight


@app.delete("/flights/{flight_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_flight(flight_id: int, db: Session = Depends(get_db)):
    success = services.delete_flight(db, flight_id=flight_id)
    if not success:
        raise HTTPException(status_code=404, detail="Flight not found")
    return None


@app.post("/stands/assign", response_model=schemas.StandResponse)
def assign_stand(request: schemas.StandAssignRequest, db: Session = Depends(get_db)):
    stand, error = services.assign_stand_to_flight(db, flight_id=request.flight_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return stand


@app.get("/stands/", response_model=List[schemas.StandResponse])
def list_stands(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_stands(db, skip=skip, limit=limit)


@app.get("/stands/{stand_id}", response_model=schemas.StandResponse)
def read_stand(stand_id: int, db: Session = Depends(get_db)):
    stand = services.get_stand(db, stand_id=stand_id)
    if stand is None:
        raise HTTPException(status_code=404, detail="Stand not found")
    return stand


@app.post("/stands/", response_model=schemas.StandResponse, status_code=status.HTTP_201_CREATED)
def create_stand(stand: schemas.StandCreate, db: Session = Depends(get_db)):
    return services.create_stand(db=db, stand=stand)


@app.put("/stands/{stand_id}", response_model=schemas.StandResponse)
def update_stand(stand_id: int, stand_update: schemas.StandUpdate, db: Session = Depends(get_db)):
    stand = services.update_stand(db, stand_id=stand_id, stand_update=stand_update)
    if stand is None:
        raise HTTPException(status_code=404, detail="Stand not found")
    return stand


@app.get("/gates/", response_model=List[schemas.GateResponse])
def list_gates(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_gates(db, skip=skip, limit=limit)


@app.post("/gates/", response_model=schemas.GateResponse, status_code=status.HTTP_201_CREATED)
def create_gate(gate: schemas.GateCreate, db: Session = Depends(get_db)):
    return services.create_gate(db=db, gate=gate)


@app.post("/gates/adjacencies", status_code=status.HTTP_201_CREATED)
def create_adjacency(adjacency: schemas.GateAdjacencyCreate, db: Session = Depends(get_db)):
    return services.create_gate_adjacency(db=db, adjacency=adjacency)


@app.post("/gates/assign", response_model=schemas.GateResponse)
def assign_gate(request: schemas.GateAssignRequest, db: Session = Depends(get_db)):
    gate, error = services.assign_gate_to_flight(db, flight_id=request.flight_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return gate


@app.get("/vehicles/", response_model=List[schemas.VehicleResponse])
def list_vehicles(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_vehicles(db, skip=skip, limit=limit)


@app.get("/vehicles/{vehicle_id}", response_model=schemas.VehicleResponse)
def read_vehicle(vehicle_id: int, db: Session = Depends(get_db)):
    vehicle = services.get_vehicle(db, vehicle_id=vehicle_id)
    if vehicle is None:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    return vehicle


@app.post("/vehicles/", response_model=schemas.VehicleResponse, status_code=status.HTTP_201_CREATED)
def create_vehicle(vehicle: schemas.VehicleCreate, db: Session = Depends(get_db)):
    return services.create_vehicle(db=db, vehicle=vehicle)


@app.put("/vehicles/{vehicle_id}", response_model=schemas.VehicleResponse)
def update_vehicle(vehicle_id: int, vehicle_update: schemas.VehicleUpdate, db: Session = Depends(get_db)):
    vehicle = services.update_vehicle(db, vehicle_id=vehicle_id, vehicle_update=vehicle_update)
    if vehicle is None:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    return vehicle


@app.get("/vehicles/{vehicle_id}/dispatches", response_model=List[schemas.VehicleDispatchResponse])
def list_vehicle_dispatches(vehicle_id: int, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_vehicle_dispatches(db, vehicle_id=vehicle_id, skip=skip, limit=limit)


@app.get("/vehicles/{vehicle_id}/maintenance", response_model=List[schemas.VehicleMaintenanceResponse])
def list_vehicle_maintenances(vehicle_id: int, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_vehicle_maintenances(db, vehicle_id=vehicle_id, skip=skip, limit=limit)


@app.post("/vehicles/{vehicle_id}/maintenance", response_model=schemas.VehicleMaintenanceResponse, status_code=status.HTTP_201_CREATED)
def create_maintenance(vehicle_id: int, maintenance: schemas.VehicleMaintenanceCreate, db: Session = Depends(get_db)):
    result = services.create_vehicle_maintenance(db, vehicle_id=vehicle_id, maintenance=maintenance)
    if result is None:
        raise HTTPException(status_code=404, detail="Vehicle not found")
    return result


@app.post("/dispatches/", response_model=schemas.VehicleDispatchResponse, status_code=status.HTTP_201_CREATED)
def dispatch_vehicle(dispatch: schemas.VehicleDispatchCreate, db: Session = Depends(get_db)):
    result, error = services.dispatch_vehicle_for_flight(db, dispatch_req=dispatch)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return result
