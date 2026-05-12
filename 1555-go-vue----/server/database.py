from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from server.models import Base, TimeWindowConfig, WeatherStatus, WeatherCondition
import os

DATABASE_URL = "sqlite:///./pilot_station.db"

engine = create_engine(
    DATABASE_URL, connect_args={"check_same_thread": False}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def init_db():
    Base.metadata.create_all(bind=engine)
    
    db = SessionLocal()
    try:
        _init_default_configs(db)
    finally:
        db.close()


def _init_default_configs(db):
    from sqlalchemy import select
    
    weather_count = db.execute(
        select(WeatherStatus).limit(1)
    ).scalar()
    
    if not weather_count:
        default_weather = WeatherStatus(
            condition=WeatherCondition.GOOD,
            description="天气良好，适合引航作业"
        )
        db.add(default_weather)
    
    config_count = db.execute(
        select(TimeWindowConfig).limit(1)
    ).scalar()
    
    if not config_count:
        night_window = TimeWindowConfig(
            name="夜间引航时段",
            start_time="22:00",
            end_time="06:00",
            description="夜间引航，仅限一级引航员",
            rule_type="night"
        )
        
        peak_window = TimeWindowConfig(
            name="客运高峰期",
            start_time="08:00",
            end_time="10:00",
            description="客运高峰期，禁止货轮进出",
            rule_type="passenger_peak"
        )
        
        db.add_all([night_window, peak_window])
    
    db.commit()
