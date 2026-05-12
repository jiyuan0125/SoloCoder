from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager
from sqlalchemy.orm import Session
from datetime import datetime, timedelta
from .database import init_db, get_db
from .routers import (
    lines,
    stations,
    trains,
    signals,
    power,
    schedules,
    logs,
)
from . import crud, schemas, models


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="地铁运营调度系统",
    description="基于 FastAPI 的地铁运营调度系统，管理列车、信号机、供电区间及运行计划",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(lines.router)
app.include_router(stations.router)
app.include_router(trains.router)
app.include_router(signals.router)
app.include_router(power.router)
app.include_router(schedules.router)
app.include_router(logs.router)


@app.get("/")
def root():
    return {
        "name": "地铁运营调度系统",
        "version": "1.0.0",
        "docs": "/docs",
        "status": "running",
    }


@app.get("/health")
def health_check():
    return {"status": "healthy", "timestamp": datetime.now().isoformat()}


@app.post("/init-sample-data", include_in_schema=False)
def create_sample_data(db: Session = Depends(get_db)):
    existing = crud.get_line_by_name(db, name="1号线")
    if existing:
        return {"message": "Sample data already exists", "line_id": existing.id}

    line = crud.create_line(
        db,
        schemas.LineCreate(name="1号线", description="测试线路"),
    )

    stations_list = [
        crud.create_station(
            db,
            schemas.StationCreate(line_id=line.id, name="起点站", sequence=1, is_terminal=True),
        ),
        crud.create_station(
            db,
            schemas.StationCreate(line_id=line.id, name="中途站1", sequence=2),
        ),
        crud.create_station(
            db,
            schemas.StationCreate(line_id=line.id, name="中途站2", sequence=3),
        ),
        crud.create_station(
            db,
            schemas.StationCreate(line_id=line.id, name="终点站", sequence=4, is_terminal=True),
        ),
    ]

    power_section = crud.create_power_section(
        db,
        schemas.PowerSectionCreate(
            line_id=line.id,
            section_code="P001",
            name="1号线供电区间",
            start_station_id=stations_list[0].id,
            end_station_id=stations_list[-1].id,
        ),
    )

    for i, station in enumerate(stations_list):
        crud.create_signal(
            db,
            schemas.SignalCreate(
                signal_code=f"S{i+1:03d}",
                station_id=station.id,
                section_id=power_section.id,
            ),
        )

    for i in range(5):
        crud.create_train(
            db,
            schemas.TrainCreate(train_number=f"T{i+1:03d}"),
        )

    now = datetime.now().replace(hour=6, minute=0, second=0, microsecond=0)
    end_time = now.replace(hour=23, minute=0)

    schedule = crud.create_schedule(
        db,
        schemas.ScheduleCreate(
            line_id=line.id,
            name="1号线早高峰计划",
            first_departure=now,
            last_departure=end_time,
            interval_seconds=300,
            auto_generate_trains=False,
        ),
    )

    current_time = now
    sequence = 1
    while current_time <= end_time:
        travel_duration = timedelta(minutes=30)
        scheduled_arrival = current_time + travel_duration

        crud.create_schedule_train(
            db,
            schemas.ScheduleTrainCreate(
                schedule_id=schedule.id,
                sequence=sequence,
                scheduled_departure=current_time,
                scheduled_arrival=scheduled_arrival,
            ),
        )
        sequence += 1
        current_time += timedelta(seconds=schedule.interval_seconds)

    for i in range(10):
        timestamp = datetime.now() - timedelta(hours=i)
        crud.create_passenger_data(
            db,
            schemas.PassengerDataCreate(
                line_id=line.id,
                station_id=stations_list[1].id,
                timestamp=timestamp,
                passenger_count=300 + i * 50,
            ),
        )

    return {
        "line_id": line.id,
        "stations": len(stations_list),
        "power_sections": 1,
        "signals": 4,
        "trains": 5,
        "schedules": 1,
    }
