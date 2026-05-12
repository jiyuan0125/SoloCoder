from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    PORT: int = 8000
    DATABASE_URL: str = "sqlite:///./logs.db"
    RETENTION_DAYS: int = 30
    ERROR_RETENTION_DAYS: int = 60

    class Config:
        env_file = ".env"
        extra = "ignore"


settings = Settings()
