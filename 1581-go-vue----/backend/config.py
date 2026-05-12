from pydantic_settings import BaseSettings
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent


class Settings(BaseSettings):
    DATABASE_URL: str = f"sqlite:///{BASE_DIR / 'metro.db'}"
    MIN_SAFE_INTERVAL: int = 120
    TURNAROUND_TIMEOUT: int = 300

    class Config:
        env_file = ".env"


settings = Settings()
