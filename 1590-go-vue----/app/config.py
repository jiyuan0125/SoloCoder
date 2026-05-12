from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    port: int = 8000
    database_url: str = "sqlite:///./cable_car.db"
    default_luggage_weight: float = 20.0
    weight_warning_threshold: float = 0.95
    weight_max_threshold: float = 1.0
    min_interval_minutes: int = 2
    max_interval_minutes: int = 10
    low_queue_threshold: int = 10
    high_queue_threshold: int = 50
    smoothing_factor: float = 0.3

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


@lru_cache()
def get_settings() -> Settings:
    return Settings()
