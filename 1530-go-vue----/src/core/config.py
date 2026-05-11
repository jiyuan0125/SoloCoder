from pydantic_settings import BaseSettings
from pathlib import Path
import os


BASE_DIR = Path(__file__).resolve().parent.parent.parent


class Settings(BaseSettings):
    DATABASE_URL: str = f"sqlite:///{BASE_DIR}/app.db"
    API_HOST: str = "127.0.0.1"
    API_PORT: int = 8000
    PRESSURE_DIFF_STDDEV_DAYS: int = 30
    LEAK_THRESHOLD_MULTIPLIER: float = 3.0
    WALL_THICKNESS_THRESHOLD_RATIO: float = 0.8
    ALARM_ESCALATION_MINUTES: int = 60
    ALARM_REMINDER_MINUTES: int = 360

    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()
