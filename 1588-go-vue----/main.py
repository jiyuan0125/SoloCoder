from fastapi import FastAPI, Depends, HTTPException
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime

import models, schemas, services
from database import engine, get_db

models.Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="轮渡公司运营管理系统",
    description="航班、码头、船舶和客流管理系统",
    version="1.0.0"
)


@app.exception_handler(ValueError)
async def value_error_handler(request, exc):
    return JSONResponse(
        status_code=400,
        content={"detail": str(exc)}
    )


@app.get("/")
def root():
    return {"message": "轮渡公司运营管理系统 API", "version": "1.0.0"}


@app.get("/terminals/", response_model=List[schemas.TerminalResponse])
def list_terminals(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_terminals(db, skip=skip, limit=limit)


@app.get("/terminals/{terminal_id}", response_model=schemas.TerminalResponse)
def get_terminal(terminal_id: int, db: Session = Depends(get_db)):
    terminal = services.get_terminal(db, terminal_id)
    if not terminal:
        raise HTTPException(status_code=404, detail="码头不存在")
    return terminal


@app.post("/terminals/", response_model=schemas.TerminalResponse)
def create_terminal(terminal: schemas.TerminalCreate, db: Session = Depends(get_db)):
    return services.create_terminal(db, terminal)


@app.put("/terminals/{terminal_id}", response_model=schemas.TerminalResponse)
def update_terminal(terminal_id: int, terminal: schemas.TerminalUpdate, db: Session = Depends(get_db)):
    updated = services.update_terminal(db, terminal_id, terminal)
    if not updated:
        raise HTTPException(status_code=404, detail="码头不存在")
    return updated


@app.delete("/terminals/{terminal_id}")
def delete_terminal(terminal_id: int, db: Session = Depends(get_db)):
    deleted = services.delete_terminal(db, terminal_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="码头不存在")
    return {"message": "码头已删除"}


@app.get("/berths/", response_model=List[schemas.BerthResponse])
def list_berths(skip: int = 0, limit: int = 100, terminal_id: Optional[int] = None, db: Session = Depends(get_db)):
    return services.get_berths(db, skip=skip, limit=limit, terminal_id=terminal_id)


@app.get("/berths/{berth_id}", response_model=schemas.BerthResponse)
def get_berth(berth_id: int, db: Session = Depends(get_db)):
    berth = services.get_berth(db, berth_id)
    if not berth:
        raise HTTPException(status_code=404, detail="泊位不存在")
    return berth


@app.post("/berths/", response_model=schemas.BerthResponse)
def create_berth(berth: schemas.BerthCreate, db: Session = Depends(get_db)):
    return services.create_berth(db, berth)


@app.put("/berths/{berth_id}", response_model=schemas.BerthResponse)
def update_berth(berth_id: int, berth: schemas.BerthUpdate, db: Session = Depends(get_db)):
    updated = services.update_berth(db, berth_id, berth)
    if not updated:
        raise HTTPException(status_code=404, detail="泊位不存在")
    return updated


@app.delete("/berths/{berth_id}")
def delete_berth(berth_id: int, db: Session = Depends(get_db)):
    deleted = services.delete_berth(db, berth_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="泊位不存在")
    return {"message": "泊位已删除"}


@app.get("/ships/", response_model=List[schemas.ShipResponse])
def list_ships(skip: int = 0, limit: int = 100, status: Optional[models.ShipStatus] = None, db: Session = Depends(get_db)):
    return services.get_ships(db, skip=skip, limit=limit, status=status)


@app.get("/ships/{ship_id}", response_model=schemas.ShipResponse)
def get_ship(ship_id: int, db: Session = Depends(get_db)):
    ship = services.get_ship(db, ship_id)
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    return ship


@app.post("/ships/", response_model=schemas.ShipResponse)
def create_ship(ship: schemas.ShipCreate, db: Session = Depends(get_db)):
    return services.create_ship(db, ship)


@app.put("/ships/{ship_id}", response_model=schemas.ShipResponse)
def update_ship(ship_id: int, ship: schemas.ShipUpdate, db: Session = Depends(get_db)):
    updated = services.update_ship(db, ship_id, ship)
    if not updated:
        raise HTTPException(status_code=404, detail="船舶不存在")
    return updated


@app.delete("/ships/{ship_id}")
def delete_ship(ship_id: int, db: Session = Depends(get_db)):
    deleted = services.delete_ship(db, ship_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="船舶不存在")
    return {"message": "船舶已删除"}


@app.get("/ships/available/", response_model=List[schemas.ShipResponse])
def list_available_ships(scheduled_time: datetime, db: Session = Depends(get_db)):
    return services.get_available_ships(db, scheduled_time)


@app.get("/flights/", response_model=List[schemas.FlightResponse])
def list_flights(
    skip: int = 0, 
    limit: int = 100, 
    status: Optional[models.FlightStatus] = None,
    departure_terminal_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    return services.get_flights(db, skip=skip, limit=limit, status=status, departure_terminal_id=departure_terminal_id)


@app.get("/flights/{flight_id}", response_model=schemas.FlightResponse)
def get_flight(flight_id: int, db: Session = Depends(get_db)):
    flight = services.get_flight(db, flight_id)
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    return flight


@app.get("/flights/number/{flight_number}", response_model=schemas.FlightResponse)
def get_flight_by_number(flight_number: str, db: Session = Depends(get_db)):
    flight = services.get_flight_by_number(db, flight_number)
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    return flight


@app.post("/flights/", response_model=schemas.FlightResponse)
def create_flight(flight: schemas.FlightCreate, db: Session = Depends(get_db)):
    return services.create_flight(db, flight)


@app.get("/flights/{flight_id}/check-departure")
def check_departure(flight_id: int, db: Session = Depends(get_db)):
    flight = services.get_flight(db, flight_id)
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    can_go, message = services.can_depart(db, flight)
    return {"can_depart": can_go, "message": message}


@app.post("/flights/{flight_id}/depart", response_model=schemas.FlightResponse)
def depart_flight(flight_id: int, db: Session = Depends(get_db)):
    return services.update_flight_departure(db, flight_id)


@app.post("/flights/{flight_id}/arrive", response_model=schemas.FlightResponse)
def arrive_flight(flight_id: int, db: Session = Depends(get_db)):
    return services.update_flight_arrival(db, flight_id)


@app.post("/flights/{flight_id}/cancel", response_model=schemas.FlightResponse)
def cancel_flight_endpoint(flight_id: int, reason: str = "航班取消", db: Session = Depends(get_db)):
    return services.cancel_flight(db, flight_id, reason)


@app.post("/flights/{flight_id}/delay", response_model=schemas.FlightResponse)
def delay_flight(flight_id: int, delay_minutes: int, db: Session = Depends(get_db)):
    return services.update_flight_delay(db, flight_id, delay_minutes)


@app.get("/flights/{flight_id}/delay-notification")
def check_delay_notification(flight_id: int, db: Session = Depends(get_db)):
    should_notify = services.check_delay_notification(db, flight_id)
    return {"should_notify": should_notify}


@app.get("/tickets/", response_model=List[schemas.TicketResponse])
def list_tickets(skip: int = 0, limit: int = 100, flight_id: Optional[int] = None, db: Session = Depends(get_db)):
    return services.get_tickets(db, skip=skip, limit=limit, flight_id=flight_id)


@app.post("/tickets/", response_model=schemas.TicketResponse)
def create_ticket(ticket: schemas.TicketCreate, db: Session = Depends(get_db)):
    return services.create_ticket(db, ticket)


@app.get("/weather/", response_model=List[schemas.WeatherResponse])
def list_weather(skip: int = 0, limit: int = 100, terminal_id: Optional[int] = None, db: Session = Depends(get_db)):
    return services.get_weather(db, skip=skip, limit=limit, terminal_id=terminal_id)


@app.post("/weather/", response_model=schemas.WeatherResponse)
def create_weather(weather: schemas.WeatherCreate, db: Session = Depends(get_db)):
    return services.create_weather(db, weather)


@app.get("/refunds/", response_model=List[schemas.RefundResponse])
def list_refunds(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_refunds(db, skip=skip, limit=limit)
