import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PORT: int = int(os.getenv("PORT", "8000"))
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite+aiosqlite:///./water_monitoring.db")
    SECRET_KEY: str = os.getenv("SECRET_KEY", "your-secret-key-change-in-production")
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 24 * 60
    FLOOD_SEASON_MONTHS: list[int] = [6, 7, 8, 9]


settings = Settings()
