from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from server.schemas import (
    WeatherUpdate, WeatherResponse, TimeWindowConfigResponse,
    StatisticsResponse
)
from server.services import (
    get_current_weather, update_weather,
    get_time_window_configs, get_statistics
)

router = APIRouter(
    prefix="/system",
    tags=["system"]
)


@router.get("/weather", response_model=WeatherResponse)
def get_weather(db: Session = Depends(get_db)):
    return get_current_weather(db)


@router.put("/weather", response_model=WeatherResponse)
def set_weather(weather_data: WeatherUpdate, db: Session = Depends(get_db)):
    return update_weather(db, weather_data.condition, weather_data.description)


@router.get("/time-windows", response_model=List[TimeWindowConfigResponse])
def list_time_windows(db: Session = Depends(get_db)):
    return get_time_window_configs(db)


@router.get("/statistics", response_model=StatisticsResponse)
def get_stats(db: Session = Depends(get_db)):
    return get_statistics(db)
